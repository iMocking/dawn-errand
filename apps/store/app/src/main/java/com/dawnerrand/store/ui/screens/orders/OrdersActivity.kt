package com.dawnerrand.store.ui.screens.orders

import android.content.Intent
import android.os.Bundle
import android.view.View
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.R
import com.dawnerrand.store.data.StoreOrder
import com.dawnerrand.store.databinding.ActivityOrdersBinding
import com.dawnerrand.store.ui.adapter.OrderAdapter
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast

/**
 * 订单管理：四个状态 Tab + 订单卡片列表（orders 稿）。
 * 支持接单/拒单/通知出餐，操作后共享状态联动刷新。
 */
class OrdersActivity : AppCompatActivity() {

    private lateinit var b: ActivityOrdersBinding
    private val vm: OrdersViewModel by viewModels()

    private val tabs = listOf(
        Triple(R.id.tabPending, R.id.tabPendingText, R.id.tabPendingLine),
        Triple(R.id.tabCooking, R.id.tabCookingText, R.id.tabCookingLine),
        Triple(R.id.tabDelivering, R.id.tabDeliveringText, R.id.tabDeliveringLine),
        Triple(R.id.tabDone, R.id.tabDoneText, R.id.tabDoneLine),
    )
    private val counts = listOf(R.id.tabPendingCount, R.id.tabCookingCount, -1, -1)

    private val adapter = OrderAdapter(
        onAction = { order, action -> handleAction(order, action) },
        onClick = { order -> openDetail(order) },
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityOrdersBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.orderList.layoutManager = LinearLayoutManager(this)
        b.orderList.adapter = adapter

        tabs.forEachIndexed { index, _ ->
            findViewById<View>(tabs[index].first).setOnClickListener { vm.selectTab(index) }
        }

        vm.currentTab.observe(this) { refreshTabs() }
        vm.filtered.observe(this) { adapter.submit(it) }

        // 各 Tab 角标（待接单/制作中）
        com.dawnerrand.store.data.StoreState.ordersLive.observe(this) { orders ->
            val pending = orders.count { it.status == com.dawnerrand.store.data.ORDER_PENDING }
            val cooking = orders.count { it.status == com.dawnerrand.store.data.ORDER_COOKING }
            findViewById<TextView>(counts[0]).apply {
                text = pending.toString()
                visibility = if (pending > 0) View.VISIBLE else View.GONE
            }
            findViewById<TextView>(counts[1]).apply {
                text = cooking.toString()
                visibility = if (cooking > 0) View.VISIBLE else View.GONE
            }
            refreshTabs()
        }

        val initialTab = intent.getIntExtra("tab", 0)
        vm.selectTab(initialTab)
    }

    private fun refreshTabs() {
        val current = vm.currentTab.value ?: 0
        tabs.forEachIndexed { index, tab ->
            val text = findViewById<TextView>(tab.second)
            val line = findViewById<View>(tab.third)
            val active = index == current
            text.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            text.setTypeface(
                text.typeface,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
            line.visibility = if (active) View.VISIBLE else View.INVISIBLE
        }
    }

    private fun handleAction(order: StoreOrder, action: String) {
        when (action) {
            "接单" -> {
                vm.accept(order.id)
                toast("已接单，请尽快制作")
            }
            "拒单" -> {
                vm.reject(order.id)
                toast("已拒单")
            }
            "通知出餐" -> {
                vm.notifyReady(order.id)
                toast("已通知骑手取餐")
            }
            "查看配送" -> openDetail(order)
            "再来一单" -> openDetail(order)
            else -> openDetail(order)
        }
    }

    private fun openDetail(order: StoreOrder) {
        startActivity(
            Intent(this, OrderDetailActivity::class.java)
                .putExtra("orderId", order.id)
        )
    }
}
