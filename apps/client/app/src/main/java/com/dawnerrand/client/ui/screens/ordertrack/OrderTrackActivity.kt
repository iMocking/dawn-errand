package com.dawnerrand.client.ui.screens.ordertrack

import android.os.Bundle
import android.view.View
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 配送进度：地图示意 + 状态卡 + 骑手卡 */
class OrderTrackActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_order_track)
        window.decorView.padTopWithStatusBar()
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
    }
}
