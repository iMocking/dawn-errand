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
        _records.value = filtered(index)
    }

    /** 从详情页返回后重新读取仓库快照（仓库已在取货/送达成功后刷新）。 */
    fun reload() {
        _records.value = filtered(_selectedTab.value ?: 0)
    }

    // Tab 按后端真实状态筛选：此前「配送中」写死空列表，等于告诉骑手没有进行中的单。
    private fun filtered(index: Int): List<OrderRecord> = when (index) {
        1 -> all.filter { it.status == "待取货" || it.status == "配送中" }
        2 -> all.filter { it.status == "已送达" || it.status == "待评价" || it.status == "已完成" }
        3 -> all.filter { it.status == "已取消" || it.status == "已退款" || it.status == "异常" }
        else -> all
    }
}
