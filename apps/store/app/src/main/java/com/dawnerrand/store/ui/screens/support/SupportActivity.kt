package com.dawnerrand.store.ui.screens.support

import android.os.Bundle
import android.view.Gravity
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.R
import com.dawnerrand.store.databinding.ActivitySupportBinding
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dial
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast

/**
 * 商家客服：服务热线（拨号）+ 在线客服 + 常见问题展开 + 意见反馈。
 */
class SupportActivity : AppCompatActivity() {

    private lateinit var b: ActivitySupportBinding

    private val faqs = listOf(
        "订单结算与提现什么时候到账？" to
            "订单完成后收入实时计入可提现余额；提现申请提交后 1 个工作日内到账，遇节假日顺延。",
        "如何修改营业时间或设置自动接单？" to
            "在「工作台 - 营业设置」中可修改营业时间、起送价、配送范围、配送费与自动接单开关，保存后立即生效。",
        "顾客申请退款怎么处理？" to
            "请在订单详情查看顾客备注并及时联系顾客协商；如顾客已申请退款，平台客服会与您确认后处理。",
        "菜品售罄或缺货怎么办？" to
            "在「菜品管理」中关闭对应菜品的上架开关即可，售罄状态会同步到顾客端。",
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivitySupportBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.navBar.onBack = { finish() }

        b.hotlineCard.setOnClickListener { dial("4000000000") }
        b.callBtn.setOnClickListener { dial("4000000000") }
        b.onlineRow.setOnClickListener {
            toast("在线客服接入中，请拨打商家服务热线 400-000-0000")
        }
        b.feedbackRow.setOnClickListener {
            toast("意见反馈已收到，感谢您的建议")
        }

        buildFaqs()
    }

    private fun buildFaqs() {
        b.faqContainer.removeAllViews()
        faqs.forEachIndexed { index, (question, answer) ->
            val item = LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
            }
            val header = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                gravity = Gravity.CENTER_VERTICAL
                setPadding(0, dp(12), 0, dp(12))
                isClickable = true
            }
            val q = TextView(this).apply {
                text = question
                textSize = 13f
                setTextColor(Design.TEXT_PRIMARY)
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            }
            val arrow = ImageView(this).apply {
                setImageResource(R.drawable.ic_chevron_right)
                setColorFilter(Design.TEXT_FAINT)
                rotation = 90f
                layoutParams = LinearLayout.LayoutParams(dp(14), dp(14))
            }
            val a = TextView(this).apply {
                text = answer
                textSize = 12f
                setTextColor(Design.TEXT_SECONDARY)
                setLineSpacing(dp(3).toFloat(), 1f)
                setPadding(0, 0, dp(20), dp(12))
                visibility = android.view.View.GONE
            }
            header.addView(q)
            header.addView(arrow)
            header.setOnClickListener {
                val opening = a.visibility != android.view.View.VISIBLE
                a.visibility = if (opening) android.view.View.VISIBLE else android.view.View.GONE
                arrow.animate().rotation(if (opening) 270f else 90f).setDuration(160).start()
            }
            item.addView(header)
            item.addView(a)
            b.faqContainer.addView(item)
            if (index != faqs.lastIndex) {
                val divider = android.view.View(this).apply {
                    setBackgroundColor(Design.DIVIDER)
                    layoutParams = LinearLayout.LayoutParams(LinearLayout.LayoutParams.MATCH_PARENT, dp(1))
                }
                b.faqContainer.addView(divider)
            }
        }
    }
}
