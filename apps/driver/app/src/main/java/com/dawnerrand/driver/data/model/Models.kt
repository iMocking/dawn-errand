package com.dawnerrand.driver.data.model

/** 骑手画像（还原 profile.html / home.html 数据） */
data class RiderProfile(
    val name: String = "张师傅",
    val level: String = "银牌骑手",
    val area: String = "绿杨乡片区",
    val serviceDays: Int = 286,
    val score: String = "4.9",
    val onTimeRate: String = "99%",
    val totalOrders: Int = 568,
    val totalIncome: String = "¥12.6k",
    val balance: String = "1,286.50",
    val todayIncome: String = "¥186.5",
    val monthIncome: String = "¥4,520",
    val monthOrders: String = "156单",
    val todayAmount: String = "¥186",
    val todayOrders: Int = 12,
)

/** 接单大厅订单（还原 order-pool.html） */
data class OrderItem(
    val type: String,
    val isGreenTag: Boolean,
    val distance: String,
    val pickup: String,
    val dropoff: String,
    val reward: String,
    val rewardLabel: String,
    val distanceKm: Float,
    /** 后端订单 ID：抢单与详情都必须带它，否则只能演给用户看。 */
    val id: String = "",
)

/** 历史订单（还原 orders.html） */
data class OrderRecord(
    val time: String,
    val status: String,
    val routeFrom: String,
    val routeTo: String,
    val typeInfo: String,
    val earn: String,
    val id: String = "",
)

/** 收入流水（还原 finance.html） */
data class FinanceRecord(
    val title: String,
    val time: String,
    val amount: String,
    val incoming: Boolean,
)

/** 进行中订单（还原 home.html active-card / order-detail.html） */
data class ActiveOrder(
    val pickName: String,
    val pickAddr: String,
    val pickContact: String,
    val dropName: String,
    val dropAddr: String,
    val dropContact: String,
    val status: String,
    /** 后端订单 ID：详情页据此拉真实数据并推进状态。 */
    val id: String = "",
)
