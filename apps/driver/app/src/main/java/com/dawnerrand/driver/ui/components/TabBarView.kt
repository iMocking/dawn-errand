package com.dawnerrand.driver.ui.components

import android.content.Context
import android.graphics.Canvas
import android.graphics.Paint
import android.util.AttributeSet
import android.view.Gravity
import android.view.View
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import com.dawnerrand.driver.R
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.dp

/**
 * 底部 TabBar：还原设计稿 64 高、四 Tab（工作台/接单/订单/我的）。
 * 页面内用 [items] 指定图标与文案，点击回调 [onTabSelected]。
 */
class TabBarView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : LinearLayout(context, attrs, defStyleAttr) {

    private val items = listOf(
        Pair(R.drawable.ic_home, "工作台"),
        Pair(R.drawable.ic_clock, "接单"),
        Pair(R.drawable.ic_package, "订单"),
        Pair(R.drawable.ic_user, "我的"),
    )
    private var selected = 0
    var onTabSelected: ((Int) -> Unit)? = null
    private val dividerPaint = Paint().apply { color = Design.DIVIDER }

    init {
        orientation = HORIZONTAL
        setBackgroundColor(Design.WHITE)
        items.forEachIndexed { index, item ->
            val (iconRes, label) = item
            addView(buildItem(index, iconRes, label))
        }
        refresh()
    }

    override fun onDraw(canvas: Canvas) {
        super.onDraw(canvas)
        // 顶部 1px 分割线（不占布局宽度，避免挤压 weight 子项）
        canvas.drawRect(0f, 0f, width.toFloat(), 1f, dividerPaint)
    }

    private fun buildItem(index: Int, iconRes: Int, label: String): LinearLayout {
        val item = LinearLayout(context).apply {
            orientation = VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LayoutParams(0, LayoutParams.MATCH_PARENT, 1f)
            setPadding(0, context.dp(8), 0, context.dp(8))
            isClickable = true
            isFocusable = true
            setOnClickListener {
                selected = index
                refresh()
                onTabSelected?.invoke(index)
            }
        }
        val icon = ImageView(context).apply {
            setImageResource(iconRes)
            layoutParams = LayoutParams(context.dp(22), context.dp(22))
        }
        val text = TextView(context).apply {
            text = label
            textSize = 10f
            layoutParams = LayoutParams(LayoutParams.WRAP_CONTENT, LayoutParams.WRAP_CONTENT)
        }
        item.addView(icon)
        item.addView(text)
        return item
    }

    private fun refresh() {
        items.forEachIndexed { index, _ ->
            val item = getChildAt(index) as? LinearLayout ?: return@forEachIndexed
            val icon = item.getChildAt(0) as ImageView
            val text = item.getChildAt(1) as TextView
            val active = index == selected
            icon.setColorFilter(if (active) Design.PRIMARY else Design.TEXT_FAINT)
            text.setTextColor(if (active) Design.PRIMARY else Design.TEXT_FAINT)
            text.setTypeface(text.typeface, if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL)
        }
    }

    var selectedIndex: Int
        get() = selected
        set(value) {
            selected = value
            refresh()
        }
}
