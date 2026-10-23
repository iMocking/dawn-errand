package com.dawnerrand.client.ui.screens.orderdetail

import android.content.ActivityNotFoundException
import android.content.Intent
import android.graphics.Color
import android.net.Uri
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.Order
import com.dawnerrand.client.data.OrderStatus
import com.dawnerrand.client.ui.screens.customer.CustomerServiceActivity
import com.dawnerrand.client.ui.screens.errand.ErrandActivity
import com.dawnerrand.client.ui.screens.merchant.MerchantDetailActivity
import com.dawnerrand.client.ui.screens.ordertrack.OrderTrackActivity
import com.dawnerrand.client.ui.screens.paysuccess.PaySuccessActivity
import com.dawnerrand.client.ui.screens.review.ReviewActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.confirmDialog
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 订单详情：状态、订单信息、商品明细、进度、售后与操作。 */
class OrderDetailActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this
    private var orderId: String? = null
    private var order: Order? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_order_detail)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        orderId = intent.getStringExtra(EXTRA_ORDER_ID)
        order = orderId?.let { DemoRepository.orderById(it) }
        if (order == null) {
            ctx.toast("订单不存在或已删除")
            finish()
            return
        }
        findViewById<View>(R.id.rowCallShop).setOnClickListener { callShop() }
        findViewById<View>(R.id.rowCustomer).setOnClickListener {
            startActivity(Intent(this, CustomerServiceActivity::class.java))
        }
        findViewById<View>(R.id.rowRefund).setOnClickListener { applyRefund() }
        render()
    }

    override fun onResume() {
        super.onResume()
        order = orderId?.let { DemoRepository.orderById(it) }
        if (order != null) render()
    }

    private fun render() {
        order = orderId?.let { DemoRepository.orderById(it) } ?: order
        val o = order ?: return
        val status = findViewById<TextView>(R.id.tvStatus)
        status.text = o.status
        status.setTextColor(statusColor(o.status))
        findViewById<TextView>(R.id.tvStatusDesc).text = statusDesc(o.status)
        findViewById<TextView>(R.id.tvStatusTime).text = o.time
        findViewById<TextView>(R.id.tvShop).text = o.shop
        findViewById<TextView>(R.id.tvOrderNo).text = "订单编号：${o.orderNo} · ${o.payMethod}"
        findViewById<TextView>(R.id.tvAddress).text =
            if (o.address.isBlank()) "配送地址：门店自取" else "配送地址：${o.address}"
        findViewById<TextView>(R.id.tvRemark).text =
            if (o.remark.isBlank()) "订单备注：无" else "订单备注：${o.remark}"
        findViewById<TextView>(R.id.tvTotal).text = o.total

        val lineList = findViewById<LinearLayout>(R.id.lineList)
        lineList.removeAllViews()
        if (o.lines.isEmpty()) {
            lineList.addView(row(o.items, o.total, Design.TEXT_PRIMARY))
        } else {
            o.lines.forEach { line ->
                lineList.addView(row(
                    "${line.name} ×${line.count}",
                    "¥%.1f".format(line.amount),
                ))
            }
        }

        val feeList = findViewById<LinearLayout>(R.id.feeList)
        feeList.removeAllViews()
        feeList.addView(row("商品小计", "¥%.1f".format(o.subtotal)))
        if (o.deliveryFee > 0) feeList.addView(row("配送费", "¥%.1f".format(o.deliveryFee)))
        if (o.packingFee > 0) feeList.addView(row("打包费", "¥%.1f".format(o.packingFee)))
        if (o.discount > 0) feeList.addView(row("优惠立减", "-¥%.1f".format(o.discount), Design.PRIMARY))
        if (o.couponAmount > 0) {
            feeList.addView(row("优惠券${o.couponTitle}", "-¥%.1f".format(o.couponAmount), Design.PRIMARY))
        }
        if (o.reviewed) {
            feeList.addView(row("我的评价", "★${o.rating} ${o.reviewTags}", Design.STAR))
        }

        buildTimeline(o)
        buildActions(o)
    }

    private fun buildTimeline(o: Order) {
        val box = findViewById<LinearLayout>(R.id.timeline)
        box.removeAllViews()
        val steps: List<Pair<String, Boolean>> = when (o.status) {
            OrderStatus.CANCELED -> listOf("提交订单" to true, "订单取消" to true)
            OrderStatus.REFUND -> listOf(
                "提交订单" to true, "支付成功" to true, "订单送达" to true, "售后处理中" to true,
            )
            OrderStatus.PENDING_PAY -> listOf("提交订单" to true, "支付订单" to false, "商家接单" to false)
            OrderStatus.PENDING_ACCEPT -> listOf(
                "提交订单" to true, "支付成功" to true, "商家接单" to false, "骑手取货" to false,
            )
            OrderStatus.DELIVERING -> listOf(
                "提交订单" to true, "支付成功" to true, "商家接单" to true,
                "骑手取货" to true, "订单送达" to false,
            )
            else -> listOf(
                "提交订单" to true, "支付成功" to true, "商家接单" to true,
                "骑手取货" to true, "订单送达" to true,
            )
        }
        steps.forEachIndexed { index, (label, done) ->
            val row = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                gravity = Gravity.CENTER_VERTICAL
            }
            val dot = View(this).apply {
                background = ContextCompat.getDrawable(
                    this@OrderDetailActivity,
                    if (done) R.drawable.bg_dot_success else R.drawable.bg_dot_faint,
                )
                layoutParams = LinearLayout.LayoutParams(ctx.dp(10), ctx.dp(10))
            }
            row.addView(dot)
            row.addView(TextView(this).apply {
                text = label
                setTextColor(if (done) Design.TEXT_PRIMARY else Design.TEXT_FAINT)
                textSize = 13f
                setPadding(ctx.dp(12), 0, 0, 0)
            })
            box.addView(row)
            if (index < steps.lastIndex) {
                box.addView(View(this).apply {
                    setBackgroundColor(Design.DIVIDER)
                    layoutParams = LinearLayout.LayoutParams(ctx.dp(1), ctx.dp(18)).apply {
                        leftMargin = ctx.dp(4)
                    }
                })
            }
        }
    }

    private fun buildActions(o: Order) {
        val bar = findViewById<LinearLayout>(R.id.actionBar)
        bar.removeAllViews()
        val actions = DemoRepository.actionsFor(o)
        if (actions.isEmpty()) {
            bar.visibility = View.GONE
            return
        }
        bar.visibility = View.VISIBLE
        actions.forEachIndexed { index, action ->
            val primary = action in listOf("查看配送", "去支付", "去评价", "确认收货")
            bar.addView(TextView(this).apply {
                text = action
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(ctx.dp(20), ctx.dp(9), ctx.dp(20), ctx.dp(9))
                background = ContextCompat.getDrawable(
                    this@OrderDetailActivity,
                    if (primary) R.drawable.bg_btn else R.drawable.bg_btn_outline,
                )
                setTextColor(if (primary) Design.WHITE else Design.TEXT_SECONDARY)
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
                if (index > 0) (layoutParams as LinearLayout.LayoutParams).leftMargin = ctx.dp(10)
                isClickable = true
                setOnClickListener { handleAction(o, action) }
            })
        }
    }

    private fun handleAction(o: Order, action: String) {
        when (action) {
            "取消订单" -> ctx.confirmDialog("取消订单", "确定取消这个订单吗？") {
                DemoRepository.cancelOrder(o.id)
                render()
                ctx.toast("订单已取消")
            }
            "去支付" -> {
                DemoRepository.payOrder(o.id)
                startActivity(Intent(this, PaySuccessActivity::class.java)
                    .putExtra("amount", o.total.removePrefix("¥"))
                    .putExtra("type", o.type)
                    .putExtra("order_no", o.orderNo)
                    .putExtra("order_id", o.id))
                render()
            }
            "查看配送" -> startActivity(Intent(this, OrderTrackActivity::class.java)
                .putExtra(OrderTrackActivity.EXTRA_ORDER_ID, o.id))
            "确认收货" -> ctx.confirmDialog("确认收货", "确认已收到商品？确认后订单进入待评价。") {
                DemoRepository.confirmReceipt(o.id)
                render()
                ctx.toast("已确认收货")
            }
            "去评价", "评价" -> startActivity(Intent(this, ReviewActivity::class.java)
                .putExtra(ReviewActivity.EXTRA_ORDER_ID, o.id))
            "再来一单" -> reorder(o)
            "删除订单" -> ctx.confirmDialog("删除订单", "删除后记录不可恢复，确定删除吗？") {
                DemoRepository.deleteOrder(o.id)
                ctx.toast("订单已删除")
                finish()
            }
            "查看进度" -> ctx.toast("售后处理中，客服会在 24 小时内联系你")
            else -> ctx.toast("当前订单暂无更多操作")
        }
    }

    private fun reorder(o: Order) {
        if (o.type == "跑腿") {
            startActivity(Intent(this, ErrandActivity::class.java))
            return
        }
        if (DemoRepository.reorder(o)) {
            startActivity(Intent(this, MerchantDetailActivity::class.java)
                .putExtra("merchant_id", o.merchantId ?: "m1"))
            ctx.toast("已把商品加入购物车")
        } else {
            ctx.toast("商品已下架，去商家看看其他选择吧")
        }
    }

    private fun applyRefund() {
        val o = order ?: return
        if (o.status == OrderStatus.REFUND) {
            ctx.toast("售后正在处理中，请耐心等待")
            return
        }
        if (o.status == OrderStatus.PENDING_PAY || o.status == OrderStatus.CANCELED) {
            ctx.toast("当前订单状态无需申请售后")
            return
        }
        ctx.confirmDialog("申请售后", "将对该订单发起退款/售后申请，客服会在 24 小时内联系你。") {
            DemoRepository.applyRefund(o.id)
            render()
            ctx.toast("售后申请已提交")
        }
    }

    private fun callShop() {
        try {
            startActivity(Intent(Intent.ACTION_DIAL, Uri.parse("tel:07138888888")))
        } catch (e: ActivityNotFoundException) {
            ctx.toast("商家电话：0713-8888888")
        }
    }

    private fun row(label: String, value: String, valueColor: Int = Design.TEXT_PRIMARY): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, ctx.dp(5), 0, ctx.dp(5))
            addView(TextView(this@OrderDetailActivity).apply {
                text = label
                setTextColor(Design.TEXT_SECONDARY)
                textSize = 13f
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            })
            addView(TextView(this@OrderDetailActivity).apply {
                text = value
                setTextColor(valueColor)
                textSize = 13f
                typeface = android.graphics.Typeface.DEFAULT_BOLD
            })
        }
    }

    private fun statusColor(status: String): Int = when (status) {
        OrderStatus.PENDING_PAY, OrderStatus.PENDING_ACCEPT, OrderStatus.DELIVERING -> Design.PRIMARY
        OrderStatus.PENDING_REVIEW -> Design.ORANGE_DEEP
        OrderStatus.DONE -> Design.SUCCESS
        OrderStatus.REFUND -> Color.rgb(0x2B, 0x7F, 0xFF)
        else -> Design.TEXT_TERTIARY
    }

    private fun statusDesc(status: String): String = when (status) {
        OrderStatus.PENDING_PAY -> "订单已生成，请尽快完成支付"
        OrderStatus.PENDING_ACCEPT -> "已支付，等待商家 / 骑手接单"
        OrderStatus.DELIVERING -> "骑手正在取货，马上送到"
        OrderStatus.PENDING_REVIEW -> "订单已送达，期待你的评价"
        OrderStatus.DONE -> "订单已完成，感谢使用曙光跑腿"
        OrderStatus.CANCELED -> "订单已取消，如已支付将原路退回"
        OrderStatus.REFUND -> "售后处理中，客服会尽快联系你"
        else -> "订单状态更新中"
    }

    companion object {
        const val EXTRA_ORDER_ID = "order_id"
    }
}
