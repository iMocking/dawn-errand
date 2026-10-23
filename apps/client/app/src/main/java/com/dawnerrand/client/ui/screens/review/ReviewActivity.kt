package com.dawnerrand.client.ui.screens.review

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

/** 评价：5 星 + 快捷标签（多选）+ 文本域 + 提交 */
class ReviewActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val tags = listOf("味道好", "分量足", "送得快", "包装好", "服务热情", "食材新鲜")
    private var rating = 5

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_review)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        // 星级
        val starRow = findViewById<LinearLayout>(R.id.starRow)
        repeat(5) { i ->
            val star = TextView(this).apply {
                text = "★"
                textSize = 30f
                setTextColor(if (i < rating) Design.STAR else Design.TEXT_FAINT)
                setPadding(ctx.dp(3), 0, ctx.dp(3), 0)
                isClickable = true
            }
            star.setOnClickListener {
                rating = i + 1
                renderStars()
            }
            starRow.addView(star)
        }

        // 标签（多选）
        val tagRow = findViewById<LinearLayout>(R.id.tagRow)
        tags.forEach { t ->
            tagRow.addView(buildTagChip(t))
        }

        findViewById<View>(R.id.btnSubmit).setOnClickListener {
            val text = findViewById<EditText>(R.id.etComment).text.toString().trim()
            if (text.isEmpty()) {
                ctx.toast("说点什么再提交吧")
            } else {
                ctx.toast("评价成功，感谢您的反馈！")
                finish()
            }
        }
    }

    private fun renderStars() {
        val starRow = findViewById<LinearLayout>(R.id.starRow)
        repeat(5) { i ->
            (starRow.getChildAt(i) as TextView).setTextColor(
                if (i < rating) Design.STAR else Design.TEXT_FAINT)
        }
    }

    private fun buildTagChip(text: String): TextView = TextView(this).apply {
        this.text = text
        textSize = 13f
        gravity = Gravity.CENTER
        setPadding(ctx.dp(14), ctx.dp(8), ctx.dp(14), ctx.dp(8))
        background = ContextCompat.getDrawable(this@ReviewActivity, R.drawable.bg_chip)
        setTextColor(Design.TEXT_SECONDARY)
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
        (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(10)
        (layoutParams as LinearLayout.LayoutParams).bottomMargin = ctx.dp(10)
        isClickable = true
        setOnClickListener {
            val selected = tag != true
            tag = selected
            setBackgroundResource(if (selected) R.drawable.bg_chip_active else R.drawable.bg_chip)
            setTextColor(if (selected) Design.WHITE else Design.TEXT_SECONDARY)
        }
    }
}
