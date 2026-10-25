package com.dawnerrand.client.data

import android.graphics.Color
import com.dawnerrand.client.R
import org.json.JSONArray
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/**
 * 用户端真实 API 聚合：/api/app/client 的商家、订单、资料、地址。
 * 登录后刷新，Home/Orders/Profile 继续复用现有模型。
 */
object ClientRemoteRepository {

    private val timeFmt = SimpleDateFormat("MM-dd HH:mm", Locale.CHINA)

    suspend fun refresh() {
        refreshMerchants()
        refreshMe()
        refreshOrders()
        refreshAddresses()
    }

    private suspend fun refreshMerchants() {
        val body = ApiClient.getObject("/api/app/client/merchants?page=1&pageSize=50") ?: return
        val arr = body.optJSONArray("list") ?: return
        val list = mutableListOf<Merchant>()
        for (i in 0 until arr.length()) {
            val o = arr.getJSONObject(i)
            val fee = o.optDouble("delivery_fee", 0.0)
            val minutes = o.optInt("delivery_minutes", 30)
            val rating = o.optDouble("rating", 4.8)
            val monthly = o.optInt("monthly_sales", 0)
            list += Merchant(
                id = o.optString("id"),
                name = o.optString("name"),
                recName = o.optString("name").take(8),
                score = String.format(Locale.CHINA, "%.1f", rating),
                distance = if (o.optDouble("distance_km", 0.0) > 0.0) {
                    String.format(Locale.CHINA, "%.1fkm", o.optDouble("distance_km", 0.0))
                } else "附近",
                monthlySales = monthly.toString(),
                deliveryMinutes = (if (minutes > 0) minutes else 30).toString(),
                tags = listOfNotNull(o.optString("category").takeIf { it.isNotBlank() }),
                freeTag = if (fee <= 0.0) "免配送费" else null,
                bannerColorStart = 0xFFFF8A3D.toInt(),
                bannerColorEnd = 0xFFFF6B2C.toInt(),
                iconRes = R.drawable.ic_food_bag,
                iconColor = Color.rgb(0xFF, 0x6B, 0x2C),
            )
        }
        DemoRepository.merchants = list
    }

    private suspend fun refreshMe() {
        val me = ApiClient.getObject("/api/app/client/me") ?: return
        val name = me.optJSONObject("user")?.optString("name").orEmpty()
            .ifBlank { me.optString("name") }
        val phone = me.optString("phone")
        if (name.isNotBlank()) {
            DemoRepository.currentUser = DemoRepository.currentUser.copy(
                name = name,
                phone = phone.ifBlank { DemoRepository.currentUser.phone },
            )
        }
    }

    private suspend fun refreshOrders() {
        val body = ApiClient.getObject("/api/app/client/orders?page=1&pageSize=50") ?: return
        val arr = body.optJSONArray("list") ?: return
        val list = mutableListOf<Order>()
        for (i in 0 until arr.length()) {
            val o = arr.getJSONObject(i)
            list += Order(
                id = o.optString("id"),
                orderNo = o.optString("order_no"),
                shop = o.optString("shop_name"),
                status = o.optString("status"),
                items = o.optString("items_summary"),
                time = timeFmt.format(Date(o.optLong("created_at", 0L) * 1000)),
                total = money(o.optDouble("pay_amount", 0.0)),
                iconRes = R.drawable.ic_food_bag,
                iconColor = Color.rgb(0xFF, 0x6B, 0x2C),
                iconBg = Color.rgb(0xFF, 0xF3, 0xE8),
                type = if (o.optString("type") == "errand") "跑腿" else "外卖",
                createdAt = o.optLong("created_at", 0L),
            )
        }
        DemoRepository.orders = list
    }

    private suspend fun refreshAddresses() {
        val body = ApiClient.getObject("/api/app/client/addresses") ?: return
        val arr = body.optJSONArray("list") ?: return
        val list = mutableListOf<Address>()
        for (i in 0 until arr.length()) {
            val a = arr.getJSONObject(i)
            list += Address(
                id = a.optString("id"),
                name = a.optString("contact_name"),
                phone = a.optString("phone"),
                detail = a.optString("detail"),
                tag = a.optString("tag"),
                isDefault = a.optBoolean("is_default", false),
            )
        }
        DemoRepository.addresses = list.toMutableList()
    }

    private fun money(v: Double): String =
        "¥" + String.format(Locale.CHINA, "%,.2f", v)
}
