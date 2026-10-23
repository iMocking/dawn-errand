package com.dawnerrand.driver.ui.screens.orderpool

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
import com.dawnerrand.driver.data.model.OrderItem
import com.dawnerrand.driver.databinding.FragmentOrderPoolBinding
import com.dawnerrand.driver.ui.screens.orderdetail.OrderDetailActivity
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.dp
import com.dawnerrand.driver.ui.util.padTopWithStatusBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 接单大厅：还原 design/driver/order-pool.html（筛选 chips + 订单卡片 + 抢单）。
 */
class OrderPoolFragment : Fragment() {

    private var _binding: FragmentOrderPoolBinding? = null
    private val binding get() = _binding!!
    private val vm: OrderPoolViewModel by viewModels()
    private lateinit var adapter: OrderPoolAdapter
    private val filterViews = mutableListOf<TextView>()

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View {
        _binding = FragmentOrderPoolBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.navBar.padTopWithStatusBar()
        buildFilterChips()

        adapter = OrderPoolAdapter(emptyList()) { order ->
            requireContext().toast("抢单成功")
            OrderDetailActivity.start(requireContext())
        }
        binding.recyclerPool.layoutManager = LinearLayoutManager(requireContext())
        binding.recyclerPool.adapter = adapter

        vm.orders.observe(viewLifecycleOwner, Observer { adapter.submit(it) })
    }

    private fun buildFilterChips() {
        val labels = listOf("全部", "外卖单", "跑腿单", "5公里内")
        filterViews.clear()
        binding.filterBar.removeAllViews()
        labels.forEachIndexed { index, label ->
            val chip = TextView(requireContext()).apply {
                text = label
                textSize = 12f
                setPadding(requireContext().dp(12), requireContext().dp(5), requireContext().dp(12), requireContext().dp(5))
                isClickable = true
                isFocusable = true
            }
            val lp = LinearLayout.LayoutParams(LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            if (index > 0) lp.marginStart = requireContext().dp(8)
            chip.layoutParams = lp
            chip.setOnClickListener { vm.selectFilter(index) }
            binding.filterBar.addView(chip)
            filterViews.add(chip)
        }
        refreshChips(0)

        vm.selectedFilter.observe(viewLifecycleOwner, Observer { refreshChips(it) })
    }

    private fun refreshChips(selected: Int) {
        filterViews.forEachIndexed { index, chip ->
            val active = index == selected
            chip.setBackgroundResource(if (active) R.drawable.bg_chip_active else R.drawable.bg_chip)
            chip.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            chip.setTypeface(chip.typeface, if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL)
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
