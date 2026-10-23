package com.dawnerrand.store.ui.screens.statistics

import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.StatOverview
import com.dawnerrand.store.data.StatTrend

/**
 * 经营数据：时间范围切换 + 概览 + 营业额趋势 + 热销 TOP5（statistics 稿）。
 * 四个时间范围对应不同概览与趋势数据（演示数据）。
 */
class StatisticsViewModel : ViewModel() {

    data class RangeStats(
        val title: String,
        val overview: StatOverview,
        val trend: StatTrend,
    )

    val ranges = listOf("今日", "近7天", "近30天", "本月")

    private val data = listOf(
        RangeStats(
            "近7天营业额趋势（元）",
            StatOverview("¥1,286", "42", "¥30.6"),
            StatTrend(
                labels = listOf("周三", "周四", "周五", "周六", "周日", "周一", "今天"),
                values = listOf(0.50f, 0.65f, 0.45f, 0.85f, 0.95f, 0.55f, 0.70f),
            ),
        ),
        RangeStats(
            "近7天营业额趋势（元）",
            StatOverview("¥7,842", "268", "¥29.3"),
            StatTrend(
                labels = listOf("周三", "周四", "周五", "周六", "周日", "周一", "今天"),
                values = listOf(0.62f, 0.70f, 0.55f, 0.90f, 0.98f, 0.66f, 0.78f),
            ),
        ),
        RangeStats(
            "近30天营业额趋势（元）",
            StatOverview("¥32,860", "1,126", "¥29.2"),
            StatTrend(
                labels = listOf("第1周", "第2周", "第3周", "第4周", "第5周", "本周", "今天"),
                values = listOf(0.55f, 0.60f, 0.72f, 0.80f, 0.88f, 0.76f, 0.92f),
            ),
        ),
        RangeStats(
            "本月营业额趋势（元）",
            StatOverview("¥26,140", "896", "¥29.2"),
            StatTrend(
                labels = listOf("1日", "5日", "10日", "15日", "20日", "25日", "今天"),
                values = listOf(0.48f, 0.66f, 0.58f, 0.80f, 0.90f, 0.72f, 0.84f),
            ),
        ),
    )

    var rangeIndex = 0
        private set

    val current: RangeStats
        get() = data.getOrElse(rangeIndex) { data[0] }

    val ranks = com.dawnerrand.store.data.StoreRepository.statRanks

    fun select(index: Int) {
        if (index in data.indices) rangeIndex = index
    }
}
