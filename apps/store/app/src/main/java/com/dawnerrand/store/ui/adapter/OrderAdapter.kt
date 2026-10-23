package com.dawnerrand.store.ui.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.LinearLayout
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.store.R
import com.dawnerrand.store.data.ORDER_COOKING
import com.dawnerrand.store.data.ORDER_DELIVERING
import com.dawnerrand.store.data.ORDER_DONE
import com.dawnerrand.store.data.ORDER_PENDING
import com.dawnerrand.store.data.StoreOrder
import com.dawnerrand.store.databinding.ItemOrderBinding
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dial
import com.dawnerrand.store.ui.util.dp

/**
 * 订单列表 Adapter：状态色 + 动态操作按钮（拒单/接单/通知出餐/查看配送/再来一单）。
 */
class OrderAdapter(
    private val onAction: (StoreOrder, String) -> Unit,
    private val onClick: (StoreOrder) -> Unit,
) : RecyclerView.Adapter<OrderAdapter.VH>() {

    private val data = mutableListOf<StoreOrder>()

    fun submit(list: List<StoreOrder>) {
        data.clear()
        data.addAll(list)
        notifyDataSetChanged()
    }

    inner class VH(val b: ItemOrderBinding) : RecyclerView.ViewHolder(b.root)

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemOrderBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = data.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val order = data[position]
        val b = holder.b
        b.orderNo.text = "订单 #${order.id}"
        b.orderStatus.text = statusText(order.status)
        b.orderStatus.setTextColor(statusColor(order.status))
        b.customerName.text = order.customerName
        b.customerAddr.text = order.address
        b.orderItems.text = order.items
        b.orderTotal.text = order.total
        b.callBtn.setOnClickListener {
            b.root.context.dial(order.customerPhone)
        }

        // 动态操作按钮
        b.actionBar.removeAllViews()
        order.actions.forEach { action ->
            b.actionBar.addView(
                buildActionBtn(b.actionBar, action) {
                    onAction(order, action)
                }
            )
        }

        b.root.setOnClickListener { onClick(order) }
    }

    private fun buildActionBtn(
        container: LinearLayout,
        label: String,
        onClick: () -> Unit,
    ): TextView {
        val btn = TextView(container.context)
        val lp = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT,
            LinearLayout.LayoutParams.WRAP_CONTENT
        ).apply { marginStart = container.context.dp(8) }
        btn.layoutParams = lp
        btn.text = label
        btn.textSize = 13f
        btn.setPadding(container.context.dp(16), container.context.dp(7), container.context.dp(16), container.context.dp(7))
        btn.gravity = android.view.Gravity.CENTER
        when (label) {
            "拒单" -> {
                btn.setBackgroundResource(R.drawable.bg_btn_danger)
                btn.setTextColor(Design.REWARD)
            }
            "通知出餐" -> {
                btn.setBackgroundResource(R.drawable.bg_btn_primary_16)
                btn.setTextColor(Design.WHITE)
            }
            else -> {
                btn.setBackgroundResource(R.drawable.bg_btn_primary_16)
                btn.setTextColor(Design.WHITE)
            }
        }
        btn.setOnClickListener { onClick() }
        return btn
    }

    private fun statusText(status: String): String = when (status) {
        ORDER_PENDING -> "待接单"
        ORDER_COOKING -> "制作中"
        ORDER_DELIVERING -> "配送中"
        ORDER_DONE -> "已完成"
        else -> status
    }

    private fun statusColor(status: String): Int = when (status) {
        ORDER_PENDING -> Design.REWARD
        ORDER_COOKING -> Design.PRIMARY
        ORDER_DELIVERING -> Design.BLUE
        else -> Design.SUCCESS
    }
}
