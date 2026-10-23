package com.dawnerrand.client.data

import org.json.JSONArray
import org.json.JSONObject

/**
 * 客户端本地持久化 JSON 映射。
 * 仅使用 Android 内置 org.json，不引入额外序列化依赖。
 */

internal fun Address.toJson(): JSONObject = JSONObject().apply {
    put("id", id)
    put("name", name)
    put("phone", phone)
    put("detail", detail)
    put("tag", tag)
    put("isDefault", isDefault)
}

internal fun addressFromJson(o: JSONObject): Address = Address(
    id = o.str("id"),
    name = o.str("name"),
    phone = o.str("phone"),
    detail = o.str("detail"),
    tag = o.str("tag", "家"),
    isDefault = o.bool("isDefault"),
)

internal fun Coupon.toJson(): JSONObject = JSONObject().apply {
    put("id", id)
    put("amount", amount)
    put("cond", cond)
    put("title", title)
    put("desc", desc)
    put("expire", expire)
    put("disabled", disabled)
    put("used", used)
    put("minSpend", minSpend)
    put("expiresAt", expiresAt)
}

internal fun couponFromJson(o: JSONObject): Coupon = Coupon(
    id = o.str("id"),
    amount = o.str("amount", "0"),
    cond = o.str("cond"),
    title = o.str("title"),
    desc = o.str("desc"),
    expire = o.str("expire"),
    disabled = o.bool("disabled"),
    used = o.bool("used"),
    minSpend = o.dbl("minSpend"),
    expiresAt = o.long("expiresAt"),
)

internal fun OrderLine.toJson(): JSONObject = JSONObject().apply {
    put("name", name)
    put("price", price)
    put("count", count)
}

internal fun orderLineFromJson(o: JSONObject): OrderLine = OrderLine(
    name = o.str("name"),
    price = o.dbl("price"),
    count = o.int("count", 1),
)

internal fun Order.toJson(): JSONObject = JSONObject().apply {
    put("id", id)
    put("orderNo", orderNo)
    put("shop", shop)
    put("status", status)
    put("items", items)
    put("time", time)
    put("total", total)
    put("iconRes", iconRes)
    put("iconColor", iconColor)
    put("iconBg", iconBg)
    put("type", type)
    put("address", address)
    put("remark", remark)
    put("lines", JSONArray().apply { lines.forEach { put(it.toJson()) } })
    put("subtotal", subtotal)
    put("deliveryFee", deliveryFee)
    put("packingFee", packingFee)
    put("discount", discount)
    put("couponTitle", couponTitle)
    put("couponAmount", couponAmount)
    put("reviewed", reviewed)
    put("rating", rating)
    put("reviewTags", reviewTags)
    put("reviewText", reviewText)
    put("createdAt", createdAt)
    put("merchantId", merchantId ?: "")
    put("payMethod", payMethod)
}

internal fun orderFromJson(o: JSONObject): Order = Order(
    id = o.str("id"),
    orderNo = o.str("orderNo"),
    shop = o.str("shop"),
    status = o.str("status", OrderStatus.DONE),
    items = o.str("items"),
    time = o.str("time"),
    total = o.str("total", "¥0.0"),
    iconRes = o.int("iconRes"),
    iconColor = o.int("iconColor"),
    iconBg = o.int("iconBg"),
    type = o.str("type", "外卖"),
    address = o.str("address"),
    remark = o.str("remark"),
    lines = o.array("lines").mapObjects { orderLineFromJson(it) },
    subtotal = o.dbl("subtotal"),
    deliveryFee = o.dbl("deliveryFee"),
    packingFee = o.dbl("packingFee"),
    discount = o.dbl("discount"),
    couponTitle = o.str("couponTitle"),
    couponAmount = o.dbl("couponAmount"),
    reviewed = o.bool("reviewed"),
    rating = o.int("rating"),
    reviewTags = o.str("reviewTags"),
    reviewText = o.str("reviewText"),
    createdAt = o.long("createdAt"),
    merchantId = o.str("merchantId").takeIf { it.isNotBlank() },
    payMethod = o.str("payMethod", "微信支付"),
)

internal fun UserProfile.toJson(): JSONObject = JSONObject().apply {
    put("name", name)
    put("phone", phone)
    put("balance", balance)
    put("points", points)
    put("level", level)
}

internal fun userFromJson(o: JSONObject): UserProfile = UserProfile(
    name = o.str("name", "王建国"),
    phone = o.str("phone"),
    balance = o.dbl("balance", 15.0),
    points = o.int("points", 128),
    level = o.str("level", "★5"),
)

internal fun ClientSettings.toJson(): JSONObject = JSONObject().apply {
    put("orderNotify", orderNotify)
    put("promoNotify", promoNotify)
    put("sound", sound)
}

internal fun settingsFromJson(o: JSONObject): ClientSettings = ClientSettings(
    orderNotify = o.bool("orderNotify", true),
    promoNotify = o.bool("promoNotify"),
    sound = o.bool("sound", true),
)

internal fun JoinApplication.toJson(): JSONObject = JSONObject().apply {
    put("type", type)
    put("status", status)
    put("submittedAt", submittedAt)
    put("summary", summary)
}

internal fun applicationFromJson(o: JSONObject): JoinApplication = JoinApplication(
    type = o.str("type"),
    status = o.str("status", "审核中"),
    submittedAt = o.long("submittedAt"),
    summary = o.str("summary"),
)

internal fun Feedback.toJson(): JSONObject = JSONObject().apply {
    put("content", content)
    put("createdAt", createdAt)
}

internal fun feedbackFromJson(o: JSONObject): Feedback = Feedback(
    content = o.str("content"),
    createdAt = o.long("createdAt"),
)

internal fun CartItem.toJson(): JSONObject = JSONObject().apply {
    put("dishId", dish.id)
    put("count", count)
}

// ---------- JSON 小工具 ----------

internal fun JSONObject.str(key: String, def: String = ""): String =
    if (isNull(key)) def else optString(key, def)

internal fun JSONObject.int(key: String, def: Int = 0): Int = optInt(key, def)

internal fun JSONObject.long(key: String, def: Long = 0L): Long = optLong(key, def)

internal fun JSONObject.dbl(key: String, def: Double = 0.0): Double = optDouble(key, def)

internal fun JSONObject.bool(key: String, def: Boolean = false): Boolean = optBoolean(key, def)

internal fun JSONObject.array(key: String): JSONArray =
    optJSONArray(key) ?: JSONArray()

internal inline fun <T> JSONArray.mapObjects(block: (JSONObject) -> T): List<T> {
    val result = ArrayList<T>(length())
    for (i in 0 until length()) {
        val obj = optJSONObject(i) ?: continue
        result.add(block(obj))
    }
    return result
}

internal fun JSONArray.toStringList(): List<String> {
    val result = ArrayList<String>(length())
    for (i in 0 until length()) result.add(optString(i))
    return result
}
