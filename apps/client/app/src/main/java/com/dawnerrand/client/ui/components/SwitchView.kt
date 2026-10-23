package com.dawnerrand.client.ui.components

import android.content.Context
import android.util.AttributeSet
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.util.dp

/**
 * 涓婄嚎鐘舵€佸紑鍏筹細杩樺師 home.html 椤堕儴"淇濇寔鍦ㄧ嚎"寮€鍏筹紙44脳24 杞ㄩ亾 + 20 鍦嗛挳锛夈€? */
class SwitchView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : FrameLayout(context, attrs, defStyleAttr) {

    private var checked: Boolean = true
    var onCheckedChange: ((Boolean) -> Unit)? = null

    /** 涓庝笟鍔″懡鍚嶅榻愶細寮€鍏冲綋鍓嶆槸鍚﹀湪绾?*/
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
