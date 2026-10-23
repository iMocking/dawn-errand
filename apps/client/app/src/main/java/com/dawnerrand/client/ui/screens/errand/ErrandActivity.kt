package com.dawnerrand.client.ui.screens.errand

import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.paysuccess.PaySuccessActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padBottomWithNavBar
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 跑腿代办：4 类型 Tab + 取/收地址 + 描述 + 费用预估 + 立即下单 */
class ErrandActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val types = listOf(
        Pair("帮我买", R.drawable.ic_cart),
        Pair("帮我送", R.drawable.ic_hexagon),
        Pair("帮我取", R.drawable.ic_mail),
        Pair("帮我办", R.drawable.ic_tab_orders),
    )
    private val tabViews = mutableListOf<LinearLayout>()
    private var current = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_errand)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        // 类型 Tab
        val tabRow = findViewById<LinearLayout>(R.id.typeTabs)
        types.forEachIndexed { i, (name, icon) ->
            tabViews.add(buildTypeTab(i, name, icon))
            tabRow.addView(tabViews[i])
        }
        renderTabs()

        // 交换按钮
        findViewById<View>(R.id.btnSwap).setOnClickListener {
            val a = findViewById<TextView>(R.id.tvPick)
            val b = findViewById<TextView>(R.id.tvDrop)
            val t = a.text; a.text = b.text; b.text = t
        }

        findViewById<View>(R.id.btnSubmit).setOnClickListener { submit() }
    }

    private fun buildTypeTab(index: Int, name: String, iconRes: Int): LinearLayout {
        val tab = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(0, ctx.dp(64), 1f)
            isClickable = true
            setOnClickListener {
                current = index
                renderTabs()
            }
        }
        val icon = ImageView(this).apply {
            setImageResource(iconRes)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(24), ctx.dp(24))
        }
        val label = TextView(this).apply {
            text = name
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
        }
        tab.addView(icon)
        tab.addView(label)
        return tab
    }

    private fun renderTabs() {
        tabViews.forEachIndexed { i, tab ->
            val icon = tab.getChildAt(0) as ImageView
            val label = tab.getChildAt(1) as TextView
            val active = i == current
            icon.setColorFilter(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            label.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            label.typeface = android.graphics.Typeface.DEFAULT.let {
                android.graphics.Typeface.create(it, if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL)
            }
        }
    }

    private fun submit() {
        val desc = findViewById<EditText>(R.id.etDesc).text.toString().trim()
            .ifEmpty { "代买代送（按清单采购）" }
        val order = com.dawnerrand.client.data.Order(
            "跑腿代办 · 镇中心", "进行中", true,
            "${types[current].first}：$desc", "今天 即刻出发", "¥8.0",
            R.drawable.ic_tab_errand, Design.SUCCESS, Color.rgb(0xE8, 0xF7, 0xF0),
            listOf("取消订单", "查看配送")
        )
        startActivity(Intent(this, PaySuccessActivity::class.java)
            .putExtra("amount", "8.0")
            .putExtra("type", "跑腿")
            .putExtra("order_no", "DD" + System.currentTimeMillis().toString().takeLast(8)))
        finish()
    }
}
