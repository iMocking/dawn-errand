package com.dawnerrand.store.ui.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.dawnerrand.store.R
import com.dawnerrand.store.data.Review
import com.dawnerrand.store.databinding.ItemReviewBinding
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp

/**
 * 顾客评价列表：星级、标签、商家回复；点击「回复」回调页面弹窗。
 */
class ReviewAdapter(
    private val onReply: (Review) -> Unit,
) : RecyclerView.Adapter<ReviewAdapter.VH>() {

    private val data = mutableListOf<Review>()

    fun submit(list: List<Review>) {
        data.clear()
        data.addAll(list)
        notifyDataSetChanged()
    }

    inner class VH(val b: ItemReviewBinding) : RecyclerView.ViewHolder(b.root)

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH =
        VH(ItemReviewBinding.inflate(LayoutInflater.from(parent.context), parent, false))

    override fun getItemCount(): Int = data.size

    override fun onBindViewHolder(holder: VH, position: Int) {
        val review = data[position]
        val b = holder.b
        b.avatar.text = review.customer.take(1)
        b.customerName.text = review.customer
        b.stars.text = "★".repeat(review.rating.coerceIn(0, 5)) +
            "☆".repeat((5 - review.rating).coerceAtLeast(0))
        b.reviewTime.text = review.time
        b.content.text = review.content

        b.tagRow.removeAllViews()
        review.tags.forEach { tag ->
            b.tagRow.addView(buildTag(b.tagRow.context, tag))
        }
        b.tagRow.visibility = if (review.tags.isEmpty()) View.GONE else View.VISIBLE

        if (review.reply.isNullOrBlank()) {
            b.replyBox.visibility = View.GONE
            b.replyBtn.text = "回复"
        } else {
            b.replyBox.visibility = View.VISIBLE
            b.replyText.text = "商家回复：${review.reply}"
            b.replyBtn.text = "修改回复"
        }
        b.replyBtn.setOnClickListener { onReply(review) }
    }

    private fun buildTag(ctx: android.content.Context, tag: String): TextView = TextView(ctx).apply {
        text = tag
        textSize = 10f
        setTextColor(Design.TEXT_TERTIARY)
        setBackgroundResource(R.drawable.bg_tag_light_grey)
        setPadding(ctx.dp(6), ctx.dp(2), ctx.dp(6), ctx.dp(2))
        layoutParams = ViewGroup.MarginLayoutParams(
            ViewGroup.LayoutParams.WRAP_CONTENT,
            ViewGroup.LayoutParams.WRAP_CONTENT
        ).apply { marginEnd = ctx.dp(6) }
    }
}
