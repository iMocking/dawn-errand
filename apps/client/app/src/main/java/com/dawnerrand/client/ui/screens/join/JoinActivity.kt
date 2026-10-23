package com.dawnerrand.client.ui.screens.join

import android.graphics.Color
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/** 商家入驻 / 成为骑手：资料表单 + 审核状态 + 本地持久化。 */
class JoinActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private var mode = TYPE_MERCHANT
    private var selectedType = ""
    private var agreed = true
    private val typeChips = mutableListOf<TextView>()
    private lateinit var options: List<String>

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_join)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        mode = intent.getStringExtra(EXTRA_TYPE) ?: TYPE_MERCHANT
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        findViewById<TextView>(R.id.etName).setText(DemoRepository.currentUser.name)
        findViewById<TextView>(R.id.etPhone).setText(DemoRepository.currentUser.phone)

        options = if (mode == TYPE_MERCHANT) {
            listOf("餐饮", "超市", "水果", "卤味", "其他")
        } else {
            listOf("电动车", "摩托车", "自行车", "步行")
        }
        selectedType = options.first()
        configure()
        renderTypeChips()
        renderStatus()
        val chk = findViewById<View>(R.id.chkAgree)
        chk.setOnClickListener {
            agreed = !agreed
            (chk as TextView).setBackgroundResource(
                if (agreed) R.drawable.bg_agreed else R.drawable.bg_agreed_off)
            (chk as TextView).text = if (agreed) "✓" else ""
        }
        findViewById<View>(R.id.btnSubmit).setOnClickListener { submit() }
    }

    private fun configure() {
        val merchant = mode == TYPE_MERCHANT
        findViewById<TextView>(R.id.tvNavTitle).text = if (merchant) "商家入驻" else "成为骑手"
        findViewById<TextView>(R.id.tvHeroTitle).text =
            if (merchant) "把乡镇好生意，搬到线上" else "空闲时间，跑出额外收入"
        findViewById<TextView>(R.id.tvHeroSub).text =
            if (merchant) "0 元入驻 · 专人运营指导 · 次日结款"
            else "按单结算 · 时间自由 · 乡镇就近派单"
        findViewById<TextView>(R.id.tvNameLabel).text = if (merchant) "联系人" else "姓名"
        findViewById<TextView>(R.id.tvDetailLabel).text = if (merchant) "店铺地址" else "服务区域"
        findViewById<TextView>(R.id.tvTypeLabel).text = if (merchant) "经营类型" else "配送工具"
        findViewById<TextView>(R.id.tvNoteTitle).text = if (merchant) "入驻说明" else "骑手说明"
        findViewById<TextView>(R.id.tvNote).text =
            if (merchant) {
                "入驻流程：提交申请 → 平台审核（1-3 个工作日）→ 开通商家工作台。\n" +
                    "审核通过后可在商家端管理菜品、接收订单与查看结算；平台不收取入驻费。"
            } else {
                "骑手流程：提交申请 → 身份审核 → 培训与开通接单。\n" +
                    "审核通过后可在骑手端上线接单，按单结算，平台提供意外险与就近派单。"
            }
        findViewById<EditText>(R.id.etDetail).hint =
            if (merchant) "请输入店铺详细地址" else "请输入常用服务区域，如：王家湾村"
        findViewById<TextView>(R.id.btnSubmit).text =
            if (merchant) "提交入驻申请" else "提交骑手申请"

        val stats = if (merchant) {
            listOf("0元" to "入驻费", "1-3天" to "审核时效", "T+1" to "结算周期")
        } else {
            listOf("30+" to "日均单量", "按单" to "结算方式", "就近" to "派单范围")
        }
        val row = findViewById<LinearLayout>(R.id.heroStats)
        stats.forEach { (num, label) ->
            row.addView(LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
                addView(TextView(this@JoinActivity).apply {
                    text = num
                    setTextColor(Design.WHITE)
                    textSize = 17f
                    typeface = android.graphics.Typeface.DEFAULT_BOLD
                })
                addView(TextView(this@JoinActivity).apply {
                    text = label
                    setTextColor(0xCCFFFFFF.toInt())
                    textSize = 11f
                    setPadding(0, ctx.dp(3), 0, 0)
                })
            })
        }
    }

    private fun renderTypeChips() {
        val row = findViewById<LinearLayout>(R.id.typeRow)
        row.removeAllViews()
        typeChips.clear()
        options.forEach { option ->
            val chip = TextView(this).apply {
                text = option
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(ctx.dp(14), ctx.dp(8), ctx.dp(14), ctx.dp(8))
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
                (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(8)
                isClickable = true
                setOnClickListener {
                    selectedType = option
                    renderTypeChips()
                }
            }
            typeChips.add(chip)
            row.addView(chip)
        }
        typeChips.forEach { chip ->
            val active = chip.text.toString() == selectedType
            chip.setTextColor(if (active) Design.WHITE else Design.TEXT_SECONDARY)
            chip.setBackgroundResource(if (active) R.drawable.bg_chip_active else R.drawable.bg_chip)
        }
    }

    private fun renderStatus() {
        val card = findViewById<View>(R.id.cardStatus)
        val app = DemoRepository.applicationOf(mode) ?: run {
            card.visibility = View.GONE
            return
        }
        card.visibility = View.VISIBLE
        findViewById<TextView>(R.id.tvStatus).text = app.status
        findViewById<TextView>(R.id.tvStatusTime).text =
            "提交于 " + SimpleDateFormat("MM-dd HH:mm", Locale.CHINA).format(Date(app.submittedAt))
        findViewById<TextView>(R.id.tvStatusSummary).text = app.summary
        findViewById<TextView>(R.id.tvStatusNote).text =
            if (mode == TYPE_MERCHANT) {
                "预计 1-3 个工作日完成审核，请保持电话畅通。"
            } else {
                "预计 1-2 个工作日完成身份审核，审核通过后可上线接单。"
            }
        val tag = findViewById<TextView>(R.id.tvStatus)
        tag.background = ContextCompat.getDrawable(this, R.drawable.bg_tag_active)
        tag.setTextColor(Color.WHITE)
        findViewById<TextView>(R.id.btnSubmit).text =
            if (mode == TYPE_MERCHANT) "更新入驻申请" else "更新骑手申请"
    }

    private fun submit() {
        val name = findViewById<EditText>(R.id.etName).text.toString().trim()
        val phone = findViewById<EditText>(R.id.etPhone).text.toString().trim()
        val detail = findViewById<EditText>(R.id.etDetail).text.toString().trim()
        when {
            name.isEmpty() -> ctx.toast(if (mode == TYPE_MERCHANT) "请填写联系人" else "请填写姓名")
            phone.length != 11 || !phone.startsWith("1") -> ctx.toast("请填写正确的11位手机号")
            detail.isEmpty() -> ctx.toast(
                if (mode == TYPE_MERCHANT) "请填写店铺地址" else "请填写服务区域")
            !agreed -> ctx.toast("请先阅读并同意服务协议")
            else -> {
                DemoRepository.submitApplication(mode, "$selectedType · $detail")
                renderStatus()
                ctx.toast(if (mode == TYPE_MERCHANT) "入驻申请已提交，等待平台审核" else "骑手申请已提交，等待平台审核")
            }
        }
    }

    companion object {
        const val EXTRA_TYPE = "join_type"
        const val TYPE_MERCHANT = "merchant"
        const val TYPE_RIDER = "rider"
    }
}
