package com.dawnerrand.client.ui.screens.search

import android.content.Intent
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
import com.dawnerrand.client.ui.screens.merchant.MerchantDetailActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 搜索：输入/热门/历史 → 匹配商家进入详情 */
class SearchActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val hot = listOf("农家小炒肉", "王家湾超市", "红富士苹果", "李记卤味")
    private val history = mutableListOf("农家小炒肉", "王家湾超市", "西红柿炒蛋")

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_search)
        window.decorView.padTopWithStatusBar()

        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        findViewById<View>(R.id.btnSearch).setOnClickListener { doSearch() }
        findViewById<EditText>(R.id.etSearch).setOnEditorActionListener { _, _, _ ->
            doSearch(); true
        }

        val hotRow = findViewById<LinearLayout>(R.id.hotRow)
        hot.forEach { hotRow.addView(hotChip(it)) }

        val historyRow = findViewById<LinearLayout>(R.id.historyRow)
        renderHistory()
    }

    private fun hotChip(keyword: String): TextView = TextView(this).apply {
        text = keyword
        textSize = 13f
        gravity = Gravity.CENTER
        setPadding(ctx.dp(14), ctx.dp(8), ctx.dp(14), ctx.dp(8))
        background = ContextCompat.getDrawable(this@SearchActivity, R.drawable.bg_chip)
        setTextColor(Design.TEXT_SECONDARY)
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
        (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(10)
        isClickable = true
        setOnClickListener { gotoKeyword(keyword) }
    }

    private fun renderHistory() {
        val row = findViewById<LinearLayout>(R.id.historyRow)
        row.removeAllViews()
        if (history.isEmpty()) {
            row.addView(TextView(this).apply {
                text = "暂无搜索记录"
                setTextColor(Design.TEXT_FAINT)
                textSize = 13f
                setPadding(0, ctx.dp(4), 0, ctx.dp(8))
            })
            return
        }
        history.forEach { kw ->
            val item = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                gravity = Gravity.CENTER_VERTICAL
                setPadding(0, ctx.dp(10), 0, ctx.dp(10))
                isClickable = true
                setOnClickListener { gotoKeyword(kw) }
            }
            item.addView(TextView(this).apply {
                text = kw
                textSize = 14f
                setTextColor(Design.TEXT_PRIMARY)
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            })
            item.addView(TextView(this).apply {
                text = "✕"
                textSize = 14f
                setTextColor(Design.TEXT_FAINT)
                setOnClickListener { history.remove(kw); renderHistory() }
            })
            row.addView(item)
        }
    }

    private fun doSearch() {
        val kw = findViewById<EditText>(R.id.etSearch).text.toString().trim()
        if (kw.isEmpty()) {
            ctx.toast("请输入搜索内容")
            return
        }
        gotoKeyword(kw)
    }

    private fun gotoKeyword(kw: String) {
        history.remove(kw)
        history.add(0, kw)
        val merchant = DemoRepository.merchants.firstOrNull {
            it.name.contains(kw) || it.tags.joinToString().contains(kw)
        }
        if (merchant != null) {
            startActivity(Intent(this, MerchantDetailActivity::class.java)
                .putExtra("merchant_id", merchant.id))
        } else {
            ctx.toast("未找到与“$kw”相关的商家")
        }
    }
}
