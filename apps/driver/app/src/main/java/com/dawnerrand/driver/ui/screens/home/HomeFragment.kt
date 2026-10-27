package com.dawnerrand.driver.ui.screens.home

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import com.dawnerrand.driver.data.repository.DriverRemoteRepository
import com.dawnerrand.driver.databinding.FragmentHomeBinding
import com.dawnerrand.driver.ui.screens.navigation.NavigationActivity
import com.dawnerrand.driver.ui.screens.orderdetail.OrderDetailActivity
import com.dawnerrand.driver.ui.util.toast
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

/**
 * 工作台：还原 design/driver/home.html（顶部渐变 + 今日收入环 + 上线开关 + 进行中订单 + 导航按钮）。
 */
class HomeFragment : Fragment() {

    private var _binding: FragmentHomeBinding? = null
    private val binding get() = _binding!!
    private val vm: HomeViewModel by viewModels()

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View {
        _binding = FragmentHomeBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.tvRiderName.text = vm.rider.name
        binding.tvRiderTag.text = vm.rider.level
        binding.tvTodayAmount.text = vm.rider.todayAmount
        binding.tvTodayOrders.text = vm.rider.todayOrders.toString()
        binding.tvTodayScore.text = vm.rider.score
        binding.ringToday.progress = 0.65f

        bindActiveOrder()

        binding.switchOnline.isOn = vm.isOnline
        binding.switchOnline.onCheckedChange = { on ->
            vm.setOnline(on)
            binding.tvOnlineSub.text = if (on) "● 正在接单 · 绿杨乡片区" else "● 已下线 · 点击上线接单"
        }
        vm.statusError.observe(viewLifecycleOwner) { msg ->
            if (msg != null) requireContext().toast(msg)
        }

        binding.btnGoNav.setOnClickListener {
            NavigationActivity.start(requireContext())
        }
    }

    override fun onResume() {
        super.onResume()
        // 从详情页返回时订单状态可能已推进（取货/送达），重拉一次再刷新卡片。
        viewLifecycleOwner.lifecycleScope.launch {
            withContext(Dispatchers.IO) { DriverRemoteRepository.refreshOrders() }
            if (_binding != null) bindActiveOrder()
        }
    }

    /** 进行中订单卡片：只展示后端真实数据，没有订单 ID 时点击给出明确提示而不是打开假详情。 */
    private fun bindActiveOrder() {
        val order = vm.activeOrder
        binding.tvPickName.text = order.pickName
        binding.tvPickAddr.text = order.pickAddr
        binding.tvDropName.text = order.dropAddr.substringBefore(" · ")
        val contact = listOf(order.dropName, order.dropContact).filter { it.isNotBlank() }.joinToString(" ")
        binding.tvDropAddr.text = if (contact.isBlank()) "送货 · 暂无联系人" else "送货 · $contact"

        binding.layoutActiveOrder.setOnClickListener {
            if (order.id.isBlank()) {
                requireContext().toast("暂无进行中订单")
                return@setOnClickListener
            }
            OrderDetailActivity.start(requireContext(), order.id)
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
