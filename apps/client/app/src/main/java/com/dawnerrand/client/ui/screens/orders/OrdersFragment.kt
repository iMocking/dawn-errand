package com.dawnerrand.client.ui.screens.orders

import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.Order
import com.dawnerrand.client.data.OrderStatus
import com.dawnerrand.client.ui.screens.merchant.MerchantDetailActivity
import com.dawnerrand.client.ui.screens.orderdetail.OrderDetailActivity
import com.dawnerrand.client.ui.screens.ordertrack.OrderTrackActivity
import com.dawnerrand.client.ui.screens.paysuccess.PaySuccessActivity
import com.dawnerrand.client.ui.screens.review.ReviewActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.confirmDialog
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/**
 * 我的订单（主页 Tab 内容 / 独立页共用）：
 * 4 筛选 Tab（进行中/已完成/已取消/退款售后）+ 订单卡列表 + 完整操作。
 */
class OrdersFragment : Fragment() {
    private val ctx: android.content.Context get() = requireContext()

    private val filters = listOf("进行中", "已完成", "已取消", "退款售后")
    private val tabs = mutableListOf<LinearLayout>()
    private val indicators = mutableListOf<View>()
    private var current = 0
    private var filterStatus: String? = null
    private var initializedView = false
    private var filterTouched = false

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?,
    ): View {
        val root = inflater.inflate(R.layout.activity_orders, container, false)
        val activity = requireActivity() as AppCompatActivity
        val nav = root.findViewById<View>(R.id.navBar)
        if (embedded()) {
            nav.padTopWithStatusBar()
            nav.setPadding(ctx.dp(16), nav.paddingTop, ctx.dp(16), nav.paddingBottom)
            root.findViewById<View>(R.id.btnBack).visibility = View.GONE
        } else {
            activity.setLightStatusBars(true)
            activity.window.decorView.padTopWithStatusBar()
            root.findViewById<View>(R.id.btnBack).setOnClickListener { activity.finish() }
        }

        if (!filterTouched) {
            filterStatus = arguments?.getString(ARG_FILTER_STATUS)
            current = arguments?.getInt(ARG_TAB, 0) ?: 0
            filterStatus?.let { status ->
                current = when (status) {
                    OrderStatus.PENDING_PAY, OrderStatus.PENDING_ACCEPT, OrderStatus.DELIVERING -> 0
                    OrderStatus.PENDING_REVIEW, OrderStatus.DONE -> 1
                    OrderStatus.CANCELED -> 2
                    OrderStatus.REFUND -> 3
                    else -> current
                }
            }
        }

        tabs.clear()
        indicators.clear()
        val tabRow = root.findViewById<LinearLayout>(R.id.filterTabs)
        tabRow.removeAllViews()
        filters.forEachIndexed { i, name ->
            tabs.add(buildTab(i, name))
            tabRow.addView(tabs[i])
        }
        initializedView = true
        renderTabs()
        renderList()
        return root
    }

    override fun onResume() {
        super.onResume()
        if (initializedView) {
            renderTabs()
            renderList()
        }
    }

    override fun onHiddenChanged(hidden: Boolean) {
        super.onHiddenChanged(hidden)
        if (!hidden && initializedView) {
            renderTabs()
            renderList()
        }
    }

    override fun onDestroyView() {
        initializedView = false
        tabs.clear()
        indicators.clear()
        super.onDestroyView()
    }

    /** 从「我的」订单状态入口切换过来时按状态过滤 */
    fun showStatus(status: String?) {
        filterTouched = true
        filterStatus = status
        current = when (status) {
            OrderStatus.PENDING_PAY, OrderStatus.PENDING_ACCEPT, OrderStatus.DELIVERING -> 0
            OrderStatus.PENDING_REVIEW, OrderStatus.DONE -> 1
            OrderStatus.CANCELED -> 2
            OrderStatus.REFUND -> 3
            else -> 0
        }
        if (initializedView) {
            renderTabs()
            renderList()
        }
    }

    /** 直接点击底部「订单」Tab 时恢复默认「进行中」 */
    fun resetFilter() {
        filterTouched = true
        filterStatus = null
        current = 0
        if (initializedView) {
            renderTabs()
            renderList()
        }
    }

    private fun embedded(): Boolean = arguments?.getBoolean(ARG_EMBEDDED, false) ?: false

    private fun buildTab(index: Int, name: String): LinearLayout {
        val tab = LinearLayout(ctx).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.MATCH_PARENT, 1f)
            isClickable = true
            setOnClickListener {
                current = index
                filterStatus = null
                renderTabs()
                renderList()
            }
        }
        tab.addView(TextView(ctx).apply {
            text = name
            textSize = 14f
            gravity = Gravity.CENTER
            setPadding(0, ctx.dp(11), 0, ctx.dp(7))
        })
        val indicator = View(ctx).apply {
            setBackgroundColor(Design.PRIMARY)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(20), ctx.dp(2)).apply {
                gravity = Gravity.CENTER_HORIZONTAL
            }
        }
        tab.addView(indicator)
        indicators.add(indicator)
        return tab
    }

    private fun renderTabs() {
        tabs.forEachIndexed { i, tab ->
            val tv = tab.getChildAt(0) as TextView
            val active = i == current
            tv.text = "${filters[i]} ${DemoRepository.orderCountForTab(i)}"
            tv.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            tv.typeface = android.graphics.Typeface.create(
                android.graphics.Typeface.DEFAULT,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
            indicators.getOrNull(i)?.visibility = if (active) View.VISIBLE else View.INVISIBLE
        }
    }

    private fun renderList() {
        val list = view?.findViewById<LinearLayout>(R.id.list) ?: return
        list.removeAllViews()
        val shown = filterStatus?.let { status -> DemoRepository.orders.filter { it.status == status } }
            ?: DemoRepository.ordersForTab(current)
        if (shown.isEmpty()) {
            list.addView(emptyView())
            return
        }
        shown.forEach { list.addView(buildOrderCard(it)) }
    }

    private fun emptyView(): LinearLayout {
        val box = LinearLayout(ctx).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(280))
        }
        box.addView(ImageView(ctx).apply {
            setImageResource(R.drawable.ic_package)
            setColorFilter(Design.TEXT_FAINT)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(48), ctx.dp(48))
        })
        box.addView(TextView(ctx).apply {
            text = "暂无相关订单"
            setTextColor(Design.TEXT_FAINT)
            textSize = 14f
            setPadding(0, ctx.dp(12), 0, 0)
        })
        box.addView(TextView(ctx).apply {
            text = "去首页逛逛，下一单马上就来"
            setTextColor(Design.TEXT_FAINT)
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        return box
    }

    private fun buildOrderCard(o: Order): LinearLayout {
        val card = LinearLayout(ctx).apply {
            orientation = LinearLayout.VERTICAL
            background = ContextCompat.getDrawable(ctx, R.drawable.bg_card)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.bottomMargin = ctx.dp(10)
            }
            isClickable = true
            setOnClickListener { openDetail(o) }
        }
        val head = LinearLayout(ctx).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(ctx.dp(16), ctx.dp(14), ctx.dp(16), ctx.dp(10))
        }
        head.addView(TextView(ctx).apply {
            text = o.shop
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        })
        head.addView(TextView(ctx).apply {
            text = o.status
            setTextColor(statusColor(o.status))
            textSize = 12f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        card.addView(head)

        val body = LinearLayout(ctx).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(ctx.dp(16), 0, ctx.dp(16), 0)
        }
        val img = LinearLayout(ctx).apply {
            gravity = Gravity.CENTER
            background = android.graphics.drawable.ColorDrawable(o.iconBg)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(56), ctx.dp(56))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(12)
        }
        img.addView(ImageView(ctx).apply {
            setImageResource(o.iconRes)
            setColorFilter(o.iconColor)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(26), ctx.dp(26))
        })
        body.addView(img)

        val info = LinearLayout(ctx).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        info.addView(TextView(ctx).apply {
            text = o.items
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 14f
            maxLines = 2
        })
        info.addView(TextView(ctx).apply {
            text = o.time
            setTextColor(Design.TEXT_TERTIARY)
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        body.addView(info)
        body.addView(TextView(ctx).apply {
            text = o.total
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 16f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT).apply {
                leftMargin = ctx.dp(8)
            }
        })
        card.addView(body)

        val actions = LinearLayout(ctx).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.END
            setPadding(ctx.dp(16), ctx.dp(12), ctx.dp(16), ctx.dp(14))
        }
        val actionList = DemoRepository.actionsFor(o)
        actionList.forEachIndexed { i, action ->
            val primary = isPrimary(action)
            val btn = TextView(ctx).apply {
                text = action
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(ctx.dp(18), ctx.dp(8), ctx.dp(18), ctx.dp(8))
                background = ContextCompat.getDrawable(
                    ctx,
                    if (primary) R.drawable.bg_btn else R.drawable.bg_btn_outline
                )
                setTextColor(if (primary) Design.WHITE else Design.TEXT_SECONDARY)
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
                if (i > 0) (layoutParams as LinearLayout.LayoutParams).leftMargin = ctx.dp(10)
                isClickable = true
                setOnClickListener { onAction(o, action) }
            }
            actions.addView(btn)
        }
        card.addView(actions)
        return card
    }

    private fun statusColor(status: String): Int = when (status) {
        OrderStatus.PENDING_PAY, OrderStatus.PENDING_ACCEPT, OrderStatus.DELIVERING -> Design.PRIMARY
        OrderStatus.PENDING_REVIEW -> Design.ORANGE_DEEP
        OrderStatus.REFUND -> Color.rgb(0x2B, 0x7F, 0xFF)
        else -> Design.TEXT_TERTIARY
    }

    private fun isPrimary(action: String): Boolean =
        action in listOf("查看配送", "去支付", "去评价", "确认收货", "查看进度")

    private fun openDetail(o: Order) {
        startActivity(Intent(requireContext(), OrderDetailActivity::class.java)
            .putExtra(OrderDetailActivity.EXTRA_ORDER_ID, o.id))
    }

    private fun onAction(o: Order, action: String) {
        when (action) {
            "取消订单" -> ctx.confirmDialog("取消订单", "确定取消「${o.shop}」的订单吗？") {
                DemoRepository.cancelOrder(o.id)
                renderTabs()
                renderList()
                ctx.toast("订单已取消")
            }
            "去支付" -> {
                DemoRepository.payOrder(o.id)
                startActivity(Intent(requireContext(), PaySuccessActivity::class.java)
                    .putExtra("amount", o.total.removePrefix("¥"))
                    .putExtra("type", o.type)
                    .putExtra("order_no", o.orderNo)
                    .putExtra("order_id", o.id))
                renderTabs()
                renderList()
            }
            "查看配送" -> startActivity(Intent(requireContext(), OrderTrackActivity::class.java)
                .putExtra(OrderTrackActivity.EXTRA_ORDER_ID, o.id))
            "确认收货" -> ctx.confirmDialog("确认收货", "请确认已收到商品，确认后订单将进入待评价状态。") {
                DemoRepository.confirmReceipt(o.id)
                renderTabs()
                renderList()
                ctx.toast("已确认收货，快去评价吧")
            }
            "去评价", "评价" -> startActivity(Intent(requireContext(), ReviewActivity::class.java)
                .putExtra(ReviewActivity.EXTRA_ORDER_ID, o.id))
            "再来一单" -> reorder(o)
            "删除订单" -> ctx.confirmDialog("删除订单", "删除后订单记录不可恢复，确定删除吗？") {
                DemoRepository.deleteOrder(o.id)
                renderTabs()
                renderList()
                ctx.toast("订单已删除")
            }
            "查看进度" -> openDetail(o)
            else -> openDetail(o)
        }
    }

    private fun reorder(o: Order) {
        if (o.type == "跑腿") {
            (activity as? com.dawnerrand.client.ui.screens.main.MainActivity)?.selectTab(1)
                ?: startActivity(Intent(requireContext(), com.dawnerrand.client.ui.screens.errand.ErrandActivity::class.java))
            return
        }
        if (DemoRepository.reorder(o)) {
            startActivity(Intent(requireContext(), MerchantDetailActivity::class.java)
                .putExtra("merchant_id", o.merchantId ?: "m1"))
            ctx.toast("已把商品加入购物车")
        } else {
            ctx.toast("商品已下架，去商家看看其他选择吧")
        }
    }

    companion object {
        private const val ARG_EMBEDDED = "embedded"
        private const val ARG_FILTER_STATUS = "filter_status"
        private const val ARG_TAB = "tab_index"

        fun newInstance(
            embedded: Boolean,
            filterStatus: String? = null,
            tab: Int = 0,
        ): OrdersFragment = OrdersFragment().apply {
            arguments = Bundle().apply {
                putBoolean(ARG_EMBEDDED, embedded)
                putString(ARG_FILTER_STATUS, filterStatus)
                putInt(ARG_TAB, tab)
            }
        }
    }
}
