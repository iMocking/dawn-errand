package com.dawnerrand.driver.ui.components

import android.content.Context
import android.util.AttributeSet
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import com.dawnerrand.driver.R
import com.dawnerrand.driver.ui.util.dp

/**
 * 上线状态开关：还原 home.html 顶部"保持在线"开关（44×24 轨道 + 20 圆钮）。
 */
class SwitchView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : FrameLayout(context, attrs, defStyleAttr) {

    private var checked: Boolean = true
    var onCheckedChange: ((Boolean) -> Unit)? = null

    /** 与业务命名对齐：开关当前是否在线 */
    var isOn: Boolean
        get() = checked
        set(value) {
            setChecked(value)
        }

    private val track: View = View(context).apply {
        setBackgroundResource(R.drawable.bg_switch_on)
        layoutParams = LayoutParams(context.dp(44), context.dp(24))
    }
    private val thumb: View = View(context).apply {
        setBackgroundResource(R.drawable.bg_switch_thumb)
        layoutParams = LayoutParams(context.dp(20), context.dp(20)).apply {
            gravity = Gravity.CENTER_VERTICAL
            marginStart = context.dp(2)
        }
    }

    init {
        addView(track)
        addView(thumb)
        refresh()
        setOnClickListener { setChecked(!checked) }
    }

    val isChecked: Boolean
        get() = checked

    fun setChecked(value: Boolean) {
        checked = value
        refresh()
        onCheckedChange?.invoke(checked)
    }

    private fun refresh() {
        track.setBackgroundResource(if (checked) R.drawable.bg_switch_on else R.drawable.bg_switch_off)
        val lp = thumb.layoutParams as LayoutParams
        lp.marginStart = if (checked) context.dp(22) else context.dp(2)
        thumb.layoutParams = lp
    }
}
