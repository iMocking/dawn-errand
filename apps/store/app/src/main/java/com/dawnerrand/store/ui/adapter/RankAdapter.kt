package com.dawnerrand.store.ui.adapter

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.store.R
import com.dawnerrand.store.data.RankItem
import com.dawnerrand.store.databinding.ItemRankBinding
import com.dawnerrand.store.ui.theme.Design

/**
 * 热销排行 Adapter：名次徽标（红/橙/金/灰）+ 菜名 + 销量（statistics 稿）。
 */
class RankAdapter : RecyclerView.Adapter<RankAdapter.VH>() {

    private val data = mutableListOf<RankItem>()

    fun submit(list: List<RankItem>) {
        data.clear()
        data.addAll(list)
        notifyDataSetChanged()
    }

    inner class VH(val b: ItemRankBinding) : RecyclerView.ViewHolder(b.root)

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemRankBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = data.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val item = data[position]
        val b = holder.b
        b.rankText.text = item.rank.toString()
        val bg = when (item.rank) {
            1 -> R.drawable.bg_rank_red
            2 -> R.drawable.bg_rank_orange
            3 -> R.drawable.bg_rank_gold
            else -> R.drawable.bg_rank_gray
        }
        val fg = if (item.rank <= 3) Design.WHITE else Design.TEXT_SECONDARY
        b.rankNum.setBackgroundResource(bg)
        b.rankText.setTextColor(fg)
        b.rankName.text = item.name
        b.rankSales.text = item.sales
    }
}
