package com.dawnerrand.store.ui.screens.orders

import android.os.Bundle
import android.view.View
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.data.ORDER_COOKING
import com.dawnerrand.store.data.ORDER_DELIVERING
import com.dawnerrand.store.data.ORDER_DONE
import com.dawnerrand.store.data.ORDER_PENDING
import com.dawnerrand.store.data.StoreOrder
import com.dawnerrand.store.data.StoreState
import com.dawnerrand.store.databinding.ActivityOrderDetailBinding
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.padBottomWithNavBar
import com.dawnerrand.store.ui.util.toast

/**
 * 订单详情：状态卡 / 顾客信息 / 商品明细与价格构成 / 订单信息 / 底部操作（order-detail 稿）。
 */
class OrderDetailActivity : AppCompatActivity() {

    private lateinit var b: ActivityOrderDetailBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityOrderDetailBinding.inflate(layoutInflater)
        setContentView(b.root)

        val orderId = intent.getStringExtra("orderId") ?: "20261022001"
        val vm = OrderDetailViewModel(orderId)
        val order = vm.order

        b.backBtn.setOnClickListener { finish() }
        b.bottomBar.padBottomWithNavBar()

        // 状态卡
        b.statusBig.text = vm.statusBig
        b.statusSmall.text = vm.statusSmall

        // 顾客信息
        b.customerName.text = order?.customerName ?: "-"
        b.customerPhone.text = order?.customerPhone ?: "-"
        b.customerAddr.text = buildString {
            append("湖北省黄冈市浠水县 · ")
            append(order?.address?.substringBefore(" · ") ?: "")
        }
        b.callBtn.setOnClickListener {
            toast("拨打 ${order?.customerName} ${order?.customerPhone}")
        }

        // 商品明细
        b.itemList.removeAllViews()
        vm.items.forEach { (name, price, _) ->
            val row = LinearLayoutRow(name, price)
            b.itemList.addView(row)
        }
        b.subTotal.text = vm.subTotal
        b.orderTotal.text = order?.total ?: ""
        b.remark.text = vm.items.firstOrNull { it.third != null }?.let { "📝 顾客备注：${it.third}" }
            ?: "📝 顾客备注：无"

        // 订单信息
        b.orderMeta.text = buildString {
            appendLine("订单编号：${order?.id ?: "-"}")
            appendLine("下单时间：今天 12:26")
            appendLine("支付方式：微信支付")
            append("配送方式：曙光跑腿骑手")
        }

        // 底部按钮
        b.secondaryBtn.text = vm.bottomSecondary
        b.primaryBtn.text = vm.bottomPrimary
        b.secondaryBtn.setOnClickListener {
            when (order?.status) {
                ORDER_PENDING -> {
                    StoreState.reject(order.id)
                    toast("已拒单")
                    finish()
                }
                else -> toast("联系顾客 ${order?.customerPhone}")
            }
        }
        b.primaryBtn.setOnClickListener {
            when (order?.status) {
                ORDER_PENDING -> {
                    StoreState.accept(order.id)
                    toast("已接单，请尽快制作")
                    finish()
                }
                ORDER_COOKING -> {
                    StoreState.notifyReady(order.id)
                    toast("已通知骑手取餐")
                    finish()
                }
                else -> toast("订单已完成")
            }
        }
    }

    private fun LinearLayoutRow(name: String, price: String): View {
        val row = android.widget.LinearLayout(this).apply {
            orientation = android.widget.LinearLayout.HORIZONTAL
            gravity = android.view.Gravity.CENTER_VERTICAL
            layoutParams = android.widget.LinearLayout.LayoutParams(
                android.widget.LinearLayout.LayoutParams.MATCH_PARENT,
                android.widget.LinearLayout.LayoutParams.WRAP_CONTENT
            ).apply { topMargin = dp(8) }
        }
        val n = TextView(this).apply {
            text = name
            textSize = 14f
            setTextColor(0xFF1D2129.toInt())
            layoutParams = android.widget.LinearLayout.LayoutParams(
                0,
                android.widget.LinearLayout.LayoutParams.WRAP_CONTENT,
                1f
            )
        }
        val p = TextView(this).apply {
            text = price
            textSize = 14f
            setTextColor(0xFF4E5969.toInt())
        }
        row.addView(n)
        row.addView(p)
        return row
    }
}
