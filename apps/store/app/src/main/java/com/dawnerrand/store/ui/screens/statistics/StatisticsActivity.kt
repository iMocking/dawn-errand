package com.dawnerrand.store.ui.screens.statistics

import android.graphics.Typeface
import android.os.Bundle
import android.view.Gravity
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.databinding.ActivityStatisticsBinding
import com.dawnerrand.store.ui.adapter.RankAdapter
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.setLightStatusBars

/**
 * 经营数据：时间范围 + 概览 + 营业额趋势柱状图 + 热销 TOP5（statistics 稿）。
 */
class StatisticsActivity : AppCompatActivity() {

    private lateinit var b: ActivityStatisticsBinding
    private val vm: StatisticsViewModel by viewModels()

    private val rankAdapter = RankAdapter()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityStatisticsBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.backBtn.setOnClickListener { finish() }

        // 排行
        b.rankList.layoutManager = LinearLayoutManager(this)
        b.rankList.adapter = rankAdapter
        rankAdapter.submit(vm.ranks)

        // 范围 Tab
        buildRangeTabs()
        applyStats()
    }

    private fun buildRangeTabs() {
        b.rangeTabs.removeAllViews()
        vm.ranges.forEachIndexed { index, range ->
            val tab = TextView(this).apply {
                text = range
                textSize = 13f
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(0, dp(32), 1f)
                isClickable = true
                setOnClickListener {
                    vm.select(index)
                    refreshRangeTabs()
                    applyStats()
                }
            }
            b.rangeTabs.addView(tab)
        }
        refreshRangeTabs()
    }

    private fun refreshRangeTabs() {
        val current = vm.rangeIndex
        for (i in 0 until b.rangeTabs.childCount) {
            val tab = b.rangeTabs.getChildAt(i) as TextView
            val active = i == current
            tab.setBackgroundResource(
                if (active) com.dawnerrand.store.R.drawable.bg_range_active
                else com.dawnerrand.store.R.drawable.bg_card_white
            )
            tab.setTextColor(if (active) Design.WHITE else Design.TEXT_SECONDARY)
            tab.setTypeface(tab.typeface, if (active) Typeface.BOLD else Typeface.NORMAL)
        }
    }

    private fun applyStats() {
        val stats = vm.current
        b.revenueValue.text = stats.overview.revenue
        b.ordersValue.text = stats.overview.orders
        b.avgValue.text = stats.overview.avgOrder
        b.chartTitle.text = stats.title
        b.barChart.setData(stats.trend.labels, stats.trend.values)
    }
}
