package com.dawnerrand.store.ui.components

import android.content.Context
import android.graphics.Canvas
import android.graphics.Paint
import android.util.AttributeSet
import android.view.Gravity
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import com.dawnerrand.store.R
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp

/**
 * 底部 TabBar：还原设计稿 64 高、四 Tab（工作台/订单/菜品/我的），选中态主色 + 加粗。
 */
class TabBarView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : LinearLayout(context, attrs, defStyleAttr) {

    private val items = listOf(
        Triple(R.drawable.ic_home, "工作台", 0),
        Triple(R.drawable.ic_doc, "订单", 0),
        Triple(R.drawable.ic_layers, "菜品", 0),
        Triple(R.drawable.ic_user, "我的", 0),
    )
    private var selected = 0
    var onTabSelected: ((Int) -> Unit)? = null
    private val dividerPaint = Paint().apply { color = Design.DIVIDER }

    init {
        orientation = HORIZONTAL
        setBackgroundColor(Design.WHITE)
        items.forEachIndexed { index, item ->
            val (iconRes, label, _) = item
            addView(buildItem(index, iconRes, label))
        }
        refresh()
    }

    override fun onDraw(canvas: Canvas) {
        super.onDraw(canvas)
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
            text.setTypeface(
                text.typeface,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
        }
    }

    var selectedIndex: Int
        get() = selected
        set(value) {
            selected = value
            refresh()
        }
}
