package com.dawnerrand.store.ui.screens.orders

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.ORDER_COOKING
import com.dawnerrand.store.data.ORDER_DELIVERING
import com.dawnerrand.store.data.ORDER_DONE
import com.dawnerrand.store.data.ORDER_PENDING
import com.dawnerrand.store.data.StoreOrder
import com.dawnerrand.store.data.StoreState

/**
 * 订单管理：四个状态 Tab 过滤 + 接单/拒单/出餐操作，数据来自共享订单状态。
 */
class OrdersViewModel : ViewModel() {

    private val _currentTab = MutableLiveData(0)
    val currentTab: LiveData<Int> = _currentTab

    private val _filtered = MutableLiveData<List<StoreOrder>>(emptyList())
    val filtered: LiveData<List<StoreOrder>> = _filtered

    init {
        StoreState.ordersLive.observeForever { refresh() }
    }

    fun selectTab(index: Int) {
        _currentTab.value = index
        refresh()
    }

    fun accept(id: String) = StoreState.accept(id)

    fun reject(id: String) = StoreState.reject(id)

    fun notifyReady(id: String) = StoreState.notifyReady(id)

    private fun refresh() {
        val all = StoreState.ordersLive.value ?: return
        val status = when (_currentTab.value ?: 0) {
            0 -> ORDER_PENDING
            1 -> ORDER_COOKING
            2 -> ORDER_DELIVERING
            else -> ORDER_DONE
        }
        _filtered.value = all.filter { it.status == status }
    }
}
