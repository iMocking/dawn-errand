package com.dawnerrand.store.ui.util

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.net.Uri
import android.view.View
import android.widget.Toast
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import kotlin.math.roundToInt

/** 设计稿 1px = 1dp，此方法把设计值换算为当前设备像素 */
fun Context.dp(v: Int): Int = (v * resources.displayMetrics.density).roundToInt()

fun Context.sp(v: Int): Float = v * resources.displayMetrics.scaledDensity

fun Context.toast(msg: String) {
    Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
}

/**
 * 状态栏图标明暗：浅色页面（白底导航栏）用深色图标，渐变/橙色头部页面用白色图标。
 * 同时把状态栏底色补成与页面头部一致的白色/品牌橙，避免默认窗口底色透出灰条。
 */
fun Activity.setLightStatusBars(light: Boolean) {
    WindowCompat.getInsetsController(window, window.decorView).isAppearanceLightStatusBars = light
    window.statusBarColor = if (light) Color.WHITE else Color.parseColor("#FF8A3D")
}

/**
 * 打开系统拨号盘。
 * 演示订单中的手机号带掩码（138****8888）无法直接拨打，统一回退到测试号码；
 * 真实号码（含 400 热线等 7 位以上号码）原样拨打。
 */
fun Context.dial(phone: String, fallback: String = "13800008888") {
    val digits = phone.filter { it.isDigit() }
    val target = when {
        phone.contains("*") -> fallback
        digits.length >= 7 -> digits
        else -> fallback
    }
    try {
        startActivity(Intent(Intent.ACTION_DIAL, Uri.parse("tel:$target")))
    } catch (e: Exception) {
        toast("无法打开拨号盘")
    }
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
