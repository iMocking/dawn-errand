package com.dawnerrand.store.ui.screens.orders

import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.StoreOrder
import com.dawnerrand.store.data.StoreRepository
import com.dawnerrand.store.data.StoreState

/**
 * 订单详情：顾客信息/商品明细/价格构成/订单信息（order-detail 稿）。
 */
class OrderDetailViewModel(orderId: String) : ViewModel() {

    val order: StoreOrder? = StoreState.orderOf(orderId)
    val items: List<Triple<String, String, String?>> =
        StoreRepository.orderItems[orderId] ?: emptyList()

    val statusBig: String
        get() = when (order?.status) {
            com.dawnerrand.store.data.ORDER_PENDING -> "待接单 · 等待处理"
            com.dawnerrand.store.data.ORDER_COOKING -> "已接单 · 待制作"
            com.dawnerrand.store.data.ORDER_DELIVERING -> "已出餐 · 配送中"
            else -> "订单已完成"
        }

    val statusSmall: String
        get() = when (order?.status) {
            com.dawnerrand.store.data.ORDER_COOKING -> "接单时间 12:28 · 预计12:50前出餐"
            com.dawnerrand.store.data.ORDER_PENDING -> "请尽快处理，顾客正在等待"
            com.dawnerrand.store.data.ORDER_DELIVERING -> "骑手正在配送，请保持联系"
            else -> "感谢您的服务"
        }

    val bottomPrimary: String
        get() = when (order?.status) {
            com.dawnerrand.store.data.ORDER_PENDING -> "接单"
            com.dawnerrand.store.data.ORDER_COOKING -> "通知骑手取餐"
            else -> "完成"
        }

    val bottomSecondary: String
        get() = when (order?.status) {
            com.dawnerrand.store.data.ORDER_PENDING -> "拒单"
            else -> "联系顾客"
        }

    /** 商品小计（明细之和） */
    val subTotal: String
        get() {
            val sum = items.sumOf { it.second.removePrefix("¥").toDouble() }
            return "¥${format(sum)}"
        }

    private fun format(v: Double): String =
        if (v % 1.0 == 0.0) v.toInt().toString() else String.format("%.1f", v)
}
