package com.dawnerrand.driver

import android.app.Application
import android.util.Log
import com.amap.api.location.AMapLocationClient
import com.amap.api.maps.MapsInitializer

class DriverApp : Application() {

    override fun onCreate() {
        super.onCreate()
        // 高德地图隐私合规声明（SDK 10.x 强制：未声明视为未同意，地图鉴权失败 errorCode 10001）
        // 地图与内置定位模块的合规接口需同时声明；正式版应接入真实隐私弹窗后置 true，演示版默认同意。
        try {
            MapsInitializer.updatePrivacyShow(this, true, true)
            MapsInitializer.updatePrivacyAgree(this, true)
            AMapLocationClient.updatePrivacyShow(this, true, true)
            AMapLocationClient.updatePrivacyAgree(this, true)
            Log.i("DriverApp", "amap privacy consent OK")
        } catch (e: Throwable) {
            Log.e("DriverApp", "amap privacy consent FAILED", e)
        }
    }
}
