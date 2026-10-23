package com.dawnerrand.client.ui.screens.address

import android.app.Activity
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
import com.dawnerrand.client.ui.util.confirmDialog
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/**
 * 收货地址：
 * - 普通模式：查看 / 编辑 / 删除 / 设为默认；
 * - 选择模式（EXTRA_PICK=true）：点击卡片把地址 id 返回给下单页。
 */
class AddressActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this
    private lateinit var list: LinearLayout
    private var pickMode = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_address)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        pickMode = intent.getBooleanExtra(EXTRA_PICK, false)

        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        findViewById<TextView>(R.id.tvTitle).text = if (pickMode) "选择收货地址" else "收货地址"
        findViewById<TextView>(R.id.btnAdd).text = "新增收货地址"
        list = findViewById(R.id.addressList)
        findViewById<View>(R.id.btnAdd).setOnClickListener {
            startActivity(Intent(this, AddressEditActivity::class.java))
        }
        render()
    }

    override fun onResume() {
        super.onResume()
        if (::list.isInitialized) render()
    }

    private fun render() {
        list.removeAllViews()
        if (DemoRepository.addresses.isEmpty()) {
            list.addView(TextView(this).apply {
                text = "还没有收货地址，点击下方按钮新增"
                setTextColor(Design.TEXT_FAINT)
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(0, ctx.dp(80), 0, ctx.dp(80))
            })
            return
        }
        DemoRepository.addresses.forEach { list.addView(buildCard(it)) }
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
            isClickable = true
            setOnClickListener {
                if (pickMode) {
                    setResult(Activity.RESULT_OK, Intent().putExtra(EXTRA_SELECTED_ID, a.id))
                    finish()
                } else {
                    startActivity(Intent(this@AddressActivity, AddressEditActivity::class.java)
                        .putExtra(AddressEditActivity.EXTRA_ADDRESS_ID, a.id))
                }
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
            text = "  ${DemoRepository.maskPhone(a.phone)}"
            setTextColor(Design.TEXT_SECONDARY)
            textSize = 13f
        })
        top.addView(View(this).apply {
            layoutParams = LinearLayout.LayoutParams(0, 1, 1f)
        })
        if (pickMode) {
            top.addView(TextView(this).apply {
                text = if (a.isDefault) "默认地址" else "选择"
                setTextColor(Design.PRIMARY)
                textSize = 12f
            })
        } else {
            top.addView(ImageView(this).apply {
                setImageResource(R.drawable.ic_edit)
                setColorFilter(Design.TEXT_FAINT)
                layoutParams = LinearLayout.LayoutParams(ctx.dp(16), ctx.dp(16))
                isClickable = true
                setOnClickListener {
                    startActivity(Intent(this@AddressActivity, AddressEditActivity::class.java)
                        .putExtra(AddressEditActivity.EXTRA_ADDRESS_ID, a.id))
                }
            })
        }
        card.addView(top)

        card.addView(TextView(this).apply {
            text = a.shortDetail
            setTextColor(Design.TEXT_SECONDARY)
            textSize = 13f
            setPadding(ctx.dp(16), 0, ctx.dp(16), ctx.dp(10))
        })

        val tags = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(ctx.dp(16), 0, ctx.dp(16), ctx.dp(12))
        }
        if (a.isDefault) tags.addView(tagView("默认", true))
        tags.addView(tagView(a.tag, false))
        tags.addView(View(this).apply {
            layoutParams = LinearLayout.LayoutParams(0, 1, 1f)
        })
        if (!pickMode) {
            if (!a.isDefault) {
                tags.addView(TextView(this).apply {
                    text = "设为默认"
                    setTextColor(Design.PRIMARY)
                    textSize = 12f
                    setPadding(ctx.dp(8), ctx.dp(4), ctx.dp(8), ctx.dp(4))
                    isClickable = true
                    setOnClickListener {
                        DemoRepository.setDefaultAddress(a.id)
                        render()
                        ctx.toast("已设为默认地址")
                    }
                })
            }
            tags.addView(TextView(this).apply {
                text = "删除"
                setTextColor(Design.TEXT_FAINT)
                textSize = 12f
                setPadding(ctx.dp(8), ctx.dp(4), ctx.dp(4), ctx.dp(4))
                isClickable = true
                setOnClickListener {
                    ctx.confirmDialog("删除地址", "确定删除「${a.name} · ${a.shortDetail}」吗？") {
                        DemoRepository.deleteAddress(a.id)
                        render()
                        ctx.toast("地址已删除")
                    }
                }
            })
        }
        card.addView(tags)
        return card
    }

    private fun tagView(text: String, active: Boolean): TextView = TextView(this).apply {
        this.text = text
        textSize = 11f
        setTextColor(if (active) Design.WHITE else Design.TEXT_TERTIARY)
        background = ContextCompat.getDrawable(
            this@AddressActivity,
            if (active) R.drawable.bg_tag_active else R.drawable.bg_tag_plain,
        )
        setPadding(ctx.dp(8), ctx.dp(2), ctx.dp(8), ctx.dp(2))
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
        (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(8)
    }

    companion object {
        const val EXTRA_PICK = "pick_address"
        const val EXTRA_SELECTED_ID = "selected_address_id"
    }
}
