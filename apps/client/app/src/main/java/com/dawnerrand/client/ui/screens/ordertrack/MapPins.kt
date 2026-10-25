package com.dawnerrand.client.ui.screens.ordertrack

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BlurMaskFilter
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Paint
import android.graphics.Path
import android.graphics.RectF
import androidx.appcompat.content.res.AppCompatResources
import androidx.core.graphics.drawable.DrawableCompat
import com.amap.api.maps.model.BitmapDescriptor
import com.amap.api.maps.model.BitmapDescriptorFactory
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.sp

/**
 * 地图标记位图工厂（竞品风气泡图钉）：
 * - 取货点：白色圆角标签「商家」+ 橙色气泡图钉 + 白色餐盒；
 * - 送达点：白色圆角标签「送达点」+ 绿色气泡图钉 + 白色房子；
 * - 骑手：橙色圆底 + 白色电动车 + 顶部方向箭头 + 脉冲圈（脉冲由地图 Circle 绘制）。
 */
object MapPins {

    private const val LABEL_HEIGHT_DP = 24
    private const val PIN_SIZE_DP = 48

    fun pickup(ctx: Context): BitmapDescriptor =
        labeledBubble(ctx, R.drawable.bg_marker_pickup, "商家", 0xFFFF6B2C.toInt())

    fun destination(ctx: Context): BitmapDescriptor =
        labeledBubble(ctx, R.drawable.bg_marker_drop, "送达点", 0xFF00B578.toInt())

    fun rider(ctx: Context): BitmapDescriptor = riderBitmap(ctx)

    // ---------------- 气泡图钉 ----------------

    private fun labeledBubble(
        ctx: Context,
        pinRes: Int,
        label: String,
        accent: Int,
    ): BitmapDescriptor {
        val pin = ctx.dp(PIN_SIZE_DP)
        val labelH = ctx.dp(LABEL_HEIGHT_DP)
        val width = ctx.dp(76)
        val height = labelH + pin + ctx.dp(2)
        val bitmap = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)
        val cx = width / 2f

        // 气泡底部投影（让 pin 在浅色地图上浮起来）
        val shadow = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            color = 0x33000000
            maskFilter = BlurMaskFilter(ctx.dp(5).toFloat(), BlurMaskFilter.Blur.NORMAL)
        }
        canvas.drawCircle(cx, labelH + pin * 0.55f, pin * 0.38f, shadow)

        // 标签胶囊（白色圆角 + 细边框 + 品牌色文字）
        val textPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            color = accent
            textSize = ctx.sp(11)
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            textAlign = Paint.Align.CENTER
        }
        val textWidth = textPaint.measureText(label)
        val labelW = textWidth + ctx.dp(22)
        val labelRect = RectF(
            cx - labelW / 2,
            ctx.dp(1).toFloat(),
            cx + labelW / 2,
            labelH - ctx.dp(3).toFloat(),
        )
        val labelBg = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            color = Color.WHITE
            setShadowLayer(ctx.dp(4).toFloat(), 0f, ctx.dp(2).toFloat(), 0x33000000)
        }
        canvas.drawRoundRect(labelRect, labelRect.height() / 2, labelRect.height() / 2, labelBg)
        labelBg.clearShadowLayer()
        val border = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            color = 0x1A000000
            style = Paint.Style.STROKE
            strokeWidth = ctx.dp(1).toFloat()
        }
        canvas.drawRoundRect(labelRect, labelRect.height() / 2, labelRect.height() / 2, border)
        val baseline = labelRect.centerY() - (textPaint.descent() + textPaint.ascent()) / 2
        canvas.drawText(label, cx, baseline, textPaint)

        // 气泡图钉（矢量按 48dp 绘制，尖角指向坐标点）
        val drawable = AppCompatResources.getDrawable(ctx, pinRes)?.mutate()
        if (drawable != null) {
            drawable.setBounds(
                (cx - pin / 2f).toInt(),
                labelH,
                (cx + pin / 2f).toInt(),
                labelH + pin,
            )
            drawable.draw(canvas)
        }
        return BitmapDescriptorFactory.fromBitmap(bitmap)
    }

    // ---------------- 骑手 ----------------

    private fun riderBitmap(ctx: Context): BitmapDescriptor {
        val size = ctx.dp(72)
        val bitmap = Bitmap.createBitmap(size, size, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)
        val cx = size / 2f
        val cy = size * 0.58f
        val radius = size * 0.29f

        // 阴影
        canvas.drawCircle(
            cx, cy + ctx.dp(2), radius * 1.04f,
            Paint(Paint.ANTI_ALIAS_FLAG).apply {
                color = 0x40000000
                maskFilter = BlurMaskFilter(ctx.dp(5).toFloat(), BlurMaskFilter.Blur.NORMAL)
            },
        )
        // 橙色圆底 + 白色描边
        canvas.drawCircle(
            cx, cy, radius,
            Paint(Paint.ANTI_ALIAS_FLAG).apply { color = 0xFFFF6B2C.toInt() },
        )
        canvas.drawCircle(
            cx, cy, radius,
            Paint(Paint.ANTI_ALIAS_FLAG).apply {
                color = Color.WHITE
                style = Paint.Style.STROKE
                strokeWidth = ctx.dp(2).toFloat()
            },
        )
        // 方向箭头：画在圆底内部顶部，随 marker rotateAngle 旋转，避免外置箭头视觉断裂
        val arrow = Path().apply {
            moveTo(cx, cy - radius + ctx.dp(3))
            lineTo(cx - ctx.dp(5), cy - radius + ctx.dp(9))
            lineTo(cx + ctx.dp(5), cy - radius + ctx.dp(9))
            close()
        }
        canvas.drawPath(
            arrow,
            Paint(Paint.ANTI_ALIAS_FLAG).apply {
                color = Color.WHITE
                style = Paint.Style.FILL
            },
        )
        // 白色电动车
        val drawable = AppCompatResources.getDrawable(ctx, R.drawable.ic_tab_errand_scooter)?.mutate()
        if (drawable != null) {
            DrawableCompat.setTint(drawable, Color.WHITE)
            val icon = (radius * 1.15f).toInt()
            val padX = (cx - icon / 2f).toInt()
            val padY = (cy - icon / 2f).toInt()
            drawable.setBounds(padX, padY, padX + icon, padY + icon)
            drawable.draw(canvas)
        }
        return BitmapDescriptorFactory.fromBitmap(bitmap)
    }
}
