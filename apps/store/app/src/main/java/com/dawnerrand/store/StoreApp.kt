package com.dawnerrand.store

import android.app.Application
import com.dawnerrand.store.data.DishStore
import com.dawnerrand.store.data.FinanceStore
import com.dawnerrand.store.data.SettlementStore
import com.dawnerrand.store.data.StoreSession
import com.dawnerrand.store.data.StoreSettings

/**
 * 初始化本地状态：登录会话 / 营业设置 / 菜品 / 财务 / 结算账户。
 */
class StoreApp : Application() {
    override fun onCreate() {
        super.onCreate()
        StoreSession.init(this)
        StoreSettings.init(this)
        DishStore.init(this)
        FinanceStore.init(this)
        SettlementStore.init(this)
    }
}
