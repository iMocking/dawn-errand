package com.dawnerrand.store.ui.components

import android.content.Context
import android.graphics.Canvas
import android.graphics.LinearGradient
import android.graphics.Paint
import android.graphics.RectF
import android.graphics.Shader
import android.util.AttributeSet
import android.view.View
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.sp
import kotlin.math.max

/**
 * 营业趋势柱状图：与 design/store/statistics.html 的 .bar-chart 一致。
 * 底部对齐、圆角柱体、柱间留白、下方 10sp 标签；柱高按最大值等比缩放。
 */
class BarChartView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0,
) : View(context, attrs, defStyleAttr) {

    private var labels: List<String> = emptyList()
    private var values: List<Float> = emptyList()

    private val barPaint = Paint(Paint.ANTI_ALIAS_FLAG)
    private val labelPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        color = Design.TEXT_TERTIARY
        textSize = context.sp(10).toFloat()
        textAlign = Paint.Align.CENTER
    }

    fun setData(labels: List<String>, values: List<Float>) {
        this.labels = labels
        this.values = values
        invalidate()
    }

    override fun onMeasure(widthMeasureSpec: Int, heightMeasureSpec: Int) {
        val w = resolveSize(suggestedMinimumWidth, widthMeasureSpec)
        val desiredH = context.dp(130)
        val h = resolveSize(if (heightMeasureSpec == 0) desiredH else desiredH, heightMeasureSpec)
        setMeasuredDimension(w, h)
    }

    override fun onDraw(canvas: Canvas) {
        super.onDraw(canvas)
        if (values.isEmpty() || labels.size != values.size) return

        val padLeft = paddingLeft.toFloat()
        val padRight = paddingRight.toFloat()
        val padTop = paddingTop.toFloat()
        val padBottom = paddingBottom.toFloat()
        val labelH = context.sp(18).toFloat()
        val usableW = width - padLeft - padRight
        val chartH = height - padTop - padBottom - labelH
        if (usableW <= 0f || chartH <= 0f) return

        val maxValue = max(values.maxOrNull() ?: 1f, 0.01f)
        val slot = usableW / values.size
        val barW = slot * 0.56f
        val radius = context.dp(4).toFloat()
        val chartBottom = padTop + chartH

        barPaint.shader = LinearGradient(
            0f, padTop, 0f, chartBottom,
            Design.PRIMARY_LIGHT, Design.PRIMARY, Shader.TileMode.CLAMP
        )

        values.forEachIndexed { i, v ->
            val ratio = (v / maxValue).coerceIn(0.08f, 1f)
            val barH = chartH * ratio
            val left = padLeft + i * slot + (slot - barW) / 2f
            val top = chartBottom - barH
            canvas.drawRoundRect(RectF(left, top, left + barW, chartBottom), radius, radius, barPaint)
            canvas.drawText(
                labels[i],
                left + barW / 2f,
                height - padBottom - context.sp(3),
                labelPaint
            )
        }
        barPaint.shader = null
    }
}
