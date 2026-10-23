package com.dawnerrand.driver.ui.screens.orders

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.LinearLayout
import android.widget.TextView
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.Observer
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.driver.R
import com.dawnerrand.driver.databinding.FragmentOrdersBinding
import com.dawnerrand.driver.ui.screens.orderdetail.OrderDetailActivity
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.dp
import com.dawnerrand.driver.ui.util.padTopWithStatusBar

/**
 * 我的订单：还原 design/driver/orders.html（筛选 Tab + 历史订单卡片）。
 */
class OrdersFragment : Fragment() {

    private var _binding: FragmentOrdersBinding? = null
    private val binding get() = _binding!!
    private val vm: OrdersViewModel by viewModels()
    private lateinit var adapter: OrdersAdapter
    private val tabViews = mutableListOf<LinearLayout>()

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View {
        _binding = FragmentOrdersBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.navBar.padTopWithStatusBar()
        buildTabs()

        adapter = OrdersAdapter(emptyList()) {
            OrderDetailActivity.start(requireContext())
        }
        binding.recyclerOrders.layoutManager = LinearLayoutManager(requireContext())
        binding.recyclerOrders.adapter = adapter

        vm.records.observe(viewLifecycleOwner, Observer { adapter.submit(it) })
    }

    private fun buildTabs() {
        val labels = listOf("全部", "配送中", "已完成", "已取消")
        tabViews.clear()
        binding.filterTabs.removeAllViews()
        labels.forEachIndexed { index, label ->
            val tab = LinearLayout(requireContext()).apply {
                orientation = LinearLayout.VERTICAL
                gravity = android.view.Gravity.CENTER
                setPadding(0, requireContext().dp(12), 0, 0)
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
                isClickable = true
                isFocusable = true
            }
            val text = TextView(requireContext()).apply {
                setText(label)
                textSize = 14f
                gravity = android.view.Gravity.CENTER
            }
            val underline = View(requireContext()).apply {
                layoutParams = LinearLayout.LayoutParams(requireContext().dp(24), requireContext().dp(3))
                setBackgroundResource(R.drawable.bg_tab_indicator)
            }
            tab.addView(text)
            tab.addView(underline)
            tab.setOnClickListener { vm.selectTab(index) }
            binding.filterTabs.addView(tab)
            tabViews.add(tab)
        }
        refreshTabs(0)

        vm.selectedTab.observe(viewLifecycleOwner, Observer { refreshTabs(it) })
    }

    private fun refreshTabs(selected: Int) {
        tabViews.forEachIndexed { index, tab ->
            val active = index == selected
            val text = (tab.getChildAt(0) as TextView)
            val underline = tab.getChildAt(1)
            text.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            text.setTypeface(text.typeface, if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL)
            underline.visibility = if (active) View.VISIBLE else View.INVISIBLE
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
