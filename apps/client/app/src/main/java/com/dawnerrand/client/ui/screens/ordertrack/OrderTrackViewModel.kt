package com.dawnerrand.client.ui.screens.ordertrack

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.OrderStatus
import com.dawnerrand.client.data.track.OrderTrack
import com.dawnerrand.client.data.track.TrackingRepository
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * 配送进度页 ViewModel（MVVM 数据驱动）。
 *
 * - 持有唯一的 [uiState] 数据源：状态卡、进度步骤、骑手卡、地图标记全部由此驱动；
 * - 配送中订单启动演示模拟：每 2s 推进骑手进度 0.04，直至送达；
 * - 后端追踪接口就绪后仅需替换 TrackingRepository，本类与 UI 零改动。
 */
class OrderTrackViewModel(private val orderId: String) : ViewModel() {

    private val _uiState = MutableStateFlow<OrderTrack?>(null)
    val uiState: StateFlow<OrderTrack?> = _uiState.asStateFlow()

    init {
        val order = DemoRepository.orderById(orderId)
        if (order == null) {
            _uiState.value = null
        } else if (order.status == OrderStatus.DELIVERING) {
            _uiState.value = TrackingRepository.stateFor(order, 0f)
            simulateDelivery(order)
        } else {
            _uiState.value = TrackingRepository.stateFor(order, 1f)
        }
    }

    /** 演示模拟：骑手沿乡道路线匀速移动，进度 0 → 1。 */
    private fun simulateDelivery(order: com.dawnerrand.client.data.Order) {
        viewModelScope.launch {
            var progress = 0f
            while (progress < 1f) {
                delay(2_000)
                progress = (progress + 0.04f).coerceAtMost(1f)
                _uiState.value = TrackingRepository.stateFor(order, progress)
            }
        }
    }

    companion object {
        fun factory(orderId: String) = viewModelFactory {
            initializer { OrderTrackViewModel(orderId) }
        }
    }
}
