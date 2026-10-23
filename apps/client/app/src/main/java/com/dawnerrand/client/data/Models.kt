package com.dawnerrand.client.data

/** 商家 */
data class Merchant(
    val id: String,
    val name: String,
    val recName: String, // 推荐横滑卡片用短名
    val score: String,
    val distance: String,
    val monthlySales: String,
    val deliveryMinutes: String,
    val tags: List<String>,
    val freeTag: String?,
    val bannerColorStart: Int,
    val bannerColorEnd: Int,
    val iconRes: Int,
    val iconColor: Int,
    val freeDeliveryThreshold: Int = 25,
)

/** 菜品 */
data class Dish(
    val id: String,
    val merchantId: String,
    val name: String,
    val category: String,
    val desc: String,
    val price: Double,
    val monthlySales: String,
    val iconRes: Int,
    val iconColor: Int,
    val bgColor: Int,
)

/** 购物车项 */
data class CartItem(val dish: Dish, var count: Int)

/** 订单明细行 */
data class OrderLine(val name: String, val price: Double, val count: Int) {
    val amount: Double get() = price * count
}

/** 订单状态常量 */
object OrderStatus {
    const val PENDING_PAY = "待支付"
    const val PENDING_ACCEPT = "待接单"
    const val DELIVERING = "配送中"
    const val PENDING_REVIEW = "待评价"
    const val DONE = "已完成"
    const val CANCELED = "已取消"
    const val REFUND = "退款售后"
}

/** 订单 */
data class Order(
    val id: String,
    val orderNo: String,
    val shop: String,
    val status: String,
    val items: String,
    val time: String,
    val total: String,
    val iconRes: Int,
    val iconColor: Int,
    val iconBg: Int,
    val type: String = "外卖",
    val address: String = "",
    val remark: String = "",
    val lines: List<OrderLine> = emptyList(),
    val subtotal: Double = 0.0,
    val deliveryFee: Double = 0.0,
    val packingFee: Double = 0.0,
    val discount: Double = 0.0,
    val couponTitle: String = "",
    val couponAmount: Double = 0.0,
    val reviewed: Boolean = false,
    val rating: Int = 0,
    val reviewTags: String = "",
    val reviewText: String = "",
    val createdAt: Long = 0L,
    val merchantId: String? = null,
    val payMethod: String = "微信支付",
) {
    /** 进行中（待支付/待接单/配送中）；已取消、退款售后归入各自 Tab */
    val ongoing: Boolean
        get() = status == OrderStatus.PENDING_PAY ||
            status == OrderStatus.PENDING_ACCEPT ||
            status == OrderStatus.DELIVERING

    /** 已结束（含待评价） */
    val finished: Boolean
        get() = status == OrderStatus.DONE || status == OrderStatus.PENDING_REVIEW
}

/** 地址 */
data class Address(
    val id: String,
    val name: String,
    val phone: String,
    val detail: String,
    val tag: String,
    val isDefault: Boolean = false,
) {
    /** 地址列表只展示两行摘要 */
    val shortDetail: String
        get() = detail.replace("\n", " · ")
}

/** 优惠券 */
data class Coupon(
    val id: String,
    val amount: String,
    val cond: String,
    val title: String,
    val desc: String,
    val expire: String,
    val disabled: Boolean = false,
    val used: Boolean = false,
    val minSpend: Double = 0.0,
    val expiresAt: Long = 0L,
) {
    val value: Double get() = amount.toDoubleOrNull() ?: 0.0
    val usable: Boolean get() = !disabled && !used
}

/** 用户资料 */
data class UserProfile(
    val name: String,
    val phone: String,
    val balance: Double = 15.0,
    val points: Int = 128,
    val level: String = "★5",
)

/** 本地设置 */
data class ClientSettings(
    var orderNotify: Boolean = true,
    var promoNotify: Boolean = false,
    var sound: Boolean = true,
)

/** 入驻申请 */
data class JoinApplication(
    val type: String,
    val status: String,
    val submittedAt: Long,
    val summary: String,
)

/** 客服反馈 */
data class Feedback(
    val content: String,
    val createdAt: Long,
)
