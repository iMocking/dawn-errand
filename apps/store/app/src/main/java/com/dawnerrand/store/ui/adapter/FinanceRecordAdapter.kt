package com.dawnerrand.store.ui.adapter

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.store.data.FinanceRecord
import com.dawnerrand.store.databinding.ItemFinanceRecordBinding
import com.dawnerrand.store.ui.theme.Design

/**
 * 收支记录 Adapter：收入绿色 + / 支出深色 -（finance 稿）。
 */
class FinanceRecordAdapter : RecyclerView.Adapter<FinanceRecordAdapter.VH>() {

    private val data = mutableListOf<FinanceRecord>()

    fun submit(list: List<FinanceRecord>) {
        data.clear()
        data.addAll(list)
        notifyDataSetChanged()
    }

    inner class VH(val b: ItemFinanceRecordBinding) : RecyclerView.ViewHolder(b.root)

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemFinanceRecordBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = data.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val record = data[position]
        val b = holder.b
        b.recordTitle.text = record.title
        b.recordTime.text = record.time
        b.recordAmount.text = record.amount
        b.recordAmount.setTextColor(
            if (record.incoming) Design.SUCCESS else Design.TEXT_PRIMARY
        )
    }
}
