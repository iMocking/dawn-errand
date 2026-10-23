package com.dawnerrand.store.ui.adapter

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.store.R
import com.dawnerrand.store.data.Dish
import com.dawnerrand.store.databinding.ItemDishBinding
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.toast

/**
 * 菜品列表 Adapter：占位图 + 价格 + 上架开关 + 编辑入口（menu 稿）。
 */
class DishAdapter(
    private val onToggleSale: (Dish) -> Unit,
    private val onEdit: (Dish) -> Unit,
) : RecyclerView.Adapter<DishAdapter.VH>() {

    private val data = mutableListOf<Dish>()

    fun submit(list: List<Dish>) {
        data.clear()
        data.addAll(list)
        notifyDataSetChanged()
    }

    inner class VH(val b: ItemDishBinding) : RecyclerView.ViewHolder(b.root)

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemDishBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = data.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val dish = data[position]
        val b = holder.b
        b.dishImage.setBackgroundResource(dish.bgRes)
        b.dishIcon.setImageResource(dish.iconRes)
        b.dishIcon.setColorFilter(dish.iconColor)
        b.dishName.text = dish.name
        b.dishMeta.text = "${dish.monthlySales} · ${dish.stockText}"
        b.dishPrice.text = "¥${dish.price}"
        b.saleSwitch.setOnCheckedChangeListener(null)
        b.saleSwitch.isOn = dish.onSale
        b.saleSwitch.setOnCheckedChangeListener { checked ->
            onToggleSale(dish.copy(onSale = checked))
        }
        b.editLink.setOnClickListener {
            onEdit(dish)
        }
        b.root.setOnClickListener { b.root.context.toast("编辑「${dish.name}」") }
    }
}
