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
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.Order
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 评价：5 星 + 快捷标签（多选）+ 文本域 + 提交，结果写回订单。 */
class ReviewActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val tags = listOf("味道好", "分量足", "送得快", "包装好", "服务热情", "食材新鲜")
    private val ratingLabels = listOf("很差", "不满意", "一般", "满意", "非常满意")
    private val selectedTags = mutableSetOf<String>()
    private var rating = 5
    private var order: Order? = null
    private lateinit var tagViews: MutableList<TextView>

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_review)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        order = intent.getStringExtra(EXTRA_ORDER_ID)?.let { DemoRepository.orderById(it) }
        order?.let {
            findViewById<TextView>(R.id.tvShop).text = "评价订单：${it.shop}"
            if (it.reviewed) {
                rating = it.rating.coerceIn(1, 5)
                selectedTags.addAll(it.reviewTags.split("、").filter { t -> t.isNotBlank() })
                findViewById<EditText>(R.id.etComment).setText(it.reviewText)
            }
        }

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
        renderRatingLabel()

        tagViews = mutableListOf()
        val tagContainer = findViewById<LinearLayout>(R.id.tagRow)
        var tagRow: LinearLayout? = null
        tags.forEachIndexed { index, t ->
            if (index % 3 == 0) {
                tagRow = LinearLayout(this).apply {
                    orientation = LinearLayout.HORIZONTAL
                    layoutParams = LinearLayout.LayoutParams(
                        LinearLayout.LayoutParams.MATCH_PARENT,
                        LinearLayout.LayoutParams.WRAP_CONTENT)
                }
                tagContainer.addView(tagRow)
            }
            val chip = buildTagChip(t)
            tagViews.add(chip)
            tagRow?.addView(chip)
        }
        renderTags()

        findViewById<View>(R.id.btnSubmit).setOnClickListener { submit() }
    }

    private fun renderStars() {
        val starRow = findViewById<LinearLayout>(R.id.starRow)
        repeat(5) { i ->
            (starRow.getChildAt(i) as TextView).setTextColor(
                if (i < rating) Design.STAR else Design.TEXT_FAINT)
        }
        renderRatingLabel()
    }

    private fun renderRatingLabel() {
        findViewById<TextView>(R.id.tvRating).text = ratingLabels[rating - 1]
    }

    private fun buildTagChip(text: String): TextView = TextView(this).apply {
        this.text = text
        textSize = 13f
        gravity = Gravity.CENTER
        setPadding(ctx.dp(14), ctx.dp(8), ctx.dp(14), ctx.dp(8))
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
        (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(10)
        (layoutParams as LinearLayout.LayoutParams).bottomMargin = ctx.dp(10)
        isClickable = true
        setOnClickListener {
            if (!selectedTags.add(text)) selectedTags.remove(text)
            renderTags()
        }
    }

    private fun renderTags() {
        tagViews.forEach { view ->
            val selected = view.text.toString() in selectedTags
            view.setBackgroundResource(if (selected) R.drawable.bg_chip_active else R.drawable.bg_chip)
            view.setTextColor(if (selected) Design.WHITE else Design.TEXT_SECONDARY)
        }
    }

    private fun submit() {
        val text = findViewById<EditText>(R.id.etComment).text.toString().trim()
        if (text.isEmpty() && selectedTags.isEmpty()) {
            ctx.toast("说点什么或选个标签再提交吧")
            return
        }
        val o = order
        if (o == null) {
            ctx.toast("评价成功，感谢您的反馈！")
            finish()
            return
        }
        DemoRepository.reviewOrder(o.id, rating, selectedTags.toList(), text)
        ctx.toast("评价成功，感谢您的反馈！")
        finish()
    }

    companion object {
        const val EXTRA_ORDER_ID = "order_id"
    }
}
