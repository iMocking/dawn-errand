package com.dawnerrand.store.ui.screens.dashboard

import android.content.Intent
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.FrameLayout
import android.widget.GridLayout
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import com.dawnerrand.store.R
import com.dawnerrand.store.data.GridItem
import com.dawnerrand.store.databinding.FragmentDashboardBinding
import com.dawnerrand.store.ui.screens.business.BusinessActivity
import com.dawnerrand.store.ui.screens.finance.FinanceActivity
import com.dawnerrand.store.ui.screens.menu.MenuActivity
import com.dawnerrand.store.ui.screens.orders.OrdersActivity
import com.dawnerrand.store.ui.screens.statistics.StatisticsActivity
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.padTopWithStatusBar

/**
 * 工作台：今日数据 + 待接单提醒 + 订单/店铺功能宫格（dashboard 稿）。
 * 角标数量来自共享订单状态，接单后自动刷新。
 */
class DashboardFragment : Fragment() {

    private var _b: FragmentDashboardBinding? = null
    private val b get() = _b!!
    private val vm: DashboardViewModel by viewModels()

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View {
        _b = FragmentDashboardBinding.inflate(inflater, container, false)
        return b.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        b.topBox.padTopWithStatusBar()

        // 店铺信息与今日数据
        b.shopName.text = vm.shop.name
        b.shopStatus.text = vm.shop.statusText
        b.todayRevenue.text = vm.shop.todayRevenue
        b.todayOrders.text = vm.shop.todayOrders
        b.goodRate.text = vm.shop.goodRate

        // 待接单提醒
        b.urgentBar.setOnClickListener { openOrders() }
        b.urgentBtn.setOnClickListener { openOrders() }

        // 功能宫格
        vm.pendingCount.observe(viewLifecycleOwner) {
            refreshGrids()
            val pending = it ?: 0
            b.urgentTitle.text = "${pending}笔新订单待接单"
            b.urgentBar.visibility = if (pending > 0) View.VISIBLE else View.GONE
        }
        vm.cookingCount.observe(viewLifecycleOwner) { refreshGrids() }
        refreshGrids()
    }

    private fun refreshGrids() {
        buildGrid(b.orderGrid, vm.gridWithBadges()) { label ->
            when (label) {
                "待接单" -> openOrders(0)
                "制作中" -> openOrders(1)
                "配送中" -> openOrders(2)
                "已完成" -> openOrders(3)
            }
        }
        buildGrid(b.shopGrid, vm.shopGrid) { label ->
            when (label) {
                "菜品管理" -> startActivity(Intent(requireContext(), MenuActivity::class.java))
                "营业设置" -> startActivity(Intent(requireContext(), BusinessActivity::class.java))
                "收入提现" -> startActivity(Intent(requireContext(), FinanceActivity::class.java))
                "经营数据" -> startActivity(Intent(requireContext(), StatisticsActivity::class.java))
            }
        }
    }

    private fun buildGrid(grid: GridLayout, items: List<GridItem>, onClick: (String) -> Unit) {
        val ctx = requireContext()
        grid.removeAllViews()
        items.forEach { item ->
            val cell = LinearLayout(ctx).apply {
                orientation = LinearLayout.VERTICAL
                gravity = android.view.Gravity.CENTER
                layoutParams = GridLayout.LayoutParams().apply {
                    width = 0
                    height = GridLayout.LayoutParams.WRAP_CONTENT
                    columnSpec = GridLayout.spec(GridLayout.UNDEFINED, 1f)
                    setMargins(0, 0, 0, ctx.dp(8))
                }
                isClickable = true
                setOnClickListener { onClick(item.label) }
            }

            val iconBox = FrameLayout(ctx).apply {
                layoutParams = LinearLayout.LayoutParams(ctx.dp(44), ctx.dp(44))
                setBackgroundResource(item.bgRes)
            }
            val icon = ImageView(ctx).apply {
                setImageResource(item.iconRes)
                setColorFilter(item.iconColor)
                layoutParams = FrameLayout.LayoutParams(ctx.dp(22), ctx.dp(22), android.view.Gravity.CENTER)
            }
            iconBox.addView(icon)

            if (item.badge > 0) {
                val badge = TextView(ctx).apply {
                    text = item.badge.toString()
                    textSize = 10f
                    setTextColor(Design.WHITE)
                    setTypeface(typeface, android.graphics.Typeface.BOLD)
                    gravity = android.view.Gravity.CENTER
                    setBackgroundResource(R.drawable.bg_count)
                    setPadding(ctx.dp(4), 0, ctx.dp(4), 0)
                    layoutParams = FrameLayout.LayoutParams(
                        FrameLayout.LayoutParams.WRAP_CONTENT,
                        ctx.dp(16),
                        android.view.Gravity.TOP or android.view.Gravity.END
                    ).apply {
                        marginEnd = ctx.dp(4)
                        topMargin = ctx.dp(-2)
                    }
                }
                iconBox.addView(badge)
            }

            val label = TextView(ctx).apply {
                text = item.label
                textSize = 12f
                setTextColor(Design.TEXT_SECONDARY)
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT
                ).apply { topMargin = ctx.dp(6) }
            }

            cell.addView(iconBox)
            cell.addView(label)
            grid.addView(cell)
        }
    }

    private fun openOrders(tab: Int = 0) {
        startActivity(Intent(requireContext(), OrdersActivity::class.java).putExtra("tab", tab))
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _b = null
    }
}
