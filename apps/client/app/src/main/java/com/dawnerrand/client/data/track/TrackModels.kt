package com.dawnerrand.client.data.track

/** 经纬度点（高德 GCJ-02 坐标系）。 */
data class TrackPoint(
    val latitude: Double,
    val longitude: Double,
)

/** 配送追踪阶段（由订单状态映射，驱动地图与状态卡）。 */
enum class TrackPhase {
    NOT_PAID,       // 待支付
    WAITING_ACCEPT, // 等待商家/骑手接单
    DELIVERING,     // 配送中（骑手在途）
    ARRIVED,        // 已送达（待评价 / 已完成）
    CANCELED,       // 已取消
    REFUND,         // 退款售后
}

/** 进度步骤状态。 */
enum class StepState { DONE, ACTIVE, TODO }

/** 进度步骤（横向 4 节点，短文案避免底栏过高）。 */
data class TrackStep(
    val text: String,
    val state: StepState,
)

/**
 * 配送追踪状态（单一数据源，同时驱动地图标记、状态卡、进度步骤与骑手卡）。
 */
data class OrderTrack(
    val orderId: String,
    val phase: TrackPhase,
    val statusTitle: String,
    val statusDesc: String,
    val etaMinutes: String,   // 预计送达（分钟），非配送阶段为 "—"
    val remainingKm: String,  // 剩余距离（公里），非配送阶段为 "—"
    val arrivalClock: String, // 预计送达时刻，如 "12:45"
    val pickup: TrackPoint,   // 取货点（商家）
    val delivery: TrackPoint, // 送达点（客户地址）
    val rider: TrackPoint?,   // 骑手当前位置；无骑手时为 null
    val heading: Float,       // 骑手行进方向（旋转骑手标记）
    val progress: Float,      // 骑手沿路线的进度 0..1（配送中递增）
    val stepProgress: Float,  // 底部横向进度条高亮比例 0..1
    val traveledKm: Double,   // 已行驶公里
    val steps: List<TrackStep>,
    val riderName: String,
    val riderMeta: String,
    val showRiderCard: Boolean,
    val canCallRider: Boolean,
)
