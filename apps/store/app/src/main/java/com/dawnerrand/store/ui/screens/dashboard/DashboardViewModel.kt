package com.dawnerrand.store.ui.screens.dashboard

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.GridItem
import com.dawnerrand.store.data.StoreRepository
import com.dawnerrand.store.data.StoreState

/**
 * 工作台：店铺信息 + 今日数据 + 待接单提醒 + 两组功能宫格。
 * 角标数量随共享订单状态联动刷新。
 */
class DashboardViewModel : ViewModel() {

    val shop = StoreRepository.shop
    val orderGrid = StoreRepository.orderGrid
    val shopGrid = StoreRepository.shopGrid

    private val _pendingCount = MutableLiveData(0)
    val pendingCount: LiveData<Int> = _pendingCount

    private val _cookingCount = MutableLiveData(0)
    val cookingCount: LiveData<Int> = _cookingCount

    init {
        StoreState.ordersLive.observeForever { orders ->
            _pendingCount.value = orders.count { it.status == com.dawnerrand.store.data.ORDER_PENDING }
            _cookingCount.value = orders.count { it.status == com.dawnerrand.store.data.ORDER_COOKING }
        }
    }

    /** 宫格角标回填（待接单/制作中） */
    fun gridWithBadges(): List<GridItem> =
        orderGrid.map {
            when (it.label) {
                "待接单" -> it.copy(badge = _pendingCount.value ?: 0)
                "制作中" -> it.copy(badge = _cookingCount.value ?: 0)
                else -> it
            }
        }
}
