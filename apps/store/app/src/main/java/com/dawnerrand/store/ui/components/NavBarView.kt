package com.dawnerrand.store.ui.components

import android.app.Activity
import android.content.Context
import android.util.AttributeSet
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.TextView
import com.dawnerrand.store.R
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp

/**
 * 统一导航栏：44dp 白底、居中标题、左侧返回箭头（与设计稿 .nav 一致）。
 */
class NavBarView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : FrameLayout(context, attrs, defStyleAttr) {

    private val titleView = TextView(context).apply {
        textSize = 17f
        setTextColor(Design.TEXT_PRIMARY)
        setTypeface(typeface, android.graphics.Typeface.BOLD)
        gravity = Gravity.CENTER
        layoutParams = LayoutParams(LayoutParams.WRAP_CONTENT, LayoutParams.WRAP_CONTENT, Gravity.CENTER)
    }

    private val backView = ImageView(context).apply {
        setImageResource(R.drawable.ic_back)
        setColorFilter(Design.TEXT_PRIMARY)
        setPadding(context.dp(10), context.dp(10), context.dp(10), context.dp(10))
        layoutParams = LayoutParams(context.dp(44), LayoutParams.MATCH_PARENT, Gravity.START).apply {
            marginStart = context.dp(8)
        }
        contentDescription = "返回"
        setOnClickListener {
            if (onBack != null) onBack?.invoke() else (context as? Activity)?.finish()
        }
    }

    var onBack: (() -> Unit)? = null

    init {
        setBackgroundColor(Design.WHITE)
        addView(titleView)
        addView(backView)

        val a = context.obtainStyledAttributes(attrs, R.styleable.NavBarView)
        val label = a.getString(R.styleable.NavBarView_navTitle)
        val showBack = a.getBoolean(R.styleable.NavBarView_showBack, true)
        a.recycle()
        titleView.text = label.orEmpty()
        backView.visibility = if (showBack) View.VISIBLE else View.GONE
    }

    fun setTitleText(text: String) {
        titleView.text = text
    }
}
