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
 * - 路线经过 Chaikin 平滑，定位点按里程插值，避免地图上出现折线拐角；
 * - 后端提供追踪接口后，把 [stateFor] 替换为对 `GET /orders/:id/track` 的轮询
 *   （或 WebSocket 推送）解析即可，ViewModel 与 UI 零改动。
 */
object TrackingRepository {

    // ---------------- 演示路线：浠水县镇中心（取货） → 王家湾村（送达） ----------------

    private val villageRoute = listOf(
        TrackPoint(30.4522, 115.2620), // 0 取货点（镇中心商家）
        TrackPoint(30.4530, 115.2660),
        TrackPoint(30.4528, 115.2700),
        TrackPoint(30.4522, 115.2740),
        TrackPoint(30.4516, 115.2780),
        TrackPoint(30.4510, 115.2820), // 5 送达点（王家湾村，河北岸道路）
    )

    /** 平滑后的路线点（取货→送达），供地图绘制与插值；可由高德路径规划替换。 */
    @Volatile
    private var activeRoute: List<TrackPoint> = smooth(villageRoute, iterations = 3)

    val routePoints: List<TrackPoint> get() = activeRoute

    @Volatile
    private var segments: List<Double> = buildSegments(activeRoute)

    @Volatile
    private var totalMeters: Double = segments.sum()

    /** 高德驾车路径规划返回后替换演示路线；失败时保留手工乡道路线。 */
    fun setRoute(points: List<TrackPoint>) {
        if (points.size < 2) return
        val cleaned = ArrayList<TrackPoint>(points.size)
        points.forEach { point ->
            if (cleaned.isEmpty() ||
                cleaned.last().latitude != point.latitude ||
                cleaned.last().longitude != point.longitude
            ) {
                cleaned.add(point)
            }
        }
        if (cleaned.size < 2) return
        val smoothed = smooth(cleaned, iterations = 1)
        activeRoute = smoothed
        segments = buildSegments(smoothed)
        totalMeters = segments.sum()
    }

    val pickup: TrackPoint get() = routePoints.first()
    val delivery: TrackPoint get() = routePoints.last()

    /** 按里程比例取路线点（progress 0..1 从取货点到送达点）。 */
    fun pointAt(progress: Float): TrackPoint {
        val target = progress.coerceIn(0f, 1f) * totalMeters
        var acc = 0.0
        for (i in 0 until routePoints.size - 1) {
            val seg = segments[i]
            if (acc + seg >= target) {
                val t = if (seg <= 1e-9) 0.0 else (target - acc) / seg
                return lerp(routePoints[i], routePoints[i + 1], t)
            }
            acc += seg
        }
        return routePoints.last()
    }

    /** 取货点 → 当前进度点的导航折线（已走过的路线） */
    fun pointsUntil(progress: Float): List<TrackPoint> {
        val target = progress.coerceIn(0f, 1f) * totalMeters
        val out = ArrayList<TrackPoint>()
        out.add(routePoints.first())
        var acc = 0.0
        for (i in 0 until routePoints.size - 1) {
            val seg = segments[i]
            if (target >= acc + seg) {
                out.add(routePoints[i + 1])
                acc += seg
            } else {
                val t = if (seg <= 1e-9) 0.0 else (target - acc) / seg
                out.add(lerp(routePoints[i], routePoints[i + 1], t))
                break
            }
        }
        return out
    }

    /** 当前进度点 → 送达点的剩余路线 */
    fun pointsAfter(progress: Float): List<TrackPoint> {
        val target = progress.coerceIn(0f, 1f) * totalMeters
        val out = ArrayList<TrackPoint>()
        var acc = 0.0
        for (i in 0 until routePoints.size - 1) {
            val seg = segments[i]
            if (target >= acc + seg) {
                acc += seg
                continue
            }
            val t = if (seg <= 1e-9) 0.0 else (target - acc) / seg
            out.add(lerp(routePoints[i], routePoints[i + 1], t))
            for (j in i + 1 until routePoints.size) out.add(routePoints[j])
            break
        }
        if (out.isEmpty()) out.add(routePoints.last())
        return out
    }

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
        val baseClock = etaClock(order)

        return when (phase) {
            TrackPhase.NOT_PAID -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "订单待支付",
                statusDesc = "完成支付后骑手将尽快接单",
                etaMinutes = "—", remainingKm = "—", arrivalClock = baseClock,
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, stepProgress = 0.12f,
                traveledKm = 0.0,
                steps = listOf(
                    TrackStep("已下单", StepState.ACTIVE),
                    TrackStep("待支付", StepState.TODO),
                    TrackStep("配送中", StepState.TODO),
                    TrackStep("已送达", StepState.TODO),
                ),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )

            TrackPhase.WAITING_ACCEPT -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "等待骑手接单",
                statusDesc = "正在分配骑手 · 预计 $baseClock 送达",
                etaMinutes = "—", remainingKm = "—", arrivalClock = baseClock,
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, stepProgress = 0.34f,
                traveledKm = 0.0,
                steps = listOf(
                    TrackStep("已提交", StepState.DONE),
                    TrackStep("待接单", StepState.ACTIVE),
                    TrackStep("配送中", StepState.TODO),
                    TrackStep("已送达", StepState.TODO),
                ),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )

            TrackPhase.DELIVERING -> {
                val p = progress.coerceIn(0f, 1f)
                val rider = pointAt(p)
                val remaining = distanceKm(rider, delivery)
                val minutes = ceil(remaining / AVG_KM_PER_MIN).roundToInt().coerceAtLeast(0)
                val arrived = p >= 0.999f
                val traveler = p - 0.02f
                OrderTrack(
                    orderId = order.id, phase = phase,
                    statusTitle = if (arrived) "骑手即将送达" else "骑手正在配送中",
                    statusDesc = if (arrived) {
                        "骑手已到达${targetName}，请准备收货"
                    } else {
                        "骑手${riderName.substringBefore(" · ")}已取餐，正赶往$targetName"
                    },
                    etaMinutes = minutes.toString(),
                    remainingKm = remaining.fmtKm(),
                    arrivalClock = if (arrived) clockAfter(0) else clockAfter(minutes),
                    pickup = pickup, delivery = delivery,
                    rider = rider,
                    heading = bearing(pointAt(traveler), rider),
                    progress = p,
                    stepProgress = 0.66f + p * 0.34f,
                    traveledKm = totalKm * p,
                    steps = listOf(
                        TrackStep("已接单", StepState.DONE),
                        TrackStep("已取餐", StepState.DONE),
                        TrackStep("配送中", if (arrived) StepState.DONE else StepState.ACTIVE),
                        TrackStep("已送达", if (arrived) StepState.ACTIVE else StepState.TODO),
                    ),
                    riderName = riderName, riderMeta = riderMeta,
                    showRiderCard = true, canCallRider = true,
                )
            }

            TrackPhase.ARRIVED -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "订单已送达",
                statusDesc = "已于 $baseClock 送达，感谢使用曙光跑腿",
                etaMinutes = "0", remainingKm = "0.0", arrivalClock = baseClock,
                pickup = pickup, delivery = delivery,
                rider = delivery, heading = 0f, progress = 1f, stepProgress = 1f,
                traveledKm = totalKm,
                steps = listOf(
                    TrackStep("已接单", StepState.DONE),
                    TrackStep("已取餐", StepState.DONE),
                    TrackStep("配送中", StepState.DONE),
                    TrackStep("已送达", StepState.DONE),
                ),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = true, canCallRider = true,
            )

            TrackPhase.CANCELED -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "订单已取消",
                statusDesc = "订单已取消，如已支付将原路退回",
                etaMinutes = "—", remainingKm = "—", arrivalClock = baseClock,
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, stepProgress = 0.25f,
                traveledKm = 0.0,
                steps = listOf(
                    TrackStep("已提交", StepState.DONE),
                    TrackStep("已取消", StepState.DONE),
                    TrackStep("配送中", StepState.TODO),
                    TrackStep("已送达", StepState.TODO),
                ),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )

            TrackPhase.REFUND -> OrderTrack(
                orderId = order.id, phase = phase,
                statusTitle = "售后处理中",
                statusDesc = "客服正在处理售后，请耐心等待",
                etaMinutes = "—", remainingKm = "—", arrivalClock = baseClock,
                pickup = pickup, delivery = delivery,
                rider = null, heading = 0f, progress = 0f, stepProgress = 0.9f,
                traveledKm = 0.0,
                steps = listOf(
                    TrackStep("已送达", StepState.DONE),
                    TrackStep("申请售后", StepState.DONE),
                    TrackStep("售后处理", StepState.ACTIVE),
                    TrackStep("处理完成", StepState.TODO),
                ),
                riderName = riderName, riderMeta = riderMeta,
                showRiderCard = false, canCallRider = false,
            )
        }
    }

    // ---------------- 工具 ----------------

    private const val EARTH_RADIUS_M = 6_371_000.0
    private const val AVG_KM_PER_MIN = 0.33 // 骑手平均 20km/h

    /** 起点→终点航向角（0 北、90 东，顺时针）。 */
    fun bearing(from: TrackPoint, to: TrackPoint): Float {
        if (from == to) return 0f
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

    private val totalKm: Double get() = totalMeters / 1000.0

    private fun etaClock(order: Order): String =
        SimpleDateFormat("HH:mm", Locale.CHINA).format(Date(order.createdAt + 35 * 60_000L))

    private fun clockAfter(minutes: Int): String =
        SimpleDateFormat("HH:mm", Locale.CHINA).format(Date(System.currentTimeMillis() + minutes * 60_000L))

    /** 从订单地址提取送达地短名（设计稿“王家湾村”样式）。 */
    private fun shortTarget(order: Order): String {
        val raw = order.address
        val part = raw.substringAfterLast(" · ").substringBefore("\n").trim()
        return part.ifBlank { "王家湾村" }
    }

    // ---------------- 几何辅助 ----------------

    private fun lerp(a: TrackPoint, b: TrackPoint, t: Double): TrackPoint =
        TrackPoint(
            latitude = a.latitude + (b.latitude - a.latitude) * t,
            longitude = a.longitude + (b.longitude - a.longitude) * t,
        )

    private fun buildSegments(points: List<TrackPoint>): List<Double> =
        points.zipWithNext { a, b -> distanceMeters(a, b) }

    /** Chaikin 角切割平滑：让乡道路线在地图上呈现连续弧线而非硬折角。 */
    private fun smooth(points: List<TrackPoint>, iterations: Int): List<TrackPoint> {
        if (points.size < 3) return points
        var current = points
        repeat(iterations) {
            val next = ArrayList<TrackPoint>(current.size * 2)
            next.add(current.first())
            for (i in 0 until current.size - 1) {
                next.add(lerp(current[i], current[i + 1], 0.25))
                next.add(lerp(current[i], current[i + 1], 0.75))
            }
            next.add(current.last())
            current = next
        }
        return current
    }
}
