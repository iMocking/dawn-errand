package com.dawnerrand.client.ui.util

import android.content.Context
import android.view.View
import android.widget.Toast
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import kotlin.math.roundToInt

/** 璁捐绋?1px = 1dp锛屾鏂规硶鎶婅璁″€兼崲绠椾负褰撳墠璁惧鍍忕礌 */
fun Context.dp(v: Int): Int = (v * resources.displayMetrics.density).roundToInt()

fun Context.sp(v: Int): Float = v * resources.displayMetrics.scaledDensity

fun Context.toast(msg: String) {
    Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
}

/**
 * 鐧藉簳椤甸潰锛氳鐧借壊瀵艰埅鏍忓欢浼稿埌绯荤粺鐘舵€佹爮涔嬪悗锛屽苟鎶婄姸鎬佹爮楂樺害琛ヨ繘瀵艰埅鏍?padding銆? */
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
 * 搴曢儴鎿嶄綔鏍忥細鎶婄郴缁熷鑸爮楂樺害琛ヨ繘 paddingBottom銆? */
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
