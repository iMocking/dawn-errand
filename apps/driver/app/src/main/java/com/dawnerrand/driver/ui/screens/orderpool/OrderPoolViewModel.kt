package com.dawnerrand.driver.ui.screens.orderpool

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.driver.data.model.OrderItem
import com.dawnerrand.driver.data.repository.DriverRepository

class OrderPoolViewModel : ViewModel() {

    private val allOrders = DriverRepository.orderPool

    private val _orders = MutableLiveData(allOrders)
    val orders: LiveData<List<OrderItem>> = _orders

    private val _selectedFilter = MutableLiveData(0)
    val selectedFilter: LiveData<Int> = _selectedFilter

    fun selectFilter(index: Int) {
        _selectedFilter.value = index
        _orders.value = allOrders.filter { order ->
            when (index) {
                1 -> order.type == "外卖订单"
                2 -> order.type != "外卖订单"
                3 -> order.distanceKm < 5f
                else -> true
            }
        }
    }
}
