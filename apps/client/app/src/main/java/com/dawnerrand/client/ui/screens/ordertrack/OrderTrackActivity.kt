package com.dawnerrand.client.ui.screens.ordertrack

import android.content.ActivityNotFoundException
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.view.View
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.Order
import com.dawnerrand.client.data.OrderStatus
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/** 配送进度：地图示意 + 动态状态卡 + 进度条 + 骑手卡。 */
class OrderTrackActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this
    private var order: Order? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_order_track)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(false)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        window.statusBarColor = android.graphics.Color.argb(0x66, 0, 0, 0)

        val id = intent.getStringExtra(EXTRA_ORDER_ID)
        order = id?.let { DemoRepository.orderById(it) }
            ?: DemoRepository.orders.firstOrNull { it.ongoing }
        render()
    }

    override fun onResume() {
        super.onResume()
        if (order != null) {
            order = DemoRepository.orderById(order!!.id)
            render()
        }
    }

    private fun render() {
        val o = order ?: run {
            ctx.toast("暂无配送中的订单")
            finish()
            return
        }
        val status = when (o.status) {
            OrderStatus.PENDING_PAY -> "订单待支付"
            OrderStatus.PENDING_ACCEPT -> "等待商家接单"
            OrderStatus.DELIVERING -> "骑手正在取货，马上送到"
            OrderStatus.PENDING_REVIEW -> "订单已送达，待评价"
            OrderStatus.DONE -> "订单已送达"
            OrderStatus.CANCELED -> "订单已取消"
            OrderStatus.REFUND -> "售后处理中"
            else -> "订单处理中"
        }
        val eta = SimpleDateFormat("HH:mm", Locale.CHINA)
            .format(Date(o.createdAt + 35 * 60_000L))
        val desc = when (o.status) {
            OrderStatus.PENDING_PAY -> "完成支付后骑手将尽快接单"
            OrderStatus.PENDING_ACCEPT -> "预计 $eta 送达 · 距您 1.2km · 正在分配骑手"
            OrderStatus.DELIVERING -> "预计 $eta 送达 · 距您 1.2km"
            OrderStatus.PENDING_REVIEW, OrderStatus.DONE -> "已于 $eta 送达，感谢使用曙光跑腿"
            OrderStatus.CANCELED -> "订单已取消，如已支付将原路退回"
            OrderStatus.REFUND -> "客服正在处理售后，请耐心等待"
            else -> "预计 $eta 送达"
        }
        findViewById<TextView>(R.id.tvTrackStatus).text = status
        findViewById<TextView>(R.id.tvTrackEta).text = desc

        val delivering = o.status == OrderStatus.DELIVERING
        val finished = o.status == OrderStatus.PENDING_REVIEW || o.status == OrderStatus.DONE
        val waiting = o.status == OrderStatus.PENDING_ACCEPT
        val showRider = delivering || finished
        findViewById<View>(R.id.riderCard).visibility = if (showRider) View.VISIBLE else View.GONE
        findViewById<View>(R.id.btnCallRider).visibility =
            if (delivering || finished) View.VISIBLE else View.GONE
        findViewById<View>(R.id.btnCallRider).setOnClickListener { callRider() }
        findViewById<TextView>(R.id.tvRiderName).text =
            if (o.type == "跑腿") "王师傅 · 鄂J 6K328" else "李师傅 · 鄂J 8D621"
        findViewById<TextView>(R.id.tvRiderMeta).text = "⭐ 4.9  ·  已接 1286 单"
        findViewById<View>(R.id.riderCard).setOnClickListener {
            if (showRider) callRider() else ctx.toast("骑手接单后可查看联系方式")
        }

        val progress1 = findViewById<View>(R.id.progress1)
        val progress2 = findViewById<View>(R.id.progress2)
        progress1.setBackgroundResource(
            if (delivering || finished) R.drawable.bg_progress_done else R.drawable.bg_progress_rest)
        progress2.setBackgroundResource(
            if (finished) R.drawable.bg_progress_done else R.drawable.bg_progress_rest)
        val step1 = findViewById<TextView>(R.id.tvStep1)
        val step2 = findViewById<TextView>(R.id.tvStep2)
        val step3 = findViewById<TextView>(R.id.tvStep3)
        step1.setTextColor(
            if (delivering || finished || waiting) Design.SUCCESS else Design.TEXT_FAINT)
        step2.setTextColor(
            if (delivering || finished) Design.SUCCESS else Design.TEXT_FAINT)
        step3.setTextColor(if (finished) Design.SUCCESS else Design.TEXT_FAINT)
        step1.text = if (waiting) "等待接单" else "商家接单"
        if (waiting) {
            step1.setTextColor(Design.PRIMARY)
        }
    }

    private fun callRider() {
        try {
            startActivity(Intent(Intent.ACTION_DIAL, Uri.parse("tel:13800001234")))
        } catch (e: ActivityNotFoundException) {
            ctx.toast("骑手电话：138-0000-1234")
        }
    }

    companion object {
        const val EXTRA_ORDER_ID = "order_id"
    }
}
