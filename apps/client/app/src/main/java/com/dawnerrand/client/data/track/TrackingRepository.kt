package com.dawnerrand.client.data.track

import com.dawnerrand.client.data.Order
import com.dawnerrand.client.data.OrderStatus
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import kotlin.math.atan2
import kotlin.math.asin
import kotlin.math.ceil
import kotlin.math.cos
import kotlin.math.pow
import kotlin.math.roundToInt
import kotlin.math.sin
import kotlin.math.sqrt

/**
 * 骑手配送追踪数据源。
 *
 * - 当前为**演示实现**：以浠水县镇中心→王家湾村的乡道路线模拟骑手匀速移动，
 *   数据经 [stateFor] 纯函数生成，与视图完全解耦；
 * - 后端提供追踪接口后，把 [stateFor] 替换为对 `GET /orders/:id/track` 的轮询
 *   （或 WebSocket 推送）解析即可，ViewModel 与 UI 零改动。
 */
object TrackingRepository {

    // ---------------- 演示路线：浠水县镇中心（取货） → 王家湾村（送达） ----------------

    private val villageRoute = listOf(
        TrackPoint(30.4522, 115.2620), // 0 取货点（镇中心商家）
        TrackPoint(30.4480, 115.2685),
        TrackPoint(30.4415, 115.2730),
        TrackPoint(30.4340, 115.2795),
        TrackPoint(30.4275, 115.2860),
        TrackPoint(30.4210, 115.2915),
        TrackPoint(30.4200, 115.2950), // 6 送达点（王家湾村）
    )

    val pickup: TrackPoint get() = villageRoute.first()
    val delivery: TrackPoint get() = villageRoute.last()

    /** 完整路线点（取货→送达），供地图绘制 Polyline。 */
    val routePoints: List<TrackPoint> get() = villageRoute

    /** 按里程比例取路线点（progress 0..1 从取货点到送达点）。 */
    fun pointAt(progress: Float): TrackPoint {
        val p = progress.coerceIn(0f, 1f).toDouble()
        val total = routeLength()
        val target = p * total
        var acc = 0.0
        for (i in 0 until villageRoute.size - 1) {
            val seg = distanceMeters(villageRoute[i], villageRoute[i + 1])
            if (acc + seg >= target) {
                val t = if (seg <= 1e-9) 0.0 else (target - acc) / seg
                return lerp(villageRoute[i], villageRoute[i + 1], t)
            }
            acc += seg
        }
        return villageRoute.last()
    }

    private fun routeLength(): Double {
        var total = 0.0
        for (i in 0 until villageRoute.size - 1) {
            total += distanceMeters(villageRoute[i], villageRoute[i + 1])
        }
        return total
    }

    private fun lerp(a: TrackPoint, b: TrackPoint, t: Double): TrackPoint =
        TrackPoint(
            latitude = a.latitude + (b.latitude - a.latitude) * t,
            longitude = a.longitude + (b.longitude - a.longitude) * t,
        )

    // ---------------- 状态生成（纯函数） ----------------

    /**
     * 根据订单与骑手进度生成完整追踪状态。
     *
     * @param progress 骑手沿路线进度 0..1；非配送阶段传 0 或 1 由内部映射。
     */
    fun stateFor(order: Order, progress: Float): OrderTrack {
        val phase = when (order.status) {
            OrderStatus.PENDING_PAY -> TrackPhase.NOT_PAID
            OrderStatus.PENDING_ACCEPT -> TrackPhase.WAITING_ACCEPT
            OrderStatus.DELIVERING -> TrackPhase.DELIVERING
            OrderStatus.PENDING_REVIEW, OrderStatus.DONE -> TrackPhase.ARRIVED
            OrderStatus.CANCELED -> TrackPhase.CANCELED
            OrderStatus.REFUND -> TrackPhase.REFUND
            else -> TrackPhase.WAITING_ACCEPT
        }
        val riderName = if (order.type == "跑腿") "王师傅 · 鄂J·6K328" else "张师傅 · 鄂J·D8888"
        val riderMeta = "接单${if (order.type == "跑腿") "986" else "1286"} · 好评率99% · 曙光骑手"
        val targetName = shortTarget(order)

        return when (phase) {
            TrackPhase.NOT_PAID -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "订单待支付",
                statusDesc = "完成支付后骑手将尽快接单",
                etaMinutes = "—", remainingKm = "—",
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, traveledKm = 0.0,
                steps = pendingSteps(),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )

            TrackPhase.WAITING_ACCEPT -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "等待骑手接单",
                statusDesc = "预计 ${etaClock(order)} 送达 · 正在分配骑手",
                etaMinutes = "—", remainingKm = "—",
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, traveledKm = 0.0,
                steps = listOf(
                    TrackStep("订单已提交，商家已接单", StepState.DONE),
                    TrackStep("骑手已到店取餐", StepState.TODO),
                    TrackStep("配送中", StepState.TODO),
                    TrackStep("送达确认，期待您的评价", StepState.TODO),
                ),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )

            TrackPhase.DELIVERING -> {
                val rider = pointAt(progress)
                val remaining = distanceKm(rider, delivery)
                val minutes = ceil(remaining / AVG_KM_PER_MIN).roundToInt().coerceAtLeast(1)
                val traveled = totalKm * progress
                OrderTrack(
                    orderId = order.id, phase = phase,
                    statusTitle = "骑手正在配送中",
                    statusDesc = "骑手${riderName.substringBefore(" · ")}已取餐，正赶往$targetName",
                    etaMinutes = minutes.toString(), remainingKm = remaining.fmtKm(),
                    pickup = pickup, delivery = delivery,
                    rider = rider, heading = bearing(pointAt(progress - 0.02f), rider),
                    progress = progress, traveledKm = traveled,
                    steps = listOf(
                        TrackStep("订单已提交，商家已接单", StepState.DONE),
                        TrackStep("骑手已到店取餐", StepState.DONE),
                        TrackStep("配送中，已行驶${traveled.fmtKm()}公里", StepState.ACTIVE),
                        TrackStep("送达确认，期待您的评价", StepState.TODO),
                    ),
                    riderName = riderName, riderMeta = riderMeta,
                    showRiderCard = true, canCallRider = true,
                )
            }

            TrackPhase.ARRIVED -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "订单已送达",
                statusDesc = "已于 ${etaClock(order)} 送达，感谢使用曙光跑腿",
                etaMinutes = "0", remainingKm = "0.0",
                pickup = pickup, delivery = delivery,
                rider = delivery, heading = 0f, progress = 1f, traveledKm = totalKm,
                steps = listOf(
                    TrackStep("订单已提交，商家已接单", StepState.DONE),
                    TrackStep("骑手已到店取餐", StepState.DONE),
                    TrackStep("配送中，已行驶${totalKm.fmtKm()}公里", StepState.DONE),
                    TrackStep("送达确认，期待您的评价", StepState.DONE),
                ),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = true, canCallRider = true,
            )

            TrackPhase.CANCELED -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "订单已取消",
                statusDesc = "订单已取消，如已支付将原路退回",
                etaMinutes = "—", remainingKm = "—",
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, traveledKm = 0.0,
                steps = pendingSteps(),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )

            TrackPhase.REFUND -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "售后处理中",
                statusDesc = "客服正在处理售后，请耐心等待",
                etaMinutes = "—", remainingKm = "—",
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, traveledKm = 0.0,
                steps = pendingSteps(),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )
        }
    }

    private fun pendingSteps() = listOf(
        TrackStep("订单已提交，商家已接单", StepState.DONE),
        TrackStep("骑手已到店取餐", StepState.TODO),
        TrackStep("配送中", StepState.TODO),
        TrackStep("送达确认，期待您的评价", StepState.TODO),
    )

    // ---------------- 工具 ----------------

    private const val EARTH_RADIUS_M = 6_371_000.0
    private const val AVG_KM_PER_MIN = 0.33 // 骑手平均 20km/h

    /** 起点→终点航向角（0 北、90 东，顺时针）。 */
    fun bearing(from: TrackPoint, to: TrackPoint): Float {
        val lat1 = Math.toRadians(from.latitude)
        val lat2 = Math.toRadians(to.latitude)
        val dLng = Math.toRadians(to.longitude - from.longitude)
        val y = sin(dLng) * cos(lat2)
        val x = cos(lat1) * sin(lat2) - sin(lat1) * cos(lat2) * cos(dLng)
        return ((Math.toDegrees(atan2(y, x)) + 360.0) % 360.0).toFloat()
    }

    /** Haversine 距离（米）。 */
    fun distanceMeters(a: TrackPoint, b: TrackPoint): Double {
        val dLat = Math.toRadians(b.latitude - a.latitude)
        val dLng = Math.toRadians(b.longitude - a.longitude)
        val s = sin(dLat / 2).pow(2) +
            cos(Math.toRadians(a.latitude)) * cos(Math.toRadians(b.latitude)) * sin(dLng / 2).pow(2)
        return 2 * EARTH_RADIUS_M * asin(sqrt(s))
    }

    private fun distanceKm(a: TrackPoint, b: TrackPoint): Double = distanceMeters(a, b) / 1000.0

    private fun Double.fmtKm(): String = String.format(Locale.CHINA, "%.1f", this)

    private val totalKm: Double get() = routeLength() / 1000.0

    private fun etaClock(order: Order): String =
        SimpleDateFormat("HH:mm", Locale.CHINA).format(Date(order.createdAt + 35 * 60_000L))

    /** 从订单地址提取送达地短名（设计稿“王家湾村”样式）。 */
    private fun shortTarget(order: Order): String {
        val raw = order.address
        val part = raw.substringAfter(" · ").substringBefore("\n").trim()
        return part.ifBlank { "王家湾村" }
    }
}
