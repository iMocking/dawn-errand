package com.dawnerrand.client.ui.screens.coupons

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.Coupon
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.main.MainActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 优惠券：可用 / 已使用 / 已过期；下单页可作为选择器返回 couponId。 */
class CouponsActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this
    private var pickMode = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_coupons)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        pickMode = intent.getBooleanExtra(EXTRA_PICK, false)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        findViewById<TextView>(R.id.tvTitle).text = if (pickMode) "选择优惠券" else "优惠券"
        findViewById<TextView>(R.id.tvSummary).text =
            "共 ${DemoRepository.coupons.size} 张 · 可用 ${DemoRepository.availableCouponCount()} 张 · 优惠券不可叠加使用"
        findViewById<View>(R.id.btnNoCoupon).visibility = if (pickMode) View.VISIBLE else View.GONE
        findViewById<View>(R.id.btnNoCoupon).setOnClickListener {
            setResult(Activity.RESULT_OK, Intent())
            finish()
        }
        render()
    }

    private fun render() {
        val list = findViewById<LinearLayout>(R.id.couponList)
        list.removeAllViews()
        val ordered = DemoRepository.coupons.sortedByDescending { it.usable }
        if (ordered.isEmpty()) {
            list.addView(TextView(this).apply {
                text = "暂无优惠券，去首页逛逛吧"
                setTextColor(Design.TEXT_FAINT)
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(0, ctx.dp(80), 0, ctx.dp(80))
            })
            return
        }
        ordered.forEach { list.addView(buildCard(it)) }
    }

    private fun buildCard(c: Coupon): LinearLayout {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            background = ContextCompat.getDrawable(
                this@CouponsActivity,
                if (c.usable) R.drawable.bg_card_14 else R.drawable.bg_card_gray,
            )
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(92))
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.bottomMargin = ctx.dp(10)
            }
            isClickable = true
            setOnClickListener { onCouponClick(c) }
        }
        // 左面额区
        val left = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(ctx.dp(96), LinearLayout.LayoutParams.MATCH_PARENT)
            background = getDrawable(
                if (c.usable) R.drawable.bg_coupon else R.drawable.bg_coupon_gray)
        }
        left.addView(TextView(this).apply {
            text = "¥${c.amount}"
            setTextColor(if (c.usable) Design.WHITE else 0xFFC9CDD4.toInt())
            textSize = 24f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        left.addView(TextView(this).apply {
            text = c.cond
            setTextColor(if (c.usable) Design.WHITE else 0xFFC9CDD4.toInt())
            textSize = 11f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        card.addView(left)

        // 中间信息区
        val right = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(ctx.dp(14), ctx.dp(10), ctx.dp(8), ctx.dp(10))
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.MATCH_PARENT, 1f)
        }
        val titleRow = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
        }
        titleRow.addView(TextView(this).apply {
            text = c.title
            setTextColor(if (c.usable) Design.TEXT_PRIMARY else 0xFF86909C.toInt())
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        })
        if (c.used || c.disabled) {
            titleRow.addView(TextView(this).apply {
                text = if (c.used) "已使用" else "已过期"
                setTextColor(0xFFC9CDD4.toInt())
                textSize = 11f
            })
        }
        right.addView(titleRow)
        right.addView(TextView(this).apply {
            text = c.desc
            setTextColor(if (c.usable) Design.TEXT_TERTIARY else 0xFFC9CDD4.toInt())
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
            maxLines = 1
            ellipsize = android.text.TextUtils.TruncateAt.END
        })
        right.addView(TextView(this).apply {
            text = c.expire
            setTextColor(if (c.usable) Design.TEXT_FAINT else 0xFFC9CDD4.toInt())
            textSize = 11f
            setPadding(0, ctx.dp(3), 0, 0)
        })
        card.addView(right)

        // 右侧操作
        val action = TextView(this).apply {
            text = when {
                c.used -> "已使用"
                c.disabled -> "已过期"
                pickMode -> "选择"
                else -> "去使用"
            }
            textSize = 12f
            gravity = Gravity.CENTER
            setPadding(ctx.dp(14), ctx.dp(7), ctx.dp(14), ctx.dp(7))
            setTextColor(if (c.usable) Design.PRIMARY else 0xFFC9CDD4.toInt())
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(12)
            if (c.usable) {
                background = getDrawable(R.drawable.bg_btn_outline)
                isClickable = true
                setOnClickListener { onCouponClick(c) }
            }
        }
        card.addView(action)
        return card
    }

    private fun onCouponClick(c: Coupon) {
        if (!c.usable) {
            ctx.toast(if (c.used) "该优惠券已使用" else "该优惠券已过期")
            return
        }
        if (pickMode) {
            setResult(Activity.RESULT_OK, Intent().putExtra(EXTRA_SELECTED_COUPON_ID, c.id))
            finish()
            return
        }
        AlertDialog.Builder(this)
            .setTitle("优惠券详情")
            .setMessage(
                "${c.title}\n\n面额：¥${c.amount}\n使用条件：${c.cond}\n适用：${c.desc}\n${c.expire}\n\n点击“去使用”返回首页选购商品，结算时自动可选用。"
            )
            .setNegativeButton("关闭", null)
            .setPositiveButton("去使用") { _, _ ->
                startActivity(Intent(this, MainActivity::class.java).apply {
                    flags = Intent.FLAG_ACTIVITY_CLEAR_TOP or Intent.FLAG_ACTIVITY_SINGLE_TOP
                })
                finish()
            }
            .show()
    }

    companion object {
        const val EXTRA_PICK = "pick_coupon"
        const val EXTRA_SELECTED_COUPON_ID = "selected_coupon_id"
    }
}
