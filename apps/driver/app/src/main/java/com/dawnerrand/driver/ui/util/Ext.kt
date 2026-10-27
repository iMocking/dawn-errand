package com.dawnerrand.driver.ui.util

import android.content.Context
import android.view.View
import android.widget.Toast
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import com.dawnerrand.driver.data.api.ApiClient
import kotlin.math.roundToInt

/** 设计稿 1px = 1dp，此方法把设计值换算为当前设备像素 */
fun Context.dp(v: Int): Int = (v * resources.displayMetrics.density).roundToInt()

fun Context.sp(v: Int): Float = v * resources.displayMetrics.scaledDensity

fun Context.toast(msg: String) {
    Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
}

/**
 * 把骑手端接口的失败翻成骑手能直接照做的话。
 * 后端 message 本身是可读中文，但少数场景需要更具体的指引（例如 1801 要提示「手慢了」）。
 */
fun ApiClient.Resp.driverMessage(): String = when (code) {
    1801 -> "手慢了，订单已被其他骑手抢走"
    1802 -> "请先上线再接单"
    1803 -> "订单不在你的服务区域"
    1804 -> "取送码不正确，请与商家或顾客核对"
    1607 -> message.ifBlank { "订单状态已变化，请刷新后重试" }
    1006 -> "登录已过期，请重新登录"
    else -> message.ifBlank { "操作失败，请稍后重试" }
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
