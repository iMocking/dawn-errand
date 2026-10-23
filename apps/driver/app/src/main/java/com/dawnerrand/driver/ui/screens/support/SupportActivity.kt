package com.dawnerrand.driver.ui.screens.support

import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.net.Uri
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.driver.R
import com.dawnerrand.driver.databinding.ActivitySupportBinding
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.dp
import com.dawnerrand.driver.ui.util.padTopWithStatusBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 骑手客服：热线呼叫（ACTION_DIAL）+ 在线客服入口 + 常见问题（展开/折叠）+ 意见反馈。
 */
class SupportActivity : AppCompatActivity() {

    private lateinit var binding: ActivitySupportBinding

    private data class Faq(val question: String, val answer: String)

    private val faqs = listOf(
        Faq("如何结算收入？", "收入按单日结：每笔配送完成后自动计入可提现余额，可在「我的-收入提现」中发起提现，到账至卡包绑定的银行卡。"),
        Faq("接单范围如何确定？", "系统默认在服务片区（绿杨乡全域）内派单，超过片区的订单会额外标注里程补贴，可联系站长调整服务范围。"),
        Faq("配送超时怎么办？", "遇到拥堵或异常请提前联系顾客说明，并在 App 内上报异常；超时判定以平台系统记录为准，非骑手原因不纳入考核。"),
        Faq("如何更换银行卡或手机号？", "银行卡可在「我的-卡包管理」中解绑旧卡、绑定新卡；更换手机号请联系热线客服协助完成实名核验后修改。"),
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivitySupportBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.navBar.padTopWithStatusBar()
        binding.btnBack.setOnClickListener { finish() }

        // 热线呼叫
        binding.cardHotline.setOnClickListener { callHotline() }

        // 在线客服
        binding.cardOnline.setOnClickListener {
            toast("在线客服接入中，请先拨打 400-000-0000")
        }

        buildFaqs()

        // 意见反馈
        binding.cardFeedback.setOnClickListener {
            toast("意见反馈已收到，感谢您的建议")
        }
    }

    private fun callHotline() {
        runCatching {
            val dial = Intent(Intent.ACTION_DIAL, Uri.parse("tel:4000000000"))
            startActivity(dial)
        }.onFailure {
            toast("无法拨号，请手动拨打 400-000-0000")
        }
    }

    private fun buildFaqs() {
        binding.faqContainer.removeAllViews()
        faqs.forEachIndexed { index, faq ->
            val header = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                gravity = Gravity.CENTER_VERTICAL
                isClickable = true
                isFocusable = true
                setPadding(dp(16), dp(14), dp(16), dp(14))
            }

            val title = TextView(this).apply {
                text = faq.question
                textSize = 14f
                setTextColor(Design.TEXT_PRIMARY)
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            }
            header.addView(title)

            val arrow = ImageView(this).apply {
                setImageResource(R.drawable.ic_chevron_down)
                setColorFilter(Design.TEXT_FAINT)
                layoutParams = LinearLayout.LayoutParams(dp(14), dp(14))
            }
            header.addView(arrow)

            val answer = TextView(this).apply {
                text = faq.answer
                textSize = 13f
                setTextColor(Design.TEXT_SECONDARY)
                setLineSpacing(0f, 1.5f)
                setPadding(dp(16), 0, dp(16), 0)
                visibility = View.GONE
                val lp = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.MATCH_PARENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                )
                lp.topMargin = dp(2)
                lp.bottomMargin = dp(14)
                layoutParams = lp
            }

            var expanded = false
            header.setOnClickListener {
                expanded = !expanded
                answer.visibility = if (expanded) View.VISIBLE else View.GONE
                arrow.rotation = if (expanded) 180f else 0f
            }

            binding.faqContainer.addView(header)
            binding.faqContainer.addView(answer)

            // 分隔线（最后一项不加）
            if (index < faqs.size - 1) {
                binding.faqContainer.addView(
                    View(this).apply {
                        setBackgroundColor(Design.DIVIDER)
                        layoutParams = LinearLayout.LayoutParams(
                            LinearLayout.LayoutParams.MATCH_PARENT,
                            1,
                        )
                    }
                )
            }
        }
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, SupportActivity::class.java))
        }
    }
}
