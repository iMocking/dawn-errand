package com.dawnerrand.client.ui.screens.coupons

import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.data.Coupon
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 优惠券：左橙渐变面额 + 右信息 + 去使用/已过期 */
class CouponsActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_coupons)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        val list = findViewById<LinearLayout>(R.id.couponList)
        DemoRepository.coupons.forEach { list.addView(buildCard(it)) }
    }

    private fun buildCard(c: Coupon): LinearLayout {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            background = android.graphics.drawable.ColorDrawable(0xFFE5E6EB.toInt())
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(88))
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.bottomMargin = ctx.dp(10)
            }
        }
        // 左面额区
        val left = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(ctx.dp(96), LinearLayout.LayoutParams.MATCH_PARENT)
            background = getDrawable(
                if (c.disabled) R.drawable.bg_coupon_gray else R.drawable.bg_coupon)
        }
        val amount = TextView(this).apply {
            text = "¥${c.amount}"
            setTextColor(Design.WHITE)
            textSize = 24f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        }
        if (c.disabled) amount.setTextColor(0xFFC9CDD4.toInt())
        left.addView(amount)
        left.addView(TextView(this).apply {
            text = c.cond
            setTextColor(Design.WHITE)
            textSize = 11f
            setPadding(0, ctx.dp(4), 0, 0)
            if (c.disabled) setTextColor(0xFFC9CDD4.toInt())
        })
        card.addView(left)

        // 右信息区
        val right = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(ctx.dp(14), ctx.dp(12), ctx.dp(12), ctx.dp(12))
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.MATCH_PARENT, 1f)
            background = android.graphics.drawable.ColorDrawable(
                if (c.disabled) 0xFFF7F8FA.toInt() else Design.WHITE)
        }
        right.addView(TextView(this).apply {
            text = c.title
            setTextColor(if (c.disabled) 0xFFC9CDD4.toInt() else Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        right.addView(TextView(this).apply {
            text = c.desc
            setTextColor(if (c.disabled) 0xFFC9CDD4.toInt() else Design.TEXT_TERTIARY)
            textSize = 12f
            setPadding(0, ctx.dp(6), 0, 0)
        })
        right.addView(TextView(this).apply {
            text = c.expire
            setTextColor(if (c.disabled) 0xFFC9CDD4.toInt() else Design.TEXT_FAINT)
            textSize = 11f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        val btnRow = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.END or Gravity.CENTER_VERTICAL
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.MATCH_PARENT)
        }
        btnRow.addView(TextView(this).apply {
            text = if (c.disabled) "已过期" else "去使用"
            textSize = 12f
            gravity = Gravity.CENTER
            setPadding(ctx.dp(16), ctx.dp(6), ctx.dp(16), ctx.dp(6))
            setTextColor(if (c.disabled) 0xFFC9CDD4.toInt() else Design.PRIMARY)
            if (!c.disabled) {
                background = getDrawable(R.drawable.bg_btn_outline)
                isClickable = true
                setOnClickListener { ctx.toast("演示环境：已为您叠加该券") }
            }
        })
        right.addView(btnRow)
        card.addView(right)
        return card
    }
}
