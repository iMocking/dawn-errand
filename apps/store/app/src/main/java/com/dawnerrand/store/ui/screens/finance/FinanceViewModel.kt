package com.dawnerrand.store.ui.screens.finance

import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.StoreRepository

/**
 * 收入提现：余额卡片 + 提现按钮 + 近期收支记录（finance 稿）。
 */
class FinanceViewModel : ViewModel() {

    val balance = StoreRepository.balance
    val todayRevenue = StoreRepository.shop.todayRevenue
    val monthRevenue = StoreRepository.monthRevenue
    val monthOrders = StoreRepository.monthOrders
    val records = StoreRepository.financeRecords
}
