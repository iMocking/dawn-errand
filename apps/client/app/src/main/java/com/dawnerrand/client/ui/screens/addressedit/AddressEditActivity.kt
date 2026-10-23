package com.dawnerrand.client.ui.screens.addressedit

import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 编辑地址：联系人/手机号/详细地址 + 标签 chip + 保存 */
class AddressEditActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val tagOptions = listOf("家", "公司", "学校", "其他")

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_address_edit)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        val tagRow = findViewById<LinearLayout>(R.id.tagRow)
        tagOptions.forEach { t ->
            tagRow.addView(tagChip(t))
        }

        findViewById<View>(R.id.btnSave).setOnClickListener {
            val name = findViewById<EditText>(R.id.etName).text.toString().trim()
            val phone = findViewById<EditText>(R.id.etPhone).text.toString().trim()
            val detail = findViewById<EditText>(R.id.etDetail).text.toString().trim()
            when {
                name.isEmpty() -> ctx.toast("请填写联系人")
                phone.length != 11 -> ctx.toast("请填写正确的11位手机号")
                detail.isEmpty() -> ctx.toast("请填写详细地址")
                else -> {
                    ctx.toast("地址已保存")
                    finish()
                }
            }
        }
    }

    private fun tagChip(text: String): TextView = TextView(this).apply {
        this.text = text
        textSize = 13f
        gravity = Gravity.CENTER
        setPadding(ctx.dp(16), ctx.dp(8), ctx.dp(16), ctx.dp(8))
        background = ContextCompat.getDrawable(this@AddressEditActivity, R.drawable.bg_chip)
        setTextColor(Design.TEXT_SECONDARY)
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
        (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(10)
        isClickable = true
        setOnClickListener {
            val selected = tag != true
            tag = selected
            setBackgroundResource(if (selected) R.drawable.bg_chip_active else R.drawable.bg_chip)
            setTextColor(if (selected) Design.WHITE else Design.TEXT_SECONDARY)
        }
    }
}
