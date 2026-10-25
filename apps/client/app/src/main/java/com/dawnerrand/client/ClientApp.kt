package com.dawnerrand.client

import android.app.Application
import android.util.Log
import com.amap.api.location.AMapLocationClient
import com.amap.api.maps.MapsInitializer
import com.dawnerrand.client.data.ClientRemoteRepository
import com.dawnerrand.client.data.DemoRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

/**
 * 应用入口：初始化本地数据仓库（会话 / 地址 / 订单 / 购物车 / 设置等）。
 */
class ClientApp : Application() {
    override fun onCreate() {
        super.onCreate()
        DemoRepository.init(this)
        if (DemoRepository.session()?.token != null) {
            CoroutineScope(Dispatchers.IO).launch { ClientRemoteRepository.refresh() }
        }
        initAmapPrivacy()
    }

    /**
     * 高德地图/定位 SDK 合规初始化（V2.3.0 起强制，必须先于任何 SDK 调用）。
     *
     * 演示环境默认按「已同意」初始化，保证开箱即用；商用上线必须改为：
     * 在用户主动勾选《隐私政策》后回调这两个方法（登录/注册页已有 LegalActivity 入口）。
     */
    private fun initAmapPrivacy() {
        try {
            MapsInitializer.updatePrivacyShow(this, true, true)
            MapsInitializer.updatePrivacyAgree(this, true)
            AMapLocationClient.updatePrivacyShow(this, true, true)
            AMapLocationClient.updatePrivacyAgree(this, true)
            Log.i("AmapPrivacy", "privacy agreed: map+loc ok")
        } catch (t: Throwable) {
            Log.e("AmapPrivacy", "privacy init failed", t)
        }
    }
}
