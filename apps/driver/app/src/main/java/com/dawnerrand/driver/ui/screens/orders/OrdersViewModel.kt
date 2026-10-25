package com.dawnerrand.driver.ui.screens.orders

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.driver.data.model.OrderRecord
import com.dawnerrand.driver.data.repository.DriverRepository

class OrdersViewModel : ViewModel() {

    private val all: List<OrderRecord>
        get() = DriverRepository.orderHistory

    private val _records = MutableLiveData(DriverRepository.orderHistory)
    val records: LiveData<List<OrderRecord>> = _records

    private val _selectedTab = MutableLiveData(0)
    val selectedTab: LiveData<Int> = _selectedTab

    fun selectTab(index: Int) {
        _selectedTab.value = index
        _records.value = when (index) {
            1 -> emptyList()
            2 -> all
            else -> all
        }
    }
}
