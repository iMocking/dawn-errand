package com.dawnerrand.client.ui.screens.paysuccess

import android.content.Intent
import android.os.Bundle
import android.view.View
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.main.MainActivity
import com.dawnerrand.client.ui.screens.ordertrack.OrderTrackActivity
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars

/** 支付成功：绿对勾 + 订单信息 + 返回首页/查看配送。 */
class PaySuccessActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_pay_success)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)

        findViewById<TextView>(R.id.tvAmount).text =
            "¥${intent.getStringExtra("amount") ?: "0.0"}"
        findViewById<TextView>(R.id.tvType).text =
            intent.getStringExtra("type") ?: "外卖"
        findViewById<TextView>(R.id.tvOrderNo).text =
            intent.getStringExtra("order_no") ?: "DD20261024"
        val order = intent.getStringExtra("order_id")?.let { DemoRepository.orderById(it) }
        findViewById<TextView>(R.id.tvAddress).text = order?.address
            ?.substringAfter("· ", order.address)
            ?.takeIf { it.isNotBlank() }
            ?: "王家湾村三组26号"

        findViewById<View>(R.id.btnHome).setOnClickListener {
            startActivity(Intent(this, MainActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
            })
            finish()
        }
        findViewById<View>(R.id.btnTrack).setOnClickListener {
            val id = intent.getStringExtra("order_id")
            startActivity(Intent(this, OrderTrackActivity::class.java).apply {
                if (id != null) putExtra(OrderTrackActivity.EXTRA_ORDER_ID, id)
            })
            finish()
        }
    }
}
