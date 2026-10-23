package com.dawnerrand.store.ui.components

import android.content.Context
import android.util.AttributeSet
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import com.dawnerrand.store.R
import com.dawnerrand.store.ui.util.dp

/**
 * 开关：还原设计稿两种规格。
 * - 小号（菜品上架）：44×24 轨道 + 18 圆钮（menu.html）
 * - 大号（营业设置）：48×28 轨道 + 22 圆钮（business.html）
 */
class SwitchView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : FrameLayout(context, attrs, defStyleAttr) {

    /** true 为大号 48×28，false 为小号 44×24 */
    var big: Boolean = false
        set(value) {
            field = value
            rebuild()
        }

    private var checked: Boolean = true
    var onCheckedChange: ((Boolean) -> Unit)? = null

    var isOn: Boolean
        get() = checked
        set(value) {
            setChecked(value)
        }

    private lateinit var track: View
    private lateinit var thumb: View
    private var trackW = 44
    private var trackH = 24
    private var thumbSize = 18
    private var margin = 2

    init {
        val a = context.obtainStyledAttributes(attrs, R.styleable.SwitchView)
        big = a.getBoolean(R.styleable.SwitchView_big, false)
        a.recycle()
        rebuild()
        setOnClickListener { setChecked(!checked) }
    }

    private fun rebuild() {
        removeAllViews()
        if (big) {
            trackW = 48; trackH = 28; thumbSize = 22; margin = 3
        } else {
            trackW = 44; trackH = 24; thumbSize = 18; margin = 2
        }
        track = View(context).apply {
            layoutParams = LayoutParams(context.dp(trackW), context.dp(trackH))
        }
        thumb = View(context).apply {
            setBackgroundResource(R.drawable.bg_switch_thumb)
            layoutParams = LayoutParams(context.dp(thumbSize), context.dp(thumbSize)).apply {
                gravity = Gravity.CENTER_VERTICAL
                marginStart = context.dp(margin)
            }
        }
        addView(track)
        addView(thumb)
        refresh()
    }

    val isChecked: Boolean
        get() = checked

    /** 与 Android 开关命名对齐的监听器注册入口 */
    fun setOnCheckedChangeListener(listener: ((Boolean) -> Unit)?) {
        onCheckedChange = listener
    }

    fun setChecked(value: Boolean) {
        checked = value
        refresh()
        onCheckedChange?.invoke(checked)
    }

    private fun refresh() {
        if (!::track.isInitialized) return
        track.setBackgroundResource(
            if (checked) if (big) R.drawable.bg_switch_big_on else R.drawable.bg_switch_on
            else if (big) R.drawable.bg_switch_big_off else R.drawable.bg_switch_off
        )
        val lp = thumb.layoutParams as LayoutParams
        lp.marginStart = if (checked) context.dp(trackW - thumbSize - margin) else context.dp(margin)
        thumb.layoutParams = lp
    }
}
