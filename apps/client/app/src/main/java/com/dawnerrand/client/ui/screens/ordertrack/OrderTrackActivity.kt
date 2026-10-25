package com.dawnerrand.client.ui.screens.ordertrack

import android.animation.ValueAnimator
import android.content.ActivityNotFoundException
import android.content.Intent
import android.content.res.ColorStateList
import android.graphics.Color
import android.net.Uri
import android.os.Bundle
import android.os.SystemClock
import android.view.View
import android.view.animation.LinearInterpolator
import android.widget.ImageView
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.WindowCompat
import androidx.lifecycle.lifecycleScope
import com.amap.api.maps.AMap
import com.amap.api.maps.CameraUpdateFactory
import com.amap.api.maps.model.CameraPosition
import com.amap.api.maps.model.Circle
import com.amap.api.maps.model.CircleOptions
import com.amap.api.maps.model.LatLng
import com.amap.api.maps.model.LatLngBounds
import com.amap.api.maps.model.Marker
import com.amap.api.maps.model.MarkerOptions
import com.amap.api.maps.model.Polyline
import com.amap.api.maps.model.PolylineOptions
import com.amap.api.services.core.AMapException
import com.amap.api.services.core.LatLonPoint
import com.amap.api.services.route.DriveRouteResult
import com.amap.api.services.route.RouteSearch
import com.dawnerrand.client.R
import com.dawnerrand.client.data.track.OrderTrack
import com.dawnerrand.client.data.track.StepState
import com.dawnerrand.client.data.track.TrackPhase
import com.dawnerrand.client.data.track.TrackPoint
import com.dawnerrand.client.data.track.TrackingRepository
import com.dawnerrand.client.ui.screens.ordertrack.OrderTrackViewModel.Companion.factory
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padBottomWithNavBar
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast
import kotlinx.coroutines.launch

/**
 * 配送进度（高德地图版 v1.2.1）：参考美团/饿了么追踪页的视觉与行为。
 *
 * - 地图：白色路线外框 + 灰色剩余路线 + 橙色已走路线；取货/送达气泡图钉；
 *   骑手圆标 + 方向箭头 + 脉冲圈；初始视角覆盖全路线，配送中平滑跟随骑手；
 * - 交互：用户拖动地图暂停跟随，右下角目标按钮一键回到骑手；
 * - 状态：浮层状态卡 + 横向 4 步进度 + 骑手卡，全部由 [OrderTrackViewModel.uiState] 驱动。
 */
class OrderTrackActivity : AppCompatActivity() {

    private val viewModel: OrderTrackViewModel by viewModels {
        factory(intent.getStringExtra(EXTRA_ORDER_ID).orEmpty())
    }

    private lateinit var mapView: com.amap.api.maps.MapView
    private var aMap: AMap? = null

    private var pickupMarker: Marker? = null
    private var destMarker: Marker? = null
    private var riderMarker: Marker? = null
    private var riderPulse: Circle? = null
    private var routeCasing: Polyline? = null
    private var routeRemaining: Polyline? = null
    private var routeTraveled: Polyline? = null
    private var routeSearch: RouteSearch? = null
    private var appliedRoute: List<LatLng> = emptyList()

    private var pulseAnimator: ValueAnimator? = null
    private var latest: OrderTrack? = null
    private var mapLoaded = false
    private var fitted = false
    private var followingRider = true
    private var followAfter = 0L
    private var lastCameraAt = 0L
    private var mapSetupAttempts = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // 全屏地图：内容延伸到状态栏/导航栏之后，由浮层自行处理安全区
        WindowCompat.setDecorFitsSystemWindows(window, false)
        window.statusBarColor = Color.TRANSPARENT
        setContentView(R.layout.activity_order_track)
        setLightStatusBars(false)

        mapView = findViewById(R.id.mapView)
        mapView.onCreate(savedInstanceState)
        findViewById<View>(R.id.topBar).padTopWithStatusBar()
        findViewById<View>(R.id.riderCard).padBottomWithNavBar(8)

        bindActions()
        setupMap()
        observeState()
        updateFollowUi()
    }

    private fun bindActions() {
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        findViewById<View>(R.id.btnChatRider).setOnClickListener {
            toast("在线联系骑手（演示环境）")
        }
        findViewById<View>(R.id.btnCallRider).setOnClickListener { callRider() }
        findViewById<View>(R.id.btnRecenter).setOnClickListener {
            followingRider = true
            updateFollowUi()
            val state = latest
            if (state?.rider != null) {
                followRider(state.rider, animateMs = 800)
            } else {
                fitted = false
                fitRouteIfNeeded()
            }
        }
    }

    // ---------------- 地图 ----------------

    private fun setupMap() {
        if (aMap != null) return
        val map = mapView.map
        if (map == null) {
            // MapView 在部分设备/Key 鉴权阶段会延迟创建 Map 对象，重试几次而不是放弃
            if (mapSetupAttempts++ < 20) {
                mapView.postDelayed({ setupMap() }, 200)
            }
            return
        }
        aMap = map.apply {
            mapType = AMap.MAP_TYPE_NORMAL
            uiSettings.isZoomControlsEnabled = false
            uiSettings.isCompassEnabled = false
            uiSettings.isScaleControlsEnabled = false
            uiSettings.isRotateGesturesEnabled = false
            uiSettings.isTiltGesturesEnabled = false
            isMyLocationEnabled = false
            // 用户拖动地图后暂停自动跟随，避免和手势抢相机
            setOnMapTouchListener { _ ->
                if (followingRider) {
                    followingRider = false
                    updateFollowUi()
                }
            }
            setOnMapLoadedListener {
                mapLoaded = true
                // 地图引擎就绪后重新做一次全路线取景，避免在引擎未 ready 时的相机指令失效
                fitted = false
                fitRouteIfNeeded()
                latest?.let { applyToMap(it) }
            }
        }
        // Key 鉴权失败时 onMapLoaded 可能不回调；仍要尽早绘制路线与标记
        mapView.postDelayed({
            fitRouteIfNeeded()
            latest?.let { applyToMap(it) }
        }, 500)
        requestRoadRoute()
    }

    /**
     * 用高德驾车路径规划替换演示折线，让路线贴合真实道路；
     * 搜索失败（无网络/Key 未开通搜索权限）时保留本地乡道路线。
     */
    private fun requestRoadRoute() {
        if (routeSearch != null) return
        val search = RouteSearch(this)
        routeSearch = search
        search.setRouteSearchListener(object : RouteSearch.OnRouteSearchListener {
            override fun onDriveRouteSearched(result: DriveRouteResult?, errorCode: Int) {
                if (errorCode != AMapException.CODE_AMAP_SUCCESS) return
                val path = result?.paths?.firstOrNull() ?: return
                val points = ArrayList<TrackPoint>()
                path.steps?.forEach { step ->
                    step.polyline?.forEach { point ->
                        val next = TrackPoint(point.latitude, point.longitude)
                        if (points.isEmpty() ||
                            points.last().latitude != next.latitude ||
                            points.last().longitude != next.longitude
                        ) {
                            points.add(next)
                        }
                    }
                }
                if (points.size < 2) return
                runOnUiThread {
                    if (isFinishing || isDestroyed) return@runOnUiThread
                    TrackingRepository.setRoute(points)
                    fitted = false
                    fitRouteIfNeeded()
                    latest?.let { applyToMap(it) }
                }
            }

            override fun onBusRouteSearched(result: com.amap.api.services.route.BusRouteResult?, errorCode: Int) = Unit
            override fun onWalkRouteSearched(result: com.amap.api.services.route.WalkRouteResult?, errorCode: Int) = Unit
            override fun onRideRouteSearched(result: com.amap.api.services.route.RideRouteResult?, errorCode: Int) = Unit
        })
        val fromTo = RouteSearch.FromAndTo(
            LatLonPoint(TrackingRepository.pickup.latitude, TrackingRepository.pickup.longitude),
            LatLonPoint(TrackingRepository.delivery.latitude, TrackingRepository.delivery.longitude),
        )
        val query = RouteSearch.DriveRouteQuery(
            fromTo,
            RouteSearch.DRIVING_SINGLE_DEFAULT,
            null,
            null,
            "",
        )
        search.calculateDriveRouteAsyn(query)
    }

    private fun fitRouteIfNeeded() {
        if (fitted) return
        val map = aMap ?: return
        if (mapView.width <= 0 || mapView.height <= 0) return
        val bounds = LatLngBounds.Builder()
            .include(TrackingRepository.pickup.toLatLng())
            .include(TrackingRepository.delivery.toLatLng())
            .build()
        fitted = true
        followAfter = SystemClock.elapsedRealtime() + 1800
        // 首次取景用 moveCamera（同步），保证 Map 引擎延迟初始化时也能落到路线上
        map.moveCamera(CameraUpdateFactory.newLatLngBounds(bounds, dp(96)))
        if (!mapLoaded) {
            // 若引擎稍后才 loaded，onMapLoaded 会重置 fitted 并重新取景
            mapView.postDelayed({ if (!mapLoaded) fitted = false }, 1200)
        }
    }

    private fun followRider(point: TrackPoint, animateMs: Long = 900) {
        val map = aMap ?: return
        val camera = CameraPosition(
            point.toLatLng(),
            FOLLOW_ZOOM,
            FOLLOW_TILT,
            0f, // 保持正北朝上，避免地图旋转造成阅读负担
        )
        map.animateCamera(CameraUpdateFactory.newCameraPosition(camera), animateMs, null)
        lastCameraAt = SystemClock.elapsedRealtime()
    }

    // ---------------- 状态驱动 ----------------

    private fun observeState() {
        lifecycleScope.launch {
            viewModel.uiState.collect { state ->
                if (state == null) {
                    toast("暂无配送中的订单")
                    finish()
                    return@collect
                }
                latest = state
                renderState(state)
                applyToMap(state)
            }
        }
    }

    private fun renderState(state: OrderTrack) {
        findViewById<TextView>(R.id.tvTrackStatus).text = state.statusTitle
        findViewById<TextView>(R.id.tvTrackDesc).text = state.statusDesc
        findViewById<TextView>(R.id.tvEtaMinutes).text = state.etaMinutes
        findViewById<TextView>(R.id.tvEtaKm).text = state.remainingKm
        findViewById<TextView>(R.id.tvEtaClock).text = state.arrivalClock
        findViewById<TextView>(R.id.tvEtaPill).text = pillText(state)
        findViewById<TextView>(R.id.tvLiveBadge).text = liveText(state)

        val riderVisible = state.showRiderCard
        findViewById<View>(R.id.riderCard).visibility = if (riderVisible) View.VISIBLE else View.GONE
        findViewById<TextView>(R.id.tvRiderName).text = state.riderName
        findViewById<TextView>(R.id.tvRiderMeta).text = state.riderMeta
        findViewById<View>(R.id.btnCallRider).visibility =
            if (state.canCallRider) View.VISIBLE else View.INVISIBLE
        findViewById<View>(R.id.btnChatRider).visibility =
            if (riderVisible) View.VISIBLE else View.INVISIBLE
        updateRecenterPosition(riderVisible)

        val dots = arrayOf(
            findViewById<View>(R.id.dot1),
            findViewById<View>(R.id.dot2),
            findViewById<View>(R.id.dot3),
            findViewById<View>(R.id.dot4),
        )
        val texts = arrayOf(
            findViewById<TextView>(R.id.step1),
            findViewById<TextView>(R.id.step2),
            findViewById<TextView>(R.id.step3),
            findViewById<TextView>(R.id.step4),
        )
        state.steps.forEachIndexed { index, step ->
            if (index < dots.size) {
                dots[index].setBackgroundResource(
                    when (step.state) {
                        StepState.DONE -> R.drawable.bg_dot_done
                        StepState.ACTIVE -> R.drawable.bg_dot_active
                        StepState.TODO -> R.drawable.bg_dot_todo
                    })
                texts[index].text = step.text
                texts[index].setTextColor(
                    when (step.state) {
                        StepState.TODO -> Design.TEXT_TERTIARY
                        StepState.ACTIVE -> Design.PRIMARY
                        StepState.DONE -> Design.TEXT_PRIMARY
                    })
                texts[index].typeface =
                    android.graphics.Typeface.DEFAULT_BOLD.takeIf { step.state == StepState.ACTIVE }
                        ?: android.graphics.Typeface.DEFAULT
            }
        }
        updateProgressLine(state.stepProgress)
    }

    private fun pillText(state: OrderTrack): String = when {
        state.etaMinutes != "—" -> "约 ${state.etaMinutes} 分钟"
        state.phase == TrackPhase.WAITING_ACCEPT -> "待接单"
        state.phase == TrackPhase.NOT_PAID -> "待支付"
        state.phase == TrackPhase.ARRIVED -> "已送达"
        state.phase == TrackPhase.CANCELED -> "已取消"
        state.phase == TrackPhase.REFUND -> "售后中"
        else -> "配送中"
    }

    private fun liveText(state: OrderTrack): String = when (state.phase) {
        TrackPhase.NOT_PAID -> "待支付"
        TrackPhase.WAITING_ACCEPT -> "待接单"
        TrackPhase.DELIVERING -> "配送中"
        TrackPhase.ARRIVED -> "已送达"
        TrackPhase.CANCELED -> "已取消"
        TrackPhase.REFUND -> "售后中"
    }

    private fun updateProgressLine(stepProgress: Float) {
        val track = findViewById<View>(R.id.progressTrack)
        val active = findViewById<View>(R.id.progressLineActive)
        track.post {
            val start = dp(24)
            val available = (track.width - start * 2).coerceAtLeast(1)
            val lp = active.layoutParams
            lp.width = (available * stepProgress.coerceIn(0f, 1f)).toInt()
            active.layoutParams = lp
        }
    }

    private fun updateRecenterPosition(showRider: Boolean) {
        val btn = findViewById<View>(R.id.btnRecenter)
        val lp = btn.layoutParams as android.widget.FrameLayout.LayoutParams
        lp.bottomMargin = if (showRider) dp(252) else dp(24)
        btn.layoutParams = lp
    }

    private fun updateFollowUi() {
        val btn = findViewById<ImageView>(R.id.btnRecenter)
        btn.backgroundTintList = ColorStateList.valueOf(
            if (followingRider) Color.WHITE else 0xFFFFF3E8.toInt(),
        )
        btn.contentDescription = if (followingRider) "正在跟随骑手" else "回到骑手位置"
    }

    /** 地图渲染：标记、路线、相机。 */
    private fun applyToMap(state: OrderTrack) {
        val map = aMap ?: return
        val route = TrackingRepository.routePoints.map { it.toLatLng() }

        if (routeCasing == null) {
            routeCasing = map.addPolyline(
                PolylineOptions().addAll(route).width(dp(9).toFloat())
                    .color(Color.WHITE).setDottedLine(false).zIndex(1f),
            )
            routeRemaining = map.addPolyline(
                PolylineOptions().addAll(route).width(dp(7).toFloat())
                    .color(0xFFC9CDD4.toInt()).setDottedLine(false).zIndex(2f),
            )
            routeTraveled = map.addPolyline(
                PolylineOptions().addAll(route).width(dp(7).toFloat())
                    .color(Design.PRIMARY).setDottedLine(false).zIndex(3f),
            )
            routeTraveled?.isVisible = false
        }
        val routeChanged = route.size != appliedRoute.size ||
            route.firstOrNull() != appliedRoute.firstOrNull() ||
            route.lastOrNull() != appliedRoute.lastOrNull()
        if (routeChanged) {
            appliedRoute = route
            routeCasing?.points = route
            routeRemaining?.points = route
        }

        // 已走过 / 剩余路线分层，实时随骑手进度增长
        val rider = state.rider
        if (rider != null) {
            val traveled = TrackingRepository.pointsUntil(state.progress).map { it.toLatLng() }
            val remaining = TrackingRepository.pointsAfter(state.progress).map { it.toLatLng() }
            routeTraveled?.points = traveled
            routeTraveled?.isVisible = traveled.size >= 2
            routeRemaining?.points = remaining
            routeRemaining?.isVisible = remaining.size >= 2
        } else {
            routeTraveled?.isVisible = false
            routeRemaining?.points = route
            routeRemaining?.isVisible = true
        }

        if (pickupMarker == null) {
            pickupMarker = map.addMarker(
                MarkerOptions().position(TrackingRepository.pickup.toLatLng())
                    .icon(MapPins.pickup(this)).anchor(0.5f, 0.94f).zIndex(10f),
            )
        }
        if (destMarker == null) {
            destMarker = map.addMarker(
                MarkerOptions().position(TrackingRepository.delivery.toLatLng())
                    .icon(MapPins.destination(this)).anchor(0.5f, 0.94f).zIndex(10f),
            )
        }

        if (rider != null && state.phase != TrackPhase.ARRIVED) {
            if (riderMarker == null) {
                riderMarker = map.addMarker(
                    MarkerOptions().position(rider.toLatLng())
                        .icon(MapPins.rider(this)).anchor(0.5f, 0.58f).zIndex(20f),
                )
                startRiderPulse(rider.toLatLng())
            } else {
                riderMarker?.position = rider.toLatLng()
            }
            riderMarker?.rotateAngle = state.heading
            riderPulse?.center = rider.toLatLng()
        } else {
            riderMarker?.remove()
            riderMarker = null
            riderPulse?.remove()
            riderPulse = null
            pulseAnimator?.cancel()
            pulseAnimator = null
        }

        if (state.rider != null && state.phase == TrackPhase.DELIVERING && followingRider) {
            val now = SystemClock.elapsedRealtime()
            if (now >= followAfter && now - lastCameraAt >= 1000) {
                followRider(state.rider)
            }
        }
        if (!fitted) fitRouteIfNeeded()
    }

    private fun startRiderPulse(center: LatLng) {
        val map = aMap ?: return
        riderPulse?.remove()
        riderPulse = map.addCircle(
            CircleOptions().center(center).radius(32.0)
                .fillColor(0x33FF6B2C).strokeWidth(0f).zIndex(4f),
        )
        pulseAnimator?.cancel()
        pulseAnimator = ValueAnimator.ofFloat(0f, 1f).apply {
            duration = 1700L
            repeatCount = ValueAnimator.INFINITE
            interpolator = LinearInterpolator()
            addUpdateListener { anim ->
                val f = anim.animatedValue as Float
                riderPulse?.radius = 28.0 + 56.0 * f
                val alpha = (64 * (1f - f)).toInt().coerceIn(0, 255)
                riderPulse?.fillColor = Color.argb(alpha, 255, 107, 44)
            }
            start()
        }
    }

    private fun callRider() {
        try {
            startActivity(Intent(Intent.ACTION_DIAL, Uri.parse("tel:13800001234")))
        } catch (e: ActivityNotFoundException) {
            toast("骑手电话：138-0000-1234")
        }
    }

    // ---------------- 生命周期 ----------------

    override fun onResume() {
        super.onResume()
        mapView.onResume()
    }

    override fun onPause() {
        super.onPause()
        mapView.onPause()
    }

    override fun onSaveInstanceState(outState: Bundle) {
        super.onSaveInstanceState(outState)
        mapView.onSaveInstanceState(outState)
    }

    override fun onDestroy() {
        pulseAnimator?.cancel()
        pulseAnimator = null
        riderPulse?.remove()
        riderPulse = null
        mapView.onDestroy()
        super.onDestroy()
    }

    private fun TrackPoint.toLatLng(): LatLng = LatLng(latitude, longitude)

    companion object {
        const val EXTRA_ORDER_ID = "order_id"
        private const val FOLLOW_ZOOM = 16.4f
        private const val FOLLOW_TILT = 30f
    }
}
