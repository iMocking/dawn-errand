package com.dawnerrand.client.ui.screens.help

import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import androidx.core.widget.doAfterTextChanged
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.screens.customer.CustomerServiceActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars

/** 帮助中心：问题搜索 + 分类筛选 + FAQ 展开收起 + 客服入口。 */
class HelpCenterActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this
    private val faqList = mutableListOf<LinearLayout>()
    private var expanded = -1
    private var category = "全部"
    private var keyword = ""

    private data class Faq(val category: String, val q: String, val a: String)

    private val faqs = listOf(
        Faq("下单与支付", "如何下单？",
            "在首页选择“点个饭”或“跑个腿”，选择商家与商品/填写跑腿需求后点击结算，确认地址、优惠券与支付方式，提交订单并完成支付即可。"),
        Faq("下单与支付", "支持哪些支付方式？",
            "目前支持微信支付、支付宝与账户余额支付。演示版本统一按支付成功处理，正式接入支付渠道后以实际扣款结果为准。"),
        Faq("下单与支付", "下单后可以修改地址吗？",
            "商家接单前可在订单详情中取消后重新下单；骑手取货后请直接联系骑手或客服协商，平台会尽力协助处理。"),
        Faq("配送与跑腿", "配送需要多久？",
            "页面展示为预计时间，通常为 15-35 分钟。恶劣天气、用餐高峰或乡镇偏远地址可能略有延迟。"),
        Faq("配送与跑腿", "跑腿费怎么计算？",
            "跑腿费根据取送距离、物品重量与时段综合计算。商品垫付费用按实际采购金额结算，多退少补。"),
        Faq("配送与跑腿", "如何联系骑手？",
            "配送中的订单可在“订单-查看配送”页面点击“呼叫”联系骑手；未接单时可联系客服协助催单。"),
        Faq("售后与退款", "如何申请退款？",
            "在订单详情中点击“申请售后”，选择退款原因并提交。商家或平台审核通过后，款项将原路退回。"),
        Faq("售后与退款", "订单取消规则？",
            "商家接单前可免费取消；接单后取消可能产生已发生的配送成本。跑腿订单骑手取货后不支持无理由取消。"),
        Faq("售后与退款", "商品有问题怎么办？",
            "请保留商品照片并在订单中申请售后，或联系客服 400-888-0000，我们会优先为您处理。"),
        Faq("账号与安全", "如何修改昵称？",
            "进入“我的-设置-编辑资料”即可修改昵称；手机号用于登录与订单联系，暂不支持自助修改。"),
        Faq("账号与安全", "收不到验证码怎么办？",
            "请检查手机号是否正确、信号是否正常，60 秒后可重新获取。演示环境会自动填充验证码。"),
        Faq("账号与安全", "如何退出登录？",
            "进入“我的-设置”，点击底部“退出登录”并确认即可。退出后本机演示数据仍会保留。"),
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_help_center)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        findViewById<View>(R.id.btnContact).setOnClickListener {
            startActivity(Intent(this, CustomerServiceActivity::class.java))
        }

        renderChips()
        renderFaqs()
        findViewById<EditText>(R.id.etSearch).doAfterTextChanged {
            keyword = it?.toString()?.trim().orEmpty()
            expanded = -1
            renderFaqs()
        }
    }

    private fun renderChips() {
        val row = findViewById<LinearLayout>(R.id.chipRow)
        row.removeAllViews()
        val categories = listOf("全部") + faqs.map { it.category }.distinct()
        categories.forEach { name ->
            val active = name == category
            row.addView(TextView(this).apply {
                text = name
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(ctx.dp(14), ctx.dp(7), ctx.dp(14), ctx.dp(7))
                setTextColor(if (active) Design.WHITE else Design.TEXT_SECONDARY)
                background = ContextCompat.getDrawable(
                    this@HelpCenterActivity,
                    if (active) R.drawable.bg_chip_active else R.drawable.bg_chip,
                )
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
                (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(8)
                setOnClickListener {
                    category = name
                    expanded = -1
                    renderChips()
                    renderFaqs()
                }
            })
        }
    }

    private fun renderFaqs() {
        val list = findViewById<LinearLayout>(R.id.faqList)
        list.removeAllViews()
        faqList.clear()
        val shown = faqs.filter {
            (category == "全部" || it.category == category) &&
                (keyword.isBlank() || it.q.contains(keyword, true) || it.a.contains(keyword, true))
        }
        if (shown.isEmpty()) {
            list.addView(TextView(this).apply {
                text = "没有找到相关问题，换个关键词试试"
                setTextColor(Design.TEXT_FAINT)
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(0, ctx.dp(60), 0, ctx.dp(60))
            })
            return
        }
        var lastCategory = ""
        shown.forEach { faq ->
            if (faq.category != lastCategory) {
                lastCategory = faq.category
                list.addView(TextView(this).apply {
                    text = faq.category
                    setTextColor(Design.TEXT_SECONDARY)
                    textSize = 12f
                    setPadding(ctx.dp(4), ctx.dp(10), 0, ctx.dp(8))
                })
            }
            list.addView(buildFaqCard(faq))
        }
    }

    private fun buildFaqCard(faq: Faq): LinearLayout {
        val index = faqs.indexOf(faq)
        val isOpen = expanded == index
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            background = ContextCompat.getDrawable(this@HelpCenterActivity, R.drawable.bg_card)
            setPadding(ctx.dp(14), ctx.dp(13), ctx.dp(14), ctx.dp(13))
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).bottomMargin = ctx.dp(10)
            isClickable = true
            setOnClickListener {
                expanded = if (isOpen) -1 else index
                renderFaqs()
            }
        }
        val head = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
        }
        head.addView(TextView(this).apply {
            text = faq.q
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 14f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        })
        head.addView(ImageView(this).apply {
            setImageResource(R.drawable.ic_chevron_down)
            setColorFilter(Design.TEXT_FAINT)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(16), ctx.dp(16))
            rotation = if (isOpen) 180f else 0f
        })
        card.addView(head)
        if (isOpen) {
            card.addView(TextView(this).apply {
                text = faq.a
                setTextColor(Design.TEXT_SECONDARY)
                textSize = 13f
                setLineSpacing(ctx.dp(5).toFloat(), 1f)
                setPadding(0, ctx.dp(10), 0, ctx.dp(2))
            })
        }
        faqList.add(card)
        return card
    }
}
