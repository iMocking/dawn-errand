package com.dawnerrand.client.ui.screens.orders

import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.Order
import com.dawnerrand.client.ui.screens.ordertrack.OrderTrackActivity
import com.dawnerrand.client.ui.screens.review.ReviewActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 我的订单：4 筛选 Tab（进行中/已完成/已取消/退款售后）+ 订单卡列表 */
class OrdersActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val filters = listOf("进行中", "已完成", "已取消", "退款售后")
    private val tabs = mutableListOf<LinearLayout>()
    private var current = 0
    private lateinit var list: LinearLayout

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_orders)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        list = findViewById(R.id.list)

        val tabRow = findViewById<LinearLayout>(R.id.filterTabs)
        filters.forEachIndexed { i, name ->
            tabs.add(buildTab(i, name))
            tabRow.addView(tabs[i])
        }
        renderTabs()
        renderList()
    }

    private fun buildTab(index: Int, name: String): LinearLayout {
        val tab = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.MATCH_PARENT, 1f)
            isClickable = true
            setOnClickListener {
                current = index
                renderTabs()
                renderList()
            }
        }
        tab.addView(TextView(this).apply {
            text = name
            textSize = 14f
            gravity = Gravity.CENTER
            setPadding(0, ctx.dp(12), 0, ctx.dp(12))
        })
        return tab
    }

    private fun renderTabs() {
        tabs.forEachIndexed { i, tab ->
            val tv = tab.getChildAt(0) as TextView
            val active = i == current
            tv.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            tv.typeface = android.graphics.Typeface.create(
                android.graphics.Typeface.DEFAULT,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
            tv.setBackgroundResource(if (active) R.drawable.bg_tab_indicator else 0)
        }
    }

    private fun renderList() {
        list.removeAllViews()
        val shown = DemoRepository.orders.filter {
            when (current) {
                0 -> it.ongoing
                1 -> !it.ongoing
                else -> false
            }
        }
        if (shown.isEmpty()) {
            list.addView(emptyView())
            return
        }
        shown.forEach { list.addView(buildOrderCard(it)) }
    }

    private fun emptyView(): LinearLayout {
        val box = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(240))
        }
        box.addView(TextView(this).apply {
            text = "暂无相关订单"
            setTextColor(Design.TEXT_FAINT)
            textSize = 14f
        })
        return box
    }

    private fun buildOrderCard(o: Order): LinearLayout {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            background = ContextCompat.getDrawable(this@OrdersActivity, R.drawable.bg_card)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.bottomMargin = ctx.dp(10)
            }
        }
        // 头部
        val head = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(ctx.dp(16), ctx.dp(14), ctx.dp(16), ctx.dp(10))
        }
        head.addView(TextView(this).apply {
            text = o.shop
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        })
        head.addView(TextView(this).apply {
            text = o.status
            setTextColor(if (o.ongoing) Design.PRIMARY else Design.TEXT_TERTIARY)
            textSize = 12f
        })
        card.addView(head)

        // 主体
        val body = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(ctx.dp(16), 0, ctx.dp(16), 0)
        }
        val img = LinearLayout(this).apply {
            gravity = Gravity.CENTER
            background = android.graphics.drawable.ColorDrawable(o.iconBg)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(56), ctx.dp(56))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(12)
        }
        img.addView(ImageView(this).apply {
            setImageResource(o.iconRes)
            setColorFilter(o.iconColor)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(26), ctx.dp(26))
        })
        body.addView(img)

        val info = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        info.addView(TextView(this).apply {
            text = o.items
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 14f
            maxLines = 2
        })
        info.addView(TextView(this).apply {
            text = o.time
            setTextColor(Design.TEXT_TERTIARY)
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        body.addView(info)
        body.addView(TextView(this).apply {
            text = o.total
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 16f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        card.addView(body)

        // 操作按钮
        val actions = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.END
            setPadding(ctx.dp(16), ctx.dp(12), ctx.dp(16), ctx.dp(14))
        }
        o.actions.forEachIndexed { i, action ->
            val primary = action == "查看配送"
            val btn = TextView(this).apply {
                text = action
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(ctx.dp(20), ctx.dp(8), ctx.dp(20), ctx.dp(8))
                background = ContextCompat.getDrawable(
                    this@OrdersActivity,
                    if (primary) R.drawable.bg_btn else R.drawable.bg_btn_outline
                )
                setTextColor(if (primary) Design.WHITE else Design.TEXT_SECONDARY)
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
                if (i > 0) (layoutParams as LinearLayout.LayoutParams).leftMargin = ctx.dp(10)
                isClickable = true
                setOnClickListener { onAction(action) }
            }
            actions.addView(btn)
        }
        card.addView(actions)
        return card
    }

    private fun onAction(action: String) {
        when (action) {
            "查看配送" -> startActivity(Intent(this, OrderTrackActivity::class.java))
            "评价" -> startActivity(Intent(this, ReviewActivity::class.java))
            else -> ctx.toast("演示环境：$action 已处理")
        }
    }
}
