package com.dawnerrand.driver.ui.screens.checkin

import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.view.Gravity
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.driver.R
import com.dawnerrand.driver.databinding.ActivityCheckinBinding
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.dp
import com.dawnerrand.driver.ui.util.padTopWithStatusBar

/**
 * 每日签到：还原 design/driver/checkin.html（Hero + 连续天数卡 + 本周签到日历 + 奖励规则）。
 */
class CheckinActivity : AppCompatActivity() {

    private lateinit var binding: ActivityCheckinBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityCheckinBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.navBar.padTopWithStatusBar()
        binding.btnBack.setOnClickListener { finish() }
        binding.btnCheckin.setOnClickListener {
            if (binding.btnCheckin.tag == "done") {
                // 今日已签到
            }
        }
        binding.btnCheckin.tag = "done"

        buildCalendar()
    }

    private fun buildCalendar() {
        val days = listOf("一", "二", "三", "四", "五", "六", "日")
        binding.calGrid.removeAllViews()
        days.forEachIndexed { index, day ->
            val cell = LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(0, dp(44), 1f)
                if (index < days.size - 1) {
                    (layoutParams as LinearLayout.LayoutParams).marginEnd = dp(8)
                }
            }
            val isToday = index == days.size - 1
            val isChecked = index < days.size - 1

            cell.setBackgroundResource(
                when {
                    isToday -> R.drawable.bg_cal_today
                    isChecked -> R.drawable.bg_cal_checked
                    else -> android.R.color.transparent
                }
            )

            val dayTv = TextView(this).apply {
                text = day
                textSize = 12f
                setTextColor(if (isToday) Design.WHITE else if (isChecked) Design.PRIMARY else Design.TEXT_FAINT)
                if (isToday) setTypeface(typeface, android.graphics.Typeface.BOLD)
            }
            val markTv = TextView(this).apply {
                text = if (isToday) "今" else if (isChecked) "✓" else ""
                textSize = 10f
                setTextColor(if (isToday) Design.WHITE else Design.PRIMARY)
                if (isToday) setTypeface(typeface, android.graphics.Typeface.BOLD)
            }
            cell.addView(dayTv)
            cell.addView(markTv)
            binding.calGrid.addView(cell)
        }
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, CheckinActivity::class.java))
        }
    }
}
