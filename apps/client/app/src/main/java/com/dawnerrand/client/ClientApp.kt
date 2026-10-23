package com.dawnerrand.client

import android.app.Application
import com.dawnerrand.client.data.DemoRepository

/** 应用入口：初始化本地数据仓库（会话 / 地址 / 订单 / 购物车 / 设置等）。 */
class ClientApp : Application() {
    override fun onCreate() {
        super.onCreate()
        DemoRepository.init(this)
    }
}
