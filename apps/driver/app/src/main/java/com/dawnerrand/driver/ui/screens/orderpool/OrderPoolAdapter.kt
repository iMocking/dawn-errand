package com.dawnerrand.driver.ui.screens.orderpool

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.driver.R
import com.dawnerrand.driver.data.model.OrderItem
import com.dawnerrand.driver.databinding.ItemOrderPoolBinding

class OrderPoolAdapter(
    private var items: List<OrderItem>,
    private val onGrab: (OrderItem) -> Unit,
) : RecyclerView.Adapter<OrderPoolAdapter.VH>() {

    class VH(val binding: ItemOrderPoolBinding) : RecyclerView.ViewHolder(binding.root)

    fun submit(newItems: List<OrderItem>) {
        items = newItems
        notifyDataSetChanged()
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemOrderPoolBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = items.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val item = items[position]
        val b = holder.binding
        b.tvType.text = item.type
        b.tvType.setBackgroundResource(
            if (item.isGreenTag) R.drawable.bg_tag_takeout else R.drawable.bg_tag_food
        )
        b.tvType.setTextColor(
            if (item.isGreenTag) {
                holder.itemView.context.getColor(R.color.success)
            } else {
                holder.itemView.context.getColor(R.color.primary)
            }
        )
        b.tvDistance.text = item.distance
        b.tvPickup.text = item.pickup
        b.tvDropoff.text = item.dropoff
        b.tvReward.text = item.reward
        b.tvRewardLabel.text = item.rewardLabel
        b.btnGrab.setOnClickListener { onGrab(item) }
        b.root.setOnClickListener { onGrab(item) }
    }
}
