package com.dawnerrand.store

import android.app.Application
import com.dawnerrand.store.data.DishStore
import com.dawnerrand.store.data.FinanceStore
import com.dawnerrand.store.data.SettlementStore
import com.dawnerrand.store.data.StoreRemoteRepository
import com.dawnerrand.store.data.StoreSession
import com.dawnerrand.store.data.StoreSettings
import com.dawnerrand.store.data.api.ApiClient
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

/**
 * 初始化本地状态：登录会话 / 营业设置 / 菜品 / 财务 / 结算账户；
 * 已登录时异步拉取真实后端数据。
 */
class StoreApp : Application() {
    override fun onCreate() {
        super.onCreate()
        StoreSession.init(this)
        ApiClient.accessToken = StoreSession.token
        StoreSettings.init(this)
        DishStore.init(this)
        FinanceStore.init(this)
        SettlementStore.init(this)
        if (StoreSession.isLoggedIn()) {
            CoroutineScope(Dispatchers.IO).launch { StoreRemoteRepository.refresh(this@StoreApp) }
        }
    }
}
