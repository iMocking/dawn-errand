package com.dawnerrand.store.ui.screens.finance

import android.content.Context
import androidx.lifecycle.LiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.FinanceRecord
import com.dawnerrand.store.data.FinanceStore
import com.dawnerrand.store.data.StoreRepository
import com.dawnerrand.store.data.SettlementStore

/**
 * 收入提现：余额卡片 + 提现按钮 + 近期收支记录（finance 稿）。
 * 余额与记录来自 FinanceStore，提现后实时扣减并追加流水。
 */
class FinanceViewModel : ViewModel() {

    val balance: LiveData<Double> = FinanceStore.balance
    val records: LiveData<List<FinanceRecord>> = FinanceStore.records
    val todayRevenue: String
        get() = StoreRepository.shop.todayRevenue
    val monthRevenue = StoreRepository.monthRevenue
    val monthOrders = StoreRepository.monthOrders

    fun balanceText(): String = "¥${FinanceStore.formatMoney(balance.value ?: 0.0)}"

    fun targetText(ctx: Context): String = SettlementStore.withdrawTarget(ctx)

    fun feeTip(): String = "提现手续费 ${FinanceStore.feeRateText()} · 预计 1 个工作日内到账"

    fun withdraw(ctx: Context, amount: String): FinanceStore.WithdrawResult =
        FinanceStore.withdraw(ctx, amount)
}
