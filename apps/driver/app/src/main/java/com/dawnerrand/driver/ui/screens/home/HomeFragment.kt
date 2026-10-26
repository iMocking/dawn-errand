package com.dawnerrand.driver.ui.screens.home

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import com.dawnerrand.driver.databinding.FragmentHomeBinding
import com.dawnerrand.driver.ui.screens.navigation.NavigationActivity
import com.dawnerrand.driver.ui.screens.orderdetail.OrderDetailActivity
import com.dawnerrand.driver.ui.util.toast

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

        // 顶部渐变延伸到状态栏：设计稿 54px 留白即状态栏区域，无需额外内边距

        binding.tvRiderName.text = vm.rider.name
        binding.tvRiderTag.text = vm.rider.level
        binding.tvTodayAmount.text = vm.rider.todayAmount
        binding.tvTodayOrders.text = vm.rider.todayOrders.toString()
        binding.tvTodayScore.text = vm.rider.score
        binding.ringToday.progress = 0.65f

        binding.tvPickName.text = vm.activeOrder.pickName
        binding.tvPickAddr.text = vm.activeOrder.pickAddr
        // 设计稿 home.html：送点主行=地址、副行=送货·联系人（与详情页的 姓名/地址 字段语义不同，此处派生）
        binding.tvDropName.text = vm.activeOrder.dropAddr.substringBefore(" · ")
        binding.tvDropAddr.text = "送货 · ${vm.activeOrder.dropName} ${vm.activeOrder.dropContact}"

        binding.switchOnline.isOn = vm.isOnline
        binding.switchOnline.onCheckedChange = { on ->
            vm.setOnline(on)
            binding.tvOnlineSub.text = if (on) "● 正在接单 · 绿杨乡片区" else "● 已下线 · 点击上线接单"
        }
        vm.statusError.observe(viewLifecycleOwner) { msg ->
            if (msg != null) requireContext().toast(msg)
        }

        binding.layoutActiveOrder.setOnClickListener {
            OrderDetailActivity.start(requireContext())
        }
        binding.btnGoNav.setOnClickListener {
            NavigationActivity.start(requireContext())
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
