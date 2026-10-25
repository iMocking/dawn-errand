package com.dawnerrand.client.ui.screens.ordertrack

import android.os.SystemClock
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
 * - 配送中订单启动演示模拟：约 42s 走完全程，每 100ms 推进一次，配合地图 8~10fps
 *   的插值已达到平滑跟车效果（真机后端推送时直接替换数据源）；
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
            _uiState.value = TrackingRepository.stateFor(order, 0.02f)
            simulateDelivery(order)
        } else {
            _uiState.value = TrackingRepository.stateFor(order, 1f)
        }
    }

    /** 演示模拟：按真实时间线性推进 0→1，骑手沿平滑路线匀速移动。 */
    private fun simulateDelivery(order: com.dawnerrand.client.data.Order) {
        viewModelScope.launch {
            val startedAt = SystemClock.elapsedRealtime()
            var lastProgress = -1f
            while (true) {
                val elapsed = SystemClock.elapsedRealtime() - startedAt
                val progress = (elapsed.toFloat() / TOTAL_DURATION_MS).coerceIn(0f, 1f)
                if (progress - lastProgress >= 0.002f || progress >= 1f) {
                    lastProgress = progress
                    _uiState.value = TrackingRepository.stateFor(order, progress)
                }
                if (progress >= 1f) break
                delay(TICK_MS)
            }
        }
    }

    companion object {
        private const val TOTAL_DURATION_MS = 42_000L
        private const val TICK_MS = 100L

        fun factory(orderId: String) = viewModelFactory {
            initializer { OrderTrackViewModel(orderId) }
        }
    }
}
