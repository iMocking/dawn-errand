package com.dawnerrand.client.ui.screens.orderconfirm

import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.paysuccess.PaySuccessActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 确认订单：地址 + 商品明细/费用 + 送达/支付 + 合计提交 */
class OrderConfirmActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_order_confirm)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        val merchant = DemoRepository.merchants.first {
            it.id == intent.getStringExtra("merchant_id") ?: "m1"
        }

        // 商品明细 + 费用
        val goods = findViewById<LinearLayout>(R.id.cardGoods)
        goods.addView(TextView(this).apply {
            text = merchant.name
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            setPadding(0, 0, 0, ctx.dp(10))
        })
        var subtotal = 0.0
        DemoRepository.cart.forEach { item ->
            val row = row("${item.dish.name} ×${item.count}", "¥%.1f".format(item.dish.price * item.count))
            goods.addView(row)
            subtotal += item.dish.price * item.count
        }
        goods.addView(divider())
        goods.addView(row("商品小计", "¥%.1f".format(subtotal)))
        goods.addView(row("配送费", "¥3.0"))
        goods.addView(row("打包费", "¥1.5"))
        val discount = 3.0
        goods.addView(row("新用户立减", "-¥3.0", Design.PRIMARY))

        val total = subtotal + 3.0 + 1.5 - discount
        findViewById<TextView>(R.id.tvTotal).text = "合计 ¥%.1f".format(total)

        findViewById<View>(R.id.btnSubmit).setOnClickListener {
            // 生成订单并清空购物车
            val items = DemoRepository.cart.joinToString("、") { "${it.dish.name}×${it.count}" }
            DemoRepository.orders.toMutableList().add(
                com.dawnerrand.client.data.Order(
                    merchant.name, "配送中", true, items,
                    "今天 立即配送 · 预计25分钟送达", "¥%.1f".format(total),
                    merchant.iconRes, merchant.iconColor, merchant.bannerColorStart,
                    listOf("取消订单", "查看配送")
                )
            )
            DemoRepository.cart.clear()
            startActivity(Intent(this, PaySuccessActivity::class.java)
                .putExtra("amount", "%.1f".format(total))
                .putExtra("type", "外卖")
                .putExtra("order_no", "DD" + System.currentTimeMillis().toString().takeLast(8)))
            finish()
        }
    }

    private fun row(label: String, value: String, valueColor: Int = Design.TEXT_PRIMARY): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, ctx.dp(6), 0, ctx.dp(6))
            addView(TextView(this@OrderConfirmActivity).apply {
                text = label
                setTextColor(Design.TEXT_SECONDARY)
                textSize = 13f
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            })
            addView(TextView(this@OrderConfirmActivity).apply {
                text = value
                setTextColor(valueColor)
                textSize = 13f
                typeface = if (label == "商品小计") {
                    android.graphics.Typeface.DEFAULT_BOLD
                } else android.graphics.Typeface.DEFAULT
            })
        }
    }

    private fun divider(): View = View(this).apply {
        setBackgroundColor(Design.DIVIDER)
        layoutParams = LinearLayout.LayoutParams(LinearLayout.LayoutParams.MATCH_PARENT, 1)
    }
}
