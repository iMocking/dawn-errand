package com.dawnerrand.driver.data.repository

import com.dawnerrand.driver.data.api.ApiClient
import com.dawnerrand.driver.data.model.ActiveOrder
import com.dawnerrand.driver.data.model.FinanceRecord
import com.dawnerrand.driver.data.model.OrderItem
import com.dawnerrand.driver.data.model.OrderRecord
import com.dawnerrand.driver.data.model.RiderProfile
import org.json.JSONArray
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/**
 * 骑手端真实 API 聚合：/api/app/driver 的 me、order-pool、orders、finance。
 * 登录后刷新，页面仍复用 DriverRepository 的展示模型。
 */
object DriverRemoteRepository {

    private val timeFmt = SimpleDateFormat("MM-dd HH:mm", Locale.CHINA)

    fun refresh() {
        ApiClient.getObject("/api/app/driver/me")?.let { me ->
            val today = me.optJSONObject("today")
            DriverRepository.rider = DriverRepository.rider.copy(
                name = me.optString("name", DriverRepository.rider.name),
                level = me.optString("level", DriverRepository.rider.level),
                area = me.optString("area", DriverRepository.rider.area),
                score = me.optDouble("rating", 4.9).toString(),
                todayAmount = "¥" + trim(today?.optDouble("gmv", 0.0) ?: 0.0),
                todayOrders = today?.optInt("orders", 0) ?: 0,
                monthIncome = "¥" + trim(me.optDouble("month_income", 0.0)),
            )
        }
        refreshOrders()
        refreshPool()
        refreshFinance()
    }

    fun refreshOrders() {
        val body = ApiClient.getObject("/api/app/driver/orders?page=1&pageSize=50") ?: return
        val arr = body.optJSONArray("list") ?: return
        DriverRepository.orderHistory = mapHistory(arr)
        DriverRepository.activeOrder = mapActive(arr)
    }

    fun refreshPool() {
        val body = ApiClient.getObject("/api/app/driver/order-pool?page=1&pageSize=50") ?: return
        val arr = body.optJSONArray("list") ?: return
        DriverRepository.orderPool = mapPool(arr)
    }

    fun refreshFinance() {
        val sum = ApiClient.getObject("/api/app/driver/finance/summary")
        val body = ApiClient.getObject("/api/app/driver/finance/records?page=1&pageSize=50")
        val records = mutableListOf<FinanceRecord>()
        body?.optJSONArray("list")?.let { arr ->
            for (i in 0 until arr.length()) {
                val o = arr.getJSONObject(i)
                val amount = o.optDouble("amount", 0.0)
                records += FinanceRecord(
                    title = o.optString("title", "账户变动"),
                    time = timeFmt.format(Date(o.optLong("created_at", 0L) * 1000)),
                    amount = (if (amount >= 0) "+" else "-") + "¥" + trim(kotlin.math.abs(amount)),
                    incoming = amount >= 0,
                )
            }
        }
        if (records.isNotEmpty()) DriverRepository.financeRecords = records
        sum?.let {
            val balance = it.optDouble("balance", 0.0)
            val todayIncome = it.optDouble("today_income", 0.0)
            DriverRepository.rider = DriverRepository.rider.copy(
                balance = trim(balance),
                todayAmount = "¥" + trim(todayIncome),
            )
        }
    }

    /** 上线/下线：PUT /api/app/driver/status；返回完整结果，失败原因由调用方决定怎么提示。 */
    fun setOnline(online: Boolean): ApiClient.Resp =
        ApiClient.putResp("/api/app/driver/status", JSONObject().put("online", online))

    /** 订单详情：GET /api/app/driver/orders/:id。返回 null 表示网络/权限/状态失败（原因已写日志）。 */
    fun orderDetail(orderId: String): JSONObject? =
        ApiClient.getObject("/api/app/driver/orders/$orderId")

    /**
     * 抢单：POST /api/app/driver/orders/:id/grab。
     * 返回完整结果而不是 Boolean——1801「已被抢」必须让骑手看到，不能只看日志。
     */
    fun grab(orderId: String): ApiClient.Resp =
        ApiClient.postResp("/api/app/driver/orders/$orderId/grab", JSONObject()).also { if (it.ok) refresh() }

    /** 取货：后端校验取货码（订单未生成取货码时为空串放行）。 */
    fun pickup(orderId: String, code: String): ApiClient.Resp =
        ApiClient.postResp("/api/app/driver/orders/$orderId/pickup", JSONObject().put("pickup_code", code))
            .also { if (it.ok) refresh() }

    /** 送达：后端校验送达码，成功后返回的 body 里带 rider_earning（实际入账金额）。 */
    fun deliver(orderId: String, code: String): ApiClient.Resp =
        ApiClient.postResp("/api/app/driver/orders/$orderId/deliver", JSONObject().put("delivery_code", code))
            .also { if (it.ok) refresh() }

    private fun mapActive(arr: JSONArray): ActiveOrder {
        for (i in 0 until arr.length()) {
            val o = arr.getJSONObject(i)
            val status = o.optString("status")
            // 后端状态用「待取货/配送中」，此前这里写成设计稿的「待取餐」，导致待取货的单永远进不了工作台卡片。
            if (status == "待取货" || status == "配送中") {
                val pickup = o.optJSONObject("pickup")
                val dropoff = o.optJSONObject("dropoff")
                return ActiveOrder(
                    pickName = pickup?.optString("name").orEmpty(),
                    pickAddr = pickup?.optString("address").orEmpty(),
                    pickContact = "",
                    dropName = dropoff?.optString("name").orEmpty(),
                    dropAddr = dropoff?.optString("address").orEmpty(),
                    dropContact = "",
                    status = status,
                    id = o.optString("id"),
                )
            }
        }
        return ActiveOrder("", "暂无进行中订单", "", "", "上线后可在接单大厅抢单", "", "待接单")
    }

    private fun mapPool(arr: JSONArray): List<OrderItem> {
        val list = mutableListOf<OrderItem>()
        for (i in 0 until arr.length()) {
            val o = arr.getJSONObject(i)
            val pickup = o.optJSONObject("pickup")
            val dropoff = o.optJSONObject("dropoff")
            list += OrderItem(
                type = if (o.optString("type") == "errand") "帮我送" else "外卖订单",
                isGreenTag = o.optDouble("distance_km", 0.0) < 2.0,
                distance = trim(o.optDouble("distance_km", 0.0)) + "km",
                pickup = pickup?.optString("name").orEmpty(),
                dropoff = dropoff?.optString("address").orEmpty(),
                reward = "¥" + trim(o.optDouble("reward", 0.0)),
                rewardLabel = o.optString("reward_label", "配送费"),
                distanceKm = o.optDouble("distance_km", 0.0).toFloat(),
                id = o.optString("order_id"),
            )
        }
        return list
    }

    private fun mapHistory(arr: JSONArray): List<OrderRecord> {
        val list = mutableListOf<OrderRecord>()
        for (i in 0 until arr.length()) {
            val o = arr.getJSONObject(i)
            val pickup = o.optJSONObject("pickup")
            val dropoff = o.optJSONObject("dropoff")
            list += OrderRecord(
                time = timeFmt.format(Date(o.optLong("created_at", 0L) * 1000)),
                status = o.optString("status"),
                routeFrom = pickup?.optString("name").orEmpty(),
                routeTo = dropoff?.optString("address").orEmpty(),
                typeInfo = o.optString("type", "外卖") + " · " + trim(o.optDouble("distance_km", 0.0)) + "km",
                earn = "+¥" + trim(o.optDouble("rider_earning", 0.0)),
                id = o.optString("id"),
            )
        }
        return list
    }

    private fun trim(v: Double): String =
        if (v % 1.0 == 0.0) v.toInt().toString() else String.format(Locale.CHINA, "%.1f", v)
}
