package com.dawnerrand.client.ui.screens.ordertrack

import android.content.Context
import com.amap.api.location.AMapLocation
import com.amap.api.location.AMapLocationClient
import com.amap.api.location.AMapLocationClientOption
import com.amap.api.location.AMapLocationListener
import com.amap.api.maps.LocationSource

/**
 * 高德定位数据源：为地图「我的位置」蓝点与定位按钮提供实时位置。
 * 仅在使用位置权限时启用（由 [AmapLocationSource] 的调用方做权限判断）。
 */
class AmapLocationSource(private val context: Context) :
    LocationSource,
    AMapLocationListener {

    private var client: AMapLocationClient? = null
    private var listener: LocationSource.OnLocationChangedListener? = null

    override fun activate(listener: LocationSource.OnLocationChangedListener?) {
        this.listener = listener
        val option = AMapLocationClientOption().apply {
            locationMode = AMapLocationClientOption.AMapLocationMode.Hight_Accuracy
            isOnceLocation = false
            interval = 5_000
        }
        client = AMapLocationClient(context).apply {
            setLocationListener(this@AmapLocationSource)
            setLocationOption(option)
        }
        client?.startLocation()
    }

    override fun deactivate() {
        listener = null
        client?.stopLocation()
        client?.onDestroy()
        client = null
    }

    override fun onLocationChanged(location: AMapLocation?) {
        if (location != null && location.errorCode == 0) {
            listener?.onLocationChanged(location)
        }
    }
}
