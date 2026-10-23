package com.dawnerrand.store.ui.screens.statistics

import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.StoreRepository

/**
 * 经营数据：时间范围切换 + 概览 + 近7天营业额趋势 + 热销 TOP5（statistics 稿）。
 */
class StatisticsViewModel : ViewModel() {

    var rangeIndex = 0

    val ranges = listOf("今日", "近7天", "近30天", "本月")

    val overview = StoreRepository.statOverview
    val trend = StoreRepository.statTrend
    val ranks = StoreRepository.statRanks
}
