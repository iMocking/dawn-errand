package com.dawnerrand.store.ui.adapter

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.store.data.ProfileMenuItem
import com.dawnerrand.store.databinding.ItemProfileMenuBinding
import com.dawnerrand.store.ui.util.toast

/**
 * 商家中心菜单 Adapter：图标 + 标签 + 右侧值 + 箭头（profile 稿）。
 */
class ProfileMenuAdapter(
    private val onItemClick: (ProfileMenuItem) -> Unit,
) : RecyclerView.Adapter<ProfileMenuAdapter.VH>() {

    private val data = mutableListOf<ProfileMenuItem>()

    fun submit(list: List<ProfileMenuItem>) {
        data.clear()
        data.addAll(list)
        notifyDataSetChanged()
    }

    inner class VH(val b: ItemProfileMenuBinding) : RecyclerView.ViewHolder(b.root)

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemProfileMenuBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = data.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val item = data[position]
        val b = holder.b
        b.menuIcon.setImageResource(item.iconRes)
        b.menuIcon.setColorFilter(item.iconColor)
        b.menuLabel.text = item.label
        b.menuValue.text = item.value.orEmpty()
        b.menuValue.visibility = if (item.value == null) android.view.View.GONE else android.view.View.VISIBLE
        b.divider.visibility = if (position < data.size - 1) android.view.View.VISIBLE else android.view.View.GONE
        b.root.setOnClickListener { onItemClick(item) }
    }
}
