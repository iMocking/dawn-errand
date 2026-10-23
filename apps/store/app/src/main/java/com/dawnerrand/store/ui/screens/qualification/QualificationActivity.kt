package com.dawnerrand.store.ui.screens.qualification

import android.os.Bundle
import android.view.Gravity
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.R
import com.dawnerrand.store.databinding.ActivityQualificationBinding
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.setLightStatusBars

/**
 * 资质证照：认证状态 + 营业执照 / 食品经营许可证 / 法人身份证（脱敏展示）。
 * 数据为演示资质，敏感字段已脱敏。
 */
class QualificationActivity : AppCompatActivity() {

    private lateinit var b: ActivityQualificationBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityQualificationBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.navBar.onBack = { finish() }

        addRows(
            b.licenseRows,
            listOf(
                "商户名称" to "村口老陈家常菜",
                "统一社会信用代码" to "91421125MA4D8X****",
                "经营者" to "陈建国",
                "有效期" to "2020-03-15 至 长期",
            )
        )
        addRows(
            b.foodRows,
            listOf(
                "许可证编号" to "JY14211250012345",
                "经营范围" to "预包装食品、热食类食品制售",
                "有效期" to "2024-01-01 至 2027-12-31",
            )
        )
        addRows(
            b.idRows,
            listOf(
                "姓名" to "陈建国",
                "身份证号" to "421125********6733",
                "有效期" to "2020-01-15 至 2040-01-15",
            )
        )
    }

    private fun addRows(container: LinearLayout, rows: List<Pair<String, String>>) {
        rows.forEachIndexed { index, (label, value) ->
            val row = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                gravity = Gravity.CENTER_VERTICAL
                setPadding(0, dp(9), 0, dp(9))
            }
            val l = TextView(this).apply {
                text = label
                textSize = 13f
                setTextColor(Design.TEXT_TERTIARY)
                layoutParams = LinearLayout.LayoutParams(dp(104), LinearLayout.LayoutParams.WRAP_CONTENT)
            }
            val v = TextView(this).apply {
                text = value
                textSize = 13f
                setTextColor(Design.TEXT_PRIMARY)
                gravity = Gravity.END
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            }
            row.addView(l)
            row.addView(v)
            container.addView(row)
            if (index != rows.lastIndex) {
                val divider = android.view.View(this).apply {
                    setBackgroundColor(Design.DIVIDER)
                    layoutParams = LinearLayout.LayoutParams(
                        LinearLayout.LayoutParams.MATCH_PARENT, dp(1)
                    ).apply { marginStart = dp(104) }
                }
                container.addView(divider)
            }
        }
    }
}
