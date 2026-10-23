package com.dawnerrand.store.ui.util

import android.content.Context
import android.view.View
import android.widget.Toast
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import kotlin.math.roundToInt

/** 设计稿 1px = 1dp，此方法把设计值换算为当前设备像素 */
fun Context.dp(v: Int): Int = (v * resources.displayMetrics.density).roundToInt()

fun Context.sp(v: Int): Float = v * resources.displayMetrics.scaledDensity

fun Context.toast(msg: String) {
    Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
}

/**
 * 白底页面：让白色导航栏延伸到系统状态栏之后，并把状态栏高度补进导航栏 padding。
 */
fun View.padTopWithStatusBar() {
    var applied = false
    ViewCompat.setOnApplyWindowInsetsListener(this) { v, insets ->
        if (!applied) {
            val top = insets.getInsets(WindowInsetsCompat.Type.statusBars()).top
            v.setPadding(v.paddingLeft, v.paddingTop + top, v.paddingRight, v.paddingBottom)
            applied = true
        }
        insets
    }
}

/**
 * 底部操作栏：把系统导航栏高度补进 paddingBottom。
 */
fun View.padBottomWithNavBar(extra: Int = 0) {
    var applied = false
    ViewCompat.setOnApplyWindowInsetsListener(this) { v, insets ->
        if (!applied) {
            val bottom = insets.getInsets(WindowInsetsCompat.Type.navigationBars()).bottom
            v.setPadding(v.paddingLeft, v.paddingTop, v.paddingRight, v.paddingBottom + bottom + extra)
            applied = true
        }
        insets
    }
}
