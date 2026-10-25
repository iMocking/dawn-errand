package com.dawnerrand.client.ui.screens.ordertrack

import android.animation.ValueAnimator
import android.content.ActivityNotFoundException
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Color
import android.net.Uri
import android.os.Bundle
import android.view.View
import android.view.animation.LinearInterpolator
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import androidx.core.view.WindowCompat
import androidx.lifecycle.lifecycleScope
import com.amap.api.maps.AMap
import com.amap.api.maps.CameraUpdateFactory
import com.amap.api.maps.model.LatLng
import com.amap.api.maps.model.LatLngBounds
import com.amap.api.maps.model.Marker
import com.amap.api.maps.model.MarkerOptions
import com.amap.api.maps.model.Polyline
import com.amap.api.maps.model.PolylineOptions
import com.dawnerrand.client.R
import com.dawnerrand.client.data.track.OrderTrack
import com.dawnerrand.client.data.track.StepState
import com.dawnerrand.client.data.track.TrackPhase
import com.dawnerrand.client.data.track.TrackingRepository
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padBottomWithNavBar
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast
import com.dawnerrand.client.ui.screens.ordertrack.OrderTrackViewModel.Companion.factory
import kotlinx.coroutines.launch

/**
 * 配送进度（高德地图版）：真实地图 + 取货/送达/骑手标记 + 乡道路线 + 骑手实时移动模拟。
 *
 * - MVVM 数据驱动：状态卡、进度步骤、骑手卡、地图标记全部由 [OrderTrackViewModel.uiState] 驱动；
 * - 生命周期与 MapView 严格绑定（onCreate/onResume/onPause/onDestroy/onSaveInstanceState）；
 * - 「我的位置」在授予定位权限后开启，高德 Key 未配置时地图不渲染（logcat 提示鉴权失败）。
 */
class OrderTrackActivity : AppCompatActivity() {

    private val viewModel: OrderTrackViewModel by viewModels {
        factory(intent.getStringExtra(EXTRA_ORDER_ID).orEmpty())
    }

    private lateinit var mapView: com.amap.api.maps.MapView
    private var aMap: AMap? = null
    private var locationSource: AmapLocationSource? = null

    private var pickupMarker: Marker? = null
    private var destMarker: Marker? = null
    private var riderMarker: Marker? = null
    private var routeMain: Polyline? = null
    private var routeOutline: Polyline? = null

    private var riderAnimator: ValueAnimator? = null
    private var mapReady = false
    private var fitted = false
    private var followingRider = true

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // 地图全屏铺满：内容延伸到状态栏/导航栏之后，由顶栏自行处理安全区
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
    }

    private fun bindActions() {
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        findViewById<View>(R.id.btnChatRider).setOnClickListener {
            toast("在线联系骑手（演示环境）")
        }
        findViewById<View>(R.id.btnCallRider).setOnClickListener { callRider() }
    }

    // ---------------- 地图 ----------------

    private fun setupMap() {
        val map = mapView.map ?: return
        aMap = map.apply {
            uiSettings.isZoomControlsEnabled = false
            uiSettings.isCompassEnabled = false
            uiSettings.isScaleControlsEnabled = true
            // 用户拖动地图后停止自动跟随骑手
            setOnMapTouchListener { _ -> followingRider = false }
            setOnMapLoadedListener {
                mapReady = true
                latest?.let { applyToMap(it) }
            }
        }
        locationSource = AmapLocationSource(this)
        map.setLocationSource(locationSource)
        enableMyLocationIfPermitted()
    }

    private fun enableMyLocationIfPermitted() {
        val map = aMap ?: return
        val granted = ContextCompat.checkSelfPermission(this, android.Manifest.permission.ACCESS_FINE_LOCATION) ==
            PackageManager.PERMISSION_GRANTED ||
            ContextCompat.checkSelfPermission(this, android.Manifest.permission.ACCESS_COARSE_LOCATION) ==
            PackageManager.PERMISSION_GRANTED
        if (granted) {
            map.isMyLocationEnabled = true
            map.uiSettings.isMyLocationButtonEnabled = true
        } else {
            requestPermissions(
                arrayOf(
                    android.Manifest.permission.ACCESS_FINE_LOCATION,
                    android.Manifest.permission.ACCESS_COARSE_LOCATION,
                ),
                REQ_LOCATION,
            )
        }
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray,
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == REQ_LOCATION) {
            enableMyLocationIfPermitted()
        }
    }

    // ---------------- 状态驱动 ----------------

    private var latest: OrderTrack? = null

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
                if (mapReady) applyToMap(state)
            }
        }
    }

    private fun renderState(state: OrderTrack) {
        findViewById<TextView>(R.id.tvTrackStatus).text = state.statusTitle
        findViewById<TextView>(R.id.tvTrackDesc).text = state.statusDesc
        findViewById<TextView>(R.id.tvEtaMinutes).text = state.etaMinutes
        findViewById<TextView>(R.id.tvEtaKm).text = state.remainingKm

        findViewById<View>(R.id.riderCard).visibility =
            if (state.showRiderCard) View.VISIBLE else View.GONE
        findViewById<TextView>(R.id.tvRiderName).text = state.riderName
        findViewById<TextView>(R.id.tvRiderMeta).text = state.riderMeta
        findViewById<View>(R.id.btnCallRider).visibility =
            if (state.canCallRider) View.VISIBLE else View.INVISIBLE
        findViewById<View>(R.id.btnChatRider).visibility =
            if (state.showRiderCard) View.VISIBLE else View.INVISIBLE

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
                    if (step.state == StepState.TODO) Design.TEXT_TERTIARY else Design.TEXT_PRIMARY)
                texts[index].typeface =
                    android.graphics.Typeface.DEFAULT_BOLD.takeIf { step.state == StepState.ACTIVE }
                        ?: android.graphics.Typeface.DEFAULT
            }
        }
    }

    /** 地图渲染：标记、路线、相机。 */
    private fun applyToMap(state: OrderTrack) {
        val map = aMap ?: return
        val route = TrackingRepository.routePoints.map { it.toLatLng() }

        if (routeMain == null) {
            routeOutline = map.addPolyline(
                PolylineOptions().addAll(route).width(dp(9).toFloat())
                    .color(Design.WHITE).setDottedLine(false).zIndex(1f))
            routeMain = map.addPolyline(
                PolylineOptions().addAll(route).width(dp(5).toFloat())
                    .color(Design.PRIMARY).setDottedLine(false).zIndex(2f))
        }
        if (pickupMarker == null) {
            pickupMarker = map.addMarker(
                MarkerOptions().position(state.pickup.toLatLng())
                    .icon(MapPins.pickup(this)).anchor(0.5f, 0.5f).zIndex(10f))
        }
        if (destMarker == null) {
            destMarker = map.addMarker(
                MarkerOptions().position(state.delivery.toLatLng())
                    .icon(MapPins.destination(this)).anchor(0.5f, 0.5f).zIndex(10f))
        }

        val rider = state.rider
        if (rider != null) {
            if (riderMarker == null) {
                riderMarker = map.addMarker(
                    MarkerOptions().position(rider.toLatLng())
                        .icon(MapPins.rider(this)).anchor(0.5f, 0.5f).zIndex(20f))
            } else {
                animateRiderMarker(riderMarker!!.position, rider.toLatLng())
            }
            riderMarker!!.rotateAngle = state.heading
            if (state.phase == TrackPhase.DELIVERING && followingRider) {
                map.animateCamera(CameraUpdateFactory.changeLatLng(rider.toLatLng()), 1600, null)
            }
        } else {
            riderMarker?.remove()
            riderMarker = null
        }

        if (!fitted) {
            fitted = true
            val bounds = LatLngBounds.Builder()
                .include(state.pickup.toLatLng())
                .include(state.delivery.toLatLng())
                .build()
            map.animateCamera(CameraUpdateFactory.newLatLngBounds(bounds, dp(110)))
        }
    }

    private fun animateRiderMarker(start: LatLng, target: LatLng) {
        riderAnimator?.cancel()
        riderAnimator = ValueAnimator.ofFloat(0f, 1f).apply {
            duration = 1600L
            interpolator = LinearInterpolator()
            addUpdateListener { anim ->
                val f = anim.animatedValue as Float
                val lat = start.latitude + (target.latitude - start.latitude) * f
                val lng = start.longitude + (target.longitude - start.longitude) * f
                riderMarker?.position = LatLng(lat, lng)
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
        riderAnimator?.cancel()
        locationSource?.deactivate()
        mapView.onDestroy()
        super.onDestroy()
    }

    private fun com.dawnerrand.client.data.track.TrackPoint.toLatLng(): LatLng =
        LatLng(latitude, longitude)

    companion object {
        const val EXTRA_ORDER_ID = "order_id"
        private const val REQ_LOCATION = 1001
    }
}
