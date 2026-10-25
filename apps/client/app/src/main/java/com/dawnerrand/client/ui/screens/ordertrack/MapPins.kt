package com.dawnerrand.client.ui.screens.ordertrack

import android.content.Context
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Paint
import androidx.appcompat.content.res.AppCompatResources
import androidx.core.graphics.drawable.DrawableCompat
import com.amap.api.maps.model.BitmapDescriptor
import com.amap.api.maps.model.BitmapDescriptorFactory
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.util.dp

/**
 * 地图标记位图工厂：按设计稿绘制 取货点(绿·购物袋) / 骑手(橙·电动车) / 送达点(红·家)。
 * 运行时由矢量图标 + 圆底绘制，避免额外 PNG 资源。
 */
object MapPins {

    /** 取货点：绿色圆 + 购物袋（商家）。 */
    fun pickup(ctx: Context): BitmapDescriptor =
        BitmapDescriptorFactory.fromBitmap(circle(ctx, 0xFF00B578.toInt(), R.drawable.ic_food_bag, 40))

    /** 骑手：橙色圆 + 电动车（随航向旋转）。 */
    fun rider(ctx: Context): BitmapDescriptor =
        BitmapDescriptorFactory.fromBitmap(circle(ctx, 0xFFFF6B2C.toInt(), R.drawable.ic_tab_errand_scooter, 48))

    /** 送达点：红色圆 + 家（客户地址）。 */
    fun destination(ctx: Context): BitmapDescriptor =
        BitmapDescriptorFactory.fromBitmap(circle(ctx, 0xFFFF4D4F.toInt(), R.drawable.ic_home, 40))

    private fun circle(ctx: Context, bgColor: Int, iconRes: Int, sizeDp: Int): Bitmap {
        val size = ctx.dp(sizeDp)
        val bitmap = Bitmap.createBitmap(size, size, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)
        val radius = size / 2f

        // 投影
        canvas.drawCircle(radius + size * 0.05f, radius + size * 0.06f, radius * 0.88f,
            Paint(Paint.ANTI_ALIAS_FLAG).apply { color = 0x33000000 })

        // 底色圆
        canvas.drawCircle(radius, radius, radius * 0.92f,
            Paint(Paint.ANTI_ALIAS_FLAG).apply { color = bgColor })

        // 白色描边
        canvas.drawCircle(radius, radius, radius * 0.86f,
            Paint(Paint.ANTI_ALIAS_FLAG).apply {
                style = Paint.Style.STROKE
                strokeWidth = ctx.dp(2).toFloat()
                color = Color.WHITE
            })

        // 居中白色图标
        val drawable = AppCompatResources.getDrawable(ctx, iconRes)?.mutate()
        if (drawable != null) {
            DrawableCompat.setTint(drawable, Color.WHITE)
            val iconSize = (size * 0.5f).toInt()
            val pad = (size - iconSize) / 2
            drawable.setBounds(pad, pad, pad + iconSize, pad + iconSize)
            drawable.draw(canvas)
        }
        return bitmap
    }
}
