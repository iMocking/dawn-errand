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
import com.dawnerrand.client.data.Merchant
import com.dawnerrand.client.ui.screens.merchant.MerchantDetailActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 搜索：热门/历史/结果；匹配商家名称、标签与菜品，结果可点击进入商家。 */
class SearchActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val hot = listOf("农家小炒肉", "王家湾超市", "红富士苹果", "李记卤味")
    private lateinit var etSearch: EditText
    private lateinit var historyRow: LinearLayout
    private lateinit var resultSection: LinearLayout

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_search)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)

        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        etSearch = findViewById(R.id.etSearch)
        historyRow = findViewById(R.id.historyRow)
        resultSection = findViewById(R.id.resultSection)
        findViewById<View>(R.id.btnSearch).setOnClickListener { doSearch() }
        etSearch.setOnEditorActionListener { _, _, _ ->
            doSearch(); true
        }
        findViewById<View>(R.id.btnClearHistory).setOnClickListener {
            DemoRepository.clearSearchHistory()
            renderHistory()
            ctx.toast("搜索记录已清空")
        }

        val hotRow = findViewById<LinearLayout>(R.id.hotRow)
        hot.forEach { hotRow.addView(hotChip(it)) }
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
        historyRow.removeAllViews()
        val history = DemoRepository.searchHistory
        if (history.isEmpty()) {
            historyRow.addView(TextView(this).apply {
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
                setPadding(ctx.dp(8), 0, ctx.dp(4), 0)
                setOnClickListener {
                    DemoRepository.removeSearchHistory(kw)
                    renderHistory()
                }
            })
            historyRow.addView(item)
        }
    }

    private fun doSearch() {
        val kw = etSearch.text.toString().trim()
        if (kw.isEmpty()) {
            ctx.toast("请输入搜索内容")
            return
        }
        search(kw, autoOpen = true)
    }

    private fun gotoKeyword(keyword: String) {
        etSearch.setText(keyword)
        etSearch.setSelection(keyword.length)
        search(keyword, autoOpen = true)
    }

    private fun search(keyword: String, autoOpen: Boolean) {
        DemoRepository.addSearchHistory(keyword)
        renderHistory()
        val matched = matchMerchants(keyword)
        if (matched.isEmpty()) {
            showResults(keyword, emptyList())
            ctx.toast("未找到与“$keyword”相关的商家")
            return
        }
        if (autoOpen && matched.size == 1) {
            openMerchant(matched.first())
            return
        }
        showResults(keyword, matched)
    }

    private fun matchMerchants(keyword: String): List<Merchant> {
        val ids = mutableListOf<String>()
        DemoRepository.merchants.forEach { m ->
            if (m.name.contains(keyword, true) || m.recName.contains(keyword, true) ||
                m.tags.joinToString().contains(keyword, true)
            ) {
                ids.add(m.id)
            }
        }
        DemoRepository.merchants.forEach { m ->
            DemoRepository.dishesFor(m.id).forEach { dish ->
                if (dish.name.contains(keyword, true) || dish.category.contains(keyword, true)) {
                    ids.add(m.id)
                }
            }
        }
        return ids.distinct().map { DemoRepository.merchantById(it) }
    }

    private fun showResults(keyword: String, merchants: List<Merchant>) {
        resultSection.visibility = View.VISIBLE
        resultSection.removeAllViews()
        resultSection.addView(TextView(this).apply {
            text = if (merchants.isEmpty()) "“$keyword”的搜索结果" else "搜索结果 · ${merchants.size} 家商家"
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        if (merchants.isEmpty()) {
            resultSection.addView(TextView(this).apply {
                text = "换个关键词试试，比如“小炒肉”“超市”“水果”"
                setTextColor(Design.TEXT_FAINT)
                textSize = 13f
                setPadding(0, ctx.dp(12), 0, ctx.dp(12))
            })
            return
        }
        merchants.forEach { m ->
            resultSection.addView(buildResultCard(m))
        }
    }

    private fun buildResultCard(m: Merchant): LinearLayout {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            background = ContextCompat.getDrawable(this@SearchActivity, R.drawable.bg_card)
            setPadding(ctx.dp(12), ctx.dp(12), ctx.dp(12), ctx.dp(12))
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).topMargin = ctx.dp(10)
            isClickable = true
            setOnClickListener { openMerchant(m) }
        }
        card.addView(android.widget.ImageView(this).apply {
            setImageResource(m.iconRes)
            setColorFilter(m.iconColor)
            background = android.graphics.drawable.ColorDrawable(m.bannerColorStart)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(48), ctx.dp(48))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(12)
            setPadding(ctx.dp(10), ctx.dp(10), ctx.dp(10), ctx.dp(10))
        })
        val info = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        info.addView(TextView(this).apply {
            text = m.name
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        info.addView(TextView(this).apply {
            text = "⭐ ${m.score} · ${m.distance} · ${m.deliveryMinutes}分钟 · 月售${m.monthlySales}"
            setTextColor(Design.TEXT_TERTIARY)
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        card.addView(info)
        card.addView(TextView(this).apply {
            text = "进店 ›"
            setTextColor(Design.PRIMARY)
            textSize = 12f
        })
        card.setOnClickListener { openMerchant(m) }
        return card
    }

    private fun openMerchant(m: Merchant) {
        startActivity(Intent(this, MerchantDetailActivity::class.java)
            .putExtra("merchant_id", m.id))
    }
}
