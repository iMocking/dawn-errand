package com.dawnerrand.store.ui.screens.orders

import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.DishStore
import com.dawnerrand.store.data.ORDER_COOKING
import com.dawnerrand.store.data.ORDER_DELIVERING
import com.dawnerrand.store.data.ORDER_DONE
import com.dawnerrand.store.data.ORDER_PENDING
import com.dawnerrand.store.data.StoreOrder
import com.dawnerrand.store.data.StoreRepository
import com.dawnerrand.store.data.StoreState

/**
 * 订单详情：顾客信息/商品明细/价格构成/订单信息（order-detail 稿）。
 * 任意订单都能从 "菜品×数量" 摘要还原明细与金额；配送中订单额外展示骑手信息。
 */
class OrderDetailViewModel(orderId: String) : ViewModel() {

    val order: StoreOrder? = StoreState.orderOf(orderId)

    val items: List<Triple<String, String, String?>> by lazy {
        StoreRepository.orderItems[orderId] ?: parseItems()
    }

    val riderName: String = "李师傅"
    val riderPhone: String = "138****6666"
    val riderStatus: String = "正在配送 · 预计 12:50 送达"

    val statusBig: String
        get() = when (order?.status) {
            ORDER_PENDING -> "待接单 · 等待处理"
            ORDER_COOKING -> "已接单 · 待制作"
            ORDER_DELIVERING -> "已出餐 · 配送中"
            else -> "订单已完成"
        }

    val statusSmall: String
        get() = when (order?.status) {
            ORDER_COOKING -> "接单时间 12:28 · 预计12:50前出餐"
            ORDER_PENDING -> "请尽快处理，顾客正在等待"
            ORDER_DELIVERING -> "骑手正在配送，请保持联系"
            else -> "感谢您的服务，期待再次合作"
        }

    val bottomPrimary: String
        get() = when (order?.status) {
            ORDER_PENDING -> "接单"
            ORDER_COOKING -> "通知骑手取餐"
            ORDER_DELIVERING -> "查看配送"
            else -> "订单已完成"
        }

    val bottomSecondary: String
        get() = when (order?.status) {
            ORDER_PENDING -> "拒单"
            else -> "联系顾客"
        }

    val isDelivering: Boolean
        get() = order?.status == ORDER_DELIVERING

    val isDone: Boolean
        get() = order?.status == ORDER_DONE

    /** 商品小计（明细之和） */
    val subTotal: String
        get() {
            val sum = items.sumOf { it.second.removePrefix("¥").toDoubleOrNull() ?: 0.0 }
            return "¥${format(sum)}"
        }

    private fun parseItems(): List<Triple<String, String, String?>> {
        val raw = order?.items.orEmpty()
        if (raw.isBlank()) return emptyList()
        return raw.split("、").mapNotNull { segment ->
            val match = Regex("^(.+?)\\s*×\\s*(\\d+)$").find(segment.trim()) ?: return@mapNotNull null
            val name = match.groupValues[1].trim()
            val qty = match.groupValues[2].toIntOrNull() ?: 1
            val unit = unitPrice(name)
            Triple("$name ×$qty", "¥${format(unit * qty)}", null)
        }
    }

    private fun unitPrice(name: String): Double {
        val dishes = DishStore.categoriesSnapshot().flatMap { it.dishes }
        val dish = dishes.firstOrNull { it.name == name }
            ?: dishes.firstOrNull { name.contains(it.name) || it.name.contains(name) }
        return dish?.price?.toDoubleOrNull() ?: 0.0
    }

    private fun format(v: Double): String = String.format("%.1f", v)
}
