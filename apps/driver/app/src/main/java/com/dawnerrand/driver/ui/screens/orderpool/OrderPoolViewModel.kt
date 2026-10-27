package com.dawnerrand.driver.ui.screens.orderpool

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.driver.data.model.OrderItem
import com.dawnerrand.driver.data.repository.DriverRepository

class OrderPoolViewModel : ViewModel() {

    private val allOrders: List<OrderItem>
        get() = DriverRepository.orderPool

    private val _orders = MutableLiveData(DriverRepository.orderPool)
    val orders: LiveData<List<OrderItem>> = _orders

    private val _selectedFilter = MutableLiveData(0)
    val selectedFilter: LiveData<Int> = _selectedFilter

    fun selectFilter(index: Int) {
        _selectedFilter.value = index
        _orders.value = filtered(index)
    }

    /** 抢单成功后池子会少一条：返回大厅时按当前筛选重新取一次仓库快照。 */
    fun reload() {
        _orders.value = filtered(_selectedFilter.value ?: 0)
    }

    private fun filtered(index: Int): List<OrderItem> = allOrders.filter { order ->
        when (index) {
            1 -> order.type == "外卖订单"
            2 -> order.type != "外卖订单"
            3 -> order.distanceKm < 5f
            else -> true
        }
    }
}
