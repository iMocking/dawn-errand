package com.dawnerrand.client.ui.screens.address

import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.Address
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.addressedit.AddressEditActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 收货地址列表：3 地址卡（标签 + 编辑） + 底部新增 */
class AddressActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_address)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        val list = findViewById<LinearLayout>(R.id.addressList)
        DemoRepository.addresses.forEach { list.addView(buildCard(it)) }

        findViewById<View>(R.id.btnAdd).setOnClickListener {
            startActivity(Intent(this, AddressEditActivity::class.java))
        }
    }

    private fun buildCard(a: Address): LinearLayout {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            background = ContextCompat.getDrawable(this@AddressActivity, R.drawable.bg_card)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.bottomMargin = ctx.dp(10)
            }
        }
        val top = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(ctx.dp(16), ctx.dp(14), ctx.dp(16), ctx.dp(6))
        }
        top.addView(TextView(this).apply {
            text = a.name
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        top.addView(TextView(this).apply {
            text = "  ${a.phone}"
            setTextColor(Design.TEXT_SECONDARY)
            textSize = 13f
        })
        top.addView(View(this).apply {
            layoutParams = LinearLayout.LayoutParams(0, 1, 1f)
        })
        top.addView(ImageView(this).apply {
            setImageResource(R.drawable.ic_edit)
            setColorFilter(Design.TEXT_FAINT)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(16), ctx.dp(16))
            isClickable = true
            setOnClickListener {
                startActivity(Intent(this@AddressActivity, AddressEditActivity::class.java))
            }
        })
        card.addView(top)

        card.addView(TextView(this).apply {
            text = a.detail
            setTextColor(Design.TEXT_SECONDARY)
            textSize = 13f
            setPadding(ctx.dp(16), 0, ctx.dp(16), ctx.dp(10))
        })

        val tags = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(ctx.dp(16), 0, ctx.dp(16), ctx.dp(14))
        }
        a.tags.forEach { t ->
            val isDefault = t == "默认"
            tags.addView(TextView(this).apply {
                text = t
                textSize = 11f
                setTextColor(if (isDefault) Design.WHITE else Design.TEXT_TERTIARY)
                background = ContextCompat.getDrawable(this@AddressActivity,
                    if (isDefault) R.drawable.bg_tag_active else R.drawable.bg_tag_plain)
                setPadding(ctx.dp(8), ctx.dp(2), ctx.dp(8), ctx.dp(2))
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
                (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(8)
            })
        }
        card.addView(tags)
        return card
    }
}
