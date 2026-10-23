package com.dawnerrand.store.ui.screens.reviews

import android.os.Bundle
import android.text.InputType
import android.view.Gravity
import android.widget.EditText
import android.widget.TextView
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.R
import com.dawnerrand.store.data.Review
import com.dawnerrand.store.data.ReviewStore
import com.dawnerrand.store.databinding.ActivityReviewsBinding
import com.dawnerrand.store.ui.adapter.ReviewAdapter
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast

/**
 * 顾客评价：评分汇总 + 全部/好评/中评/差评筛选 + 商家回复。
 */
class ReviewsActivity : AppCompatActivity() {

    private lateinit var b: ActivityReviewsBinding
    private var filter = 0

    private val adapter = ReviewAdapter { review -> showReplyDialog(review) }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityReviewsBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.navBar.onBack = { finish() }

        b.reviewList.layoutManager = LinearLayoutManager(this)
        b.reviewList.adapter = adapter

        findViewById<TextView>(R.id.filterAll).setOnClickListener { selectFilter(0) }
        findViewById<TextView>(R.id.filterGood).setOnClickListener { selectFilter(1) }
        findViewById<TextView>(R.id.filterMid).setOnClickListener { selectFilter(2) }
        findViewById<TextView>(R.id.filterBad).setOnClickListener { selectFilter(3) }

        ReviewStore.reviews.observe(this) {
            adapter.submit(ReviewStore.filtered(filter))
            refreshTabs()
        }
        selectFilter(0)
    }

    private fun selectFilter(value: Int) {
        filter = value
        adapter.submit(ReviewStore.filtered(value))
        refreshTabs()
    }

    private fun refreshTabs() {
        val items = listOf(
            R.id.filterAll to 0,
            R.id.filterGood to 1,
            R.id.filterMid to 2,
            R.id.filterBad to 3,
        )
        items.forEach { (id, index) ->
            val view = findViewById<TextView>(id)
            val active = index == filter
            view.setBackgroundResource(
                if (active) com.dawnerrand.store.R.drawable.bg_range_active else 0
            )
            view.setTextColor(if (active) Design.WHITE else Design.TEXT_SECONDARY)
            view.setTypeface(
                view.typeface,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
            view.text = when (index) {
                0 -> "全部"
                1 -> "好评"
                2 -> "中评"
                else -> "差评"
            }
        }
    }

    private fun showReplyDialog(review: Review) {
        val input = EditText(this).apply {
            setText(review.reply.orEmpty())
            hint = "回复顾客，如：感谢支持，欢迎常来！"
            inputType = InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_FLAG_MULTI_LINE
            minLines = 3
            gravity = Gravity.TOP
            setPadding(dp(20), dp(8), dp(20), dp(8))
            setSelection(text.length)
        }
        val dialog = AlertDialog.Builder(this)
            .setTitle("回复 ${review.customer}")
            .setView(input)
            .setPositiveButton("发布回复", null)
            .setNegativeButton("取消", null)
            .create()
        dialog.setOnShowListener {
            dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener {
                val text = input.text.toString().trim()
                if (text.isEmpty()) {
                    input.error = "回复内容不能为空"
                } else {
                    ReviewStore.reply(review.id, text)
                    dialog.dismiss()
                    toast("回复已发布")
                }
            }
        }
        dialog.show()
    }
}
