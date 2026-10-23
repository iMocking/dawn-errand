package com.dawnerrand.client.ui.screens.paysuccess

import android.content.Intent
import android.os.Bundle
import android.view.View
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.screens.main.MainActivity
import com.dawnerrand.client.ui.screens.ordertrack.OrderTrackActivity
import com.dawnerrand.client.ui.util.padTopWithStatusBar

/** 支付成功：绿对勾 + 订单信息 + 返回首页/查看配送 */
class PaySuccessActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_pay_success)
        window.decorView.padTopWithStatusBar()

        findViewById<TextView>(R.id.tvAmount).text =
            "¥${intent.getStringExtra("amount") ?: "0.0"}"
        findViewById<TextView>(R.id.tvType).text =
            intent.getStringExtra("type") ?: "外卖"
        findViewById<TextView>(R.id.tvOrderNo).text =
            intent.getStringExtra("order_no") ?: "DD20261024"

        findViewById<View>(R.id.btnHome).setOnClickListener {
            startActivity(Intent(this, MainActivity::class.java))
            finishAffinity()
        }
        findViewById<View>(R.id.btnTrack).setOnClickListener {
            startActivity(Intent(this, OrderTrackActivity::class.java))
        }
    }
}
