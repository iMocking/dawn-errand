package com.dawnerrand.driver.ui.components

import android.content.Context
import android.graphics.Canvas
import android.graphics.Paint
import android.util.AttributeSet
import android.view.View
import com.dawnerrand.driver.ui.theme.Design

/**
 * 今日收入圆环：底环白色 20% 透明 + 进度环白色圆头描边，还原 home.html 的 110×110 圆环。
 */
class RingView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : View(context, attrs, defStyleAttr) {

    private val trackPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        style = Paint.Style.STROKE
        strokeWidth = resources.displayMetrics.density * 6f
        color = Design.WHITE_20
    }
    private val progressPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        style = Paint.Style.STROKE
        strokeWidth = resources.displayMetrics.density * 6f
        strokeCap = Paint.Cap.ROUND
        color = Design.WHITE
    }

    /** 0..1 */
    var progress: Float = 0.65f
        set(value) {
            field = value.coerceIn(0f, 1f)
            invalidate()
        }

    override fun onDraw(canvas: Canvas) {
        super.onDraw(canvas)
        val cx = width / 2f
        val cy = height / 2f
        val radius = cx - strokeWidthPx() / 2f
        canvas.drawArc(cx - radius, cy - radius, cx + radius, cy + radius, -90f, 360f, false, trackPaint)
        canvas.drawArc(cx - radius, cy - radius, cx + radius, cy + radius, -90f, 360f * progress, false, progressPaint)
    }

    private fun strokeWidthPx(): Float = resources.displayMetrics.density * 6f
}
