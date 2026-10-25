package com.dawnerrand.driver.ui.screens.navigation

import android.Manifest
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Color
import android.os.Bundle
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.annotation.DrawableRes
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.amap.api.maps.AMap
import com.amap.api.maps.CameraUpdateFactory
import com.amap.api.maps.MapView
import com.amap.api.maps.model.BitmapDescriptor
import com.amap.api.maps.model.BitmapDescriptorFactory
import com.amap.api.maps.model.CameraPosition
import com.amap.api.maps.model.LatLng
import com.amap.api.maps.model.MarkerOptions
import com.amap.api.maps.model.PolylineOptions
import com.dawnerrand.driver.R
import com.dawnerrand.driver.databinding.ActivityNavigationBinding
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.padBottomWithNavBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 导航页：还原 design/driver/navigation.html。
 * 已接入高德地图 SDK（3dmap）：真实底图 + 取餐点/送达点标记 + 路线 + 定位蓝点。
 * 注意：地图底图显示依赖「高德开放平台 API Key」（AndroidManifest 中 com.amap.v2.ApiKey）。
 * 未配置 Key 时底图不加载，顶部覆盖与底部面板等 UI 不受影响。
 */
class NavigationActivity : AppCompatActivity() {

    private lateinit var binding: ActivityNavigationBinding
    private var mapView: MapView? = null
    private var aMap: AMap? = null

    /** 演示点位（浠水县绿杨乡） */
    private val pickPoint = LatLng(30.4525, 115.2612)   // 取餐点：镇中心
    private val dropPoint = LatLng(30.4581, 115.2684)   // 送达点：王家湾村

    private val locationPermLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
            if (granted) {
                enableMyLocation()
            } else {
                toast("未授予定位权限，将无法显示实时位置")
            }
        }

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.dark(Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityNavigationBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.btnNavBack.setOnClickListener { finish() }
        binding.btnStartNav.setOnClickListener { toast("开始导航，请沿当前路线前往取餐点") }

        binding.root.padBottomWithNavBar(extra = 0)

        mapView = binding.mapView
        mapView?.onCreate(savedInstanceState)
        mapView?.map?.let { initMap(it) }

        requestLocationPermissionIfNeeded()
    }

    private fun initMap(map: AMap) {
        aMap = map
        with(map.uiSettings) {
            isZoomControlsEnabled = false
            isCompassEnabled = false
            isScaleControlsEnabled = false
            isRotateGesturesEnabled = false
            isTiltGesturesEnabled = false
            logoPosition = 3 // 右下角显示高德 logo
        }
        // 取餐点 / 送达点标记（竞品风气泡图钉：尖角指向坐标，参照美团骑手/饿了么蜂鸟）
        map.addMarker(
            MarkerOptions().position(pickPoint)
                .title("取餐点")
                .snippet("镇中心 · 村口老陈家常菜")
                .icon(markerIcon(R.drawable.bg_marker_pickup))
                .anchor(0.5f, 1f)
        )
        map.addMarker(
            MarkerOptions().position(dropPoint)
                .title("送达点")
                .snippet("王家湾村三组26号")
                .icon(markerIcon(R.drawable.bg_marker_drop))
                .anchor(0.5f, 1f)
        )
        // 配送路线（橙色实线 + 白色描边感：加一条更宽的浅色底）
        map.addPolyline(
            PolylineOptions().add(pickPoint, dropPoint)
                .width(12f).color(Color.argb(90, 255, 255, 255))
        )
        map.addPolyline(
            PolylineOptions().add(pickPoint, dropPoint)
                .width(7f).color(Design.PRIMARY)
        )
        // 相机初始视角：包含取送两点
        map.moveCamera(
            CameraUpdateFactory.newCameraPosition(
                CameraPosition(
                    LatLng((pickPoint.latitude + dropPoint.latitude) / 2, (pickPoint.longitude + dropPoint.longitude) / 2),
                    15f, 0f, 0f,
                )
            )
        )
    }

    /** 矢量标记渲染为 BitmapDescriptor（fromResource 对 vector 支持不稳定，显式绘制最可靠） */
    private fun markerIcon(@DrawableRes resId: Int): BitmapDescriptor {
        val d = ContextCompat.getDrawable(this, resId) ?: return BitmapDescriptorFactory.defaultMarker()
        val bmp = Bitmap.createBitmap(d.intrinsicWidth, d.intrinsicHeight, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bmp)
        d.setBounds(0, 0, canvas.width, canvas.height)
        d.draw(canvas)
        return BitmapDescriptorFactory.fromBitmap(bmp)
    }

    private fun requestLocationPermissionIfNeeded() {
        val fine = ContextCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION)
        if (fine == PackageManager.PERMISSION_GRANTED) {
            enableMyLocation()
        } else {
            locationPermLauncher.launch(Manifest.permission.ACCESS_FINE_LOCATION)
        }
    }

    private fun enableMyLocation() {
        try {
            aMap?.isMyLocationEnabled = true
            aMap?.setMyLocationType(AMap.LOCATION_TYPE_LOCATE)
        } catch (_: Exception) {
            // 未配置高德 Key 时定位不可用，忽略
        }
    }

    override fun onResume() {
        super.onResume()
        mapView?.onResume()
    }

    override fun onPause() {
        super.onPause()
        mapView?.onPause()
    }

    override fun onSaveInstanceState(outState: Bundle) {
        super.onSaveInstanceState(outState)
        mapView?.onSaveInstanceState(outState)
    }

    override fun onDestroy() {
        super.onDestroy()
        mapView?.onDestroy()
        mapView = null
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, NavigationActivity::class.java))
        }
    }
}
