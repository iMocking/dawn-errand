package com.dawnerrand.store.data

import androidx.lifecycle.MutableLiveData

/**
 * 共享订单状态：工作台角标与订单列表来自同一数据源，接单/拒单/出餐后各页联动刷新。
 */
object StoreState {

    val ordersLive = MutableLiveData<List<StoreOrder>>()

    init {
        ordersLive.postValue(StoreRepository.orders)
    }

    fun orderOf(id: String): StoreOrder? = ordersLive.value?.find { it.id == id }

    /** 接单：待接单 → 制作中 */
    fun accept(id: String) {
        update(id) { it.copy(status = ORDER_COOKING, actions = listOf("通知出餐")) }
    }

    /** 拒单：移除该单 */
    fun reject(id: String) {
        val list = ordersLive.value?.filterNot { it.id == id } ?: return
        ordersLive.value = list
    }

    /** 出餐：制作中 → 配送中 */
    fun notifyReady(id: String) {
        update(id) { it.copy(status = ORDER_DELIVERING, actions = listOf("查看配送")) }
    }

    private fun update(id: String, transform: (StoreOrder) -> StoreOrder) {
        val list = ordersLive.value ?: return
        ordersLive.value = list.map { if (it.id == id) transform(it) else it }
    }
}
