package com.dawnerrand.store.ui.screens.statistics

import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.databinding.ActivityStatisticsBinding
import com.dawnerrand.store.ui.adapter.RankAdapter
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp

/**
 * 经营数据：时间范围 + 概览 + 近7天营业额柱状图 + 热销 TOP5（statistics 稿）。
 */
class StatisticsActivity : AppCompatActivity() {

    private lateinit var b: ActivityStatisticsBinding
    private val vm: StatisticsViewModel by viewModels()

    private val rankAdapter = RankAdapter()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityStatisticsBinding.inflate(layoutInflater)
        setContentView(b.root)

        b.backBtn.setOnClickListener { finish() }

        // 概览
        b.revenueValue.text = vm.overview.revenue
        b.ordersValue.text = vm.overview.orders
        b.avgValue.text = vm.overview.avgOrder

        // 排行
        b.rankList.layoutManager = LinearLayoutManager(this)
        b.rankList.adapter = rankAdapter
        rankAdapter.submit(vm.ranks)

        // 范围 Tab
        buildRangeTabs()

        // 柱状图
        buildBarChart()
    }

    private fun buildRangeTabs() {
        b.rangeTabs.removeAllViews()
        vm.ranges.forEachIndexed { index, range ->
            val tab = TextView(this).apply {
                text = range
                textSize = 13f
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(
                    0,
                    dp(32),
                    1f
                )
                isClickable = true
                setOnClickListener {
                    vm.rangeIndex = index
                    refreshRangeTabs()
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
            tab.setTypeface(
                tab.typeface,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
        }
    }

    private fun buildBarChart() {
        b.barChart.removeAllViews()
        val max = vm.trend.values.maxOrNull()?.takeIf { it > 0f } ?: 1f
        vm.trend.labels.zip(vm.trend.values).forEach { (label, value) ->
            val col = LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER_HORIZONTAL
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.MATCH_PARENT, 1f)
            }
            val bar = View(this).apply {
                setBackgroundResource(com.dawnerrand.store.R.drawable.bg_bar_grad)
                val h = dp((value / max * 100).toInt().coerceIn(8, 100))
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.MATCH_PARENT,
                    h
                )
            }
            val text = TextView(this).apply {
                text = label
                textSize = 10f
                setTextColor(Design.TEXT_TERTIARY)
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT
                ).apply { topMargin = dp(4) }
            }
            col.addView(bar, 0)
            col.addView(text)
            b.barChart.addView(col)
        }
    }
}
