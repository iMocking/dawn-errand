package com.dawnerrand.store.data

import android.content.Context
import com.dawnerrand.store.data.api.ApiClient
import org.json.JSONArray
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/**
 * 商家端真实 API 聚合：登录后把 /api/app/store 的数据写入现有 UI 仓库，
 * 页面仍使用原有 StoreOrder/Dish/FinanceRecord 模型，避免重写视图层。
 */
object StoreRemoteRepository {

    private val timeFmt = SimpleDateFormat("MM-dd HH:mm", Locale.CHINA)

    /** 拉取店铺、订单、菜单、财务；任一接口失败时保留当前数据。 */
    fun refresh(ctx: Context) {
        ApiClient.getObject("/api/app/store/me")?.let { body ->
            body.optJSONObject("shop")?.let { shop ->
                val status = shop.optString("status", "休息中")
                StoreRepository.shop = StoreRepository.shop.copy(
                    name = shop.optString("name", StoreRepository.shop.name),
                    desc = shop.optString("category", StoreRepository.shop.desc),
                    statusText = status,
                    open = status == "营业中",
                    todayRevenue = money(body.optJSONObject("today")?.optDouble("revenue", 0.0) ?: 0.0),
                    todayOrders = (body.optJSONObject("today")?.optInt("orders", 0) ?: 0).toString(),
                    goodRate = "${((shop.optDouble("rating", 4.8) / 5.0) * 100).toInt()}%",
                    monthlySales = StoreRepository.shop.monthlySales,
                    goodScore = "${shop.optDouble("rating", 4.8)}分",
                    reviewCount = StoreRepository.shop.reviewCount,
                )
            }
            body.optJSONObject("business")?.let { biz ->
                StoreSettings.applyRemote(
                    open = biz.optBoolean("open", false),
                    hours = biz.optString("open_hours", ""),
                    autoAccept = biz.optBoolean("auto_accept", false),
                    minPrice = money(biz.optDouble("min_price", 0.0)),
                    deliveryRange = biz.optString("delivery_range", ""),
                    deliveryFee = money(biz.optDouble("delivery_fee", 0.0)),
                    announcement = biz.optString("announcement", ""),
                )
            }
        }
        refreshOrders()
        refreshMenu()
        refreshFinance()
    }

    fun refreshOrders() {
        val body = ApiClient.getObject("/api/app/store/orders?tab=all&pageSize=100") ?: return
        val arr = body.optJSONArray("list") ?: return
        StoreState.ordersLive.postValue(mapOrders(arr))
    }

    fun refreshMenu() {
        val body = ApiClient.getObject("/api/app/store/menu") ?: return
        val arr = body.optJSONArray("categories") ?: return
        val cats = mutableListOf<DishCategory>()
        for (i in 0 until arr.length()) {
            val c = arr.getJSONObject(i)
            val dishes = mutableListOf<Dish>()
            val da = c.optJSONArray("dishes") ?: JSONArray()
            for (j in 0 until da.length()) {
                val d = da.getJSONObject(j)
                val stock = d.optInt("stock", 0)
                dishes += Dish(
                    id = d.optString("id"),
                    name = d.optString("name"),
                    monthlySales = "月售${d.optInt("monthly_sales", 0)}",
                    stockText = if (d.optBoolean("on_sale", true)) {
                        if (stock > 0) "库存${stock}份" else "已售罄"
                    } else "已下架",
                    price = trimPrice(d.optDouble("price", 0.0)),
                    iconRes = com.dawnerrand.store.R.drawable.ic_shop,
                    iconColor = 0xFFFF6B2C.toInt(),
                    bgRes = com.dawnerrand.store.R.drawable.bg_dish_1,
                    onSale = d.optBoolean("on_sale", true),
                    desc = d.optString("description", ""),
                )
            }
            cats += DishCategory(c.optString("name"), dishes)
        }
        DishStore.replaceFromApi(cats)
    }

    fun refreshFinance() {
        val sum = ApiClient.getObject("/api/app/store/finance/summary")
        val body = ApiClient.getObject("/api/app/store/finance/records?page=1&pageSize=50")
        val records = mutableListOf<FinanceRecord>()
        body?.optJSONArray("list")?.let { arr ->
            for (i in 0 until arr.length()) {
                val o = arr.getJSONObject(i)
                val amount = o.optDouble("amount", 0.0)
                val incoming = amount >= 0
                records += FinanceRecord(
                    title = o.optString("title", "账户变动"),
                    time = timeFmt.format(Date(o.optLong("created_at", 0L) * 1000)),
                    amount = (if (incoming) "+" else "-") + money(kotlin.math.abs(amount)),
                    incoming = incoming,
                )
            }
        }
        FinanceStore.replaceFromApi(
            balance = sum?.optDouble("withdrawable", 0.0) ?: FinanceStore.balance.value ?: 0.0,
            records = if (records.isEmpty()) null else records,
        )
    }

    /** 订单动作：接单/拒单/出餐，成功后重新拉取订单。 */
    fun accept(id: String): Boolean =
        ApiClient.postOk("/api/app/store/orders/$id/accept", JSONObject()).also { if (it) refreshOrders() }

    fun reject(id: String, reason: String = "商家繁忙"): Boolean =
        ApiClient.postOk(
            "/api/app/store/orders/$id/reject",
            JSONObject().put("reason", reason),
        ).also { if (it) refreshOrders() }

    fun ready(id: String): Boolean =
        ApiClient.postOk("/api/app/store/orders/$id/ready", JSONObject()).also { if (it) refreshOrders() }

    fun setOpen(open: Boolean): Boolean =
        ApiClient.postOk("/api/app/store/status", JSONObject().put("open", open)).also {
            if (it) {
                StoreRepository.shop = StoreRepository.shop.copy(
                    open = open, statusText = if (open) "营业中" else "休息中",
                )
            }
        }

    private fun mapOrders(arr: JSONArray): List<StoreOrder> {
        val list = mutableListOf<StoreOrder>()
        for (i in 0 until arr.length()) {
            val o = arr.getJSONObject(i)
            val status = when (o.optString("status_code")) {
                "pending_accept", "待接单" -> ORDER_PENDING
                "preparing", "制作中" -> ORDER_COOKING
                "ready", "delivering", "待取餐", "配送中" -> ORDER_DELIVERING
                "delivered", "completed", "pending_review", "已完成", "已送达", "待评价" -> ORDER_DONE
                else -> ORDER_DONE
            }
            val actions = mutableListOf<String>()
            o.optJSONArray("actions")?.let { aa ->
                for (j in 0 until aa.length()) {
                    actions += when (aa.optString(j)) {
                        "accept" -> "接单"
                        "reject" -> "拒单"
                        "ready" -> "通知出餐"
                        "detail" -> "查看配送"
                        else -> "查看详情"
                    }
                }
            }
            if (actions.isEmpty() && status == ORDER_DONE) actions += "再来一单"
            val customer = o.optJSONObject("customer")
            list += StoreOrder(
                id = o.optString("id"),
                status = status,
                customerName = customer?.optString("name").orEmpty(),
                customerPhone = customer?.optString("phone").orEmpty(),
                address = o.optString("address"),
                items = o.optString("items_summary"),
                total = money(o.optDouble("pay_amount", 0.0)),
                actions = actions,
            )
        }
        return list
    }

    private fun money(v: Double): String {
        val s = String.format(Locale.CHINA, "%,.2f", v)
        return "¥$s"
    }

    private fun trimPrice(v: Double): String =
        if (v % 1.0 == 0.0) v.toInt().toString() else String.format(Locale.CHINA, "%.1f", v)
}
