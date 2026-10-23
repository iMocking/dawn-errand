package com.dawnerrand.driver.ui.screens.orders

import android.view.LayoutInflater
import android.view.ViewGroup
import android.text.SpannableString
import android.text.Spanned
import android.text.style.ForegroundColorSpan
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.driver.data.model.OrderRecord
import com.dawnerrand.driver.databinding.ItemOrderHistoryBinding
import com.dawnerrand.driver.ui.theme.Design

class OrdersAdapter(
    private var items: List<OrderRecord>,
    private val onClick: (OrderRecord) -> Unit,
) : RecyclerView.Adapter<OrdersAdapter.VH>() {

    class VH(val binding: ItemOrderHistoryBinding) : RecyclerView.ViewHolder(binding.root)

    fun submit(newItems: List<OrderRecord>) {
        items = newItems
        notifyDataSetChanged()
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemOrderHistoryBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = items.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val item = items[position]
        val b = holder.binding
        b.tvTime.text = item.time
        b.tvStatus.text = item.status
        b.tvStatus.setTextColor(holder.itemView.context.getColor(com.dawnerrand.driver.R.color.success))
        b.tvTypeInfo.text = item.typeInfo
        b.tvEarn.text = item.earn

        // 路线：from 橙色 → 灰箭头 → to 主色
        val route = item.routeFrom + " → " + item.routeTo
        val spannable = SpannableString(route)
        spannable.setSpan(
            ForegroundColorSpan(Design.PRIMARY),
            0,
            item.routeFrom.length,
            Spanned.SPAN_EXCLUSIVE_EXCLUSIVE,
        )
        spannable.setSpan(
            ForegroundColorSpan(Design.TEXT_FAINT),
            item.routeFrom.length,
            item.routeFrom.length + 3,
            Spanned.SPAN_EXCLUSIVE_EXCLUSIVE,
        )
        b.tvRoute.text = spannable
        b.root.setOnClickListener { onClick(item) }
    }
}
