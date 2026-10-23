package com.dawnerrand.driver.ui.screens.profile

import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import com.dawnerrand.driver.R
import com.dawnerrand.driver.data.repository.CardWalletRepository
import com.dawnerrand.driver.databinding.FragmentProfileBinding
import com.dawnerrand.driver.ui.screens.cardwallet.CardWalletActivity
import com.dawnerrand.driver.ui.screens.checkin.CheckinActivity
import com.dawnerrand.driver.ui.screens.support.SupportActivity
import com.dawnerrand.driver.ui.screens.settings.SettingsActivity
import com.dawnerrand.driver.ui.screens.finance.FinanceActivity
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.dp
import com.dawnerrand.driver.ui.util.toast

/**
 * 骑手中心：还原 design/driver/profile.html（渐变头部 + 统计 + 菜单卡片）。
 */
class ProfileFragment : Fragment() {

    private var _binding: FragmentProfileBinding? = null
    private val binding get() = _binding!!
    private val vm: ProfileViewModel by viewModels()

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View {
        _binding = FragmentProfileBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.tvProfileName.text = vm.rider.name
        binding.tvProfileDesc.text = "${vm.rider.area} · 已跑单${vm.rider.serviceDays}天"
        binding.tvProfileBadge.text = "🥈 ${vm.rider.level}"
        binding.tvStatScore.text = vm.rider.score
        binding.tvStatOnTime.text = vm.rider.onTimeRate
        binding.tvStatOrders.text = vm.rider.totalOrders.toString()
        binding.tvStatIncome.text = vm.rider.totalIncome

        buildMenuCards()
    }

    private fun buildMenuCards() {
        // 第一组：收入提现 / 我的配送范围 / 每日签到（签到入口为保证可达性的补充项）
        binding.menuCard1.removeAllViews()
        binding.menuCard1.addView(
            menuRow(R.drawable.ic_money, Design.PRIMARY, "收入提现", "¥1,286.5") {
                FinanceActivity.start(requireContext())
            }
        )
        binding.menuCard1.addView(
            menuRow(R.drawable.ic_pin, Design.PRIMARY, "我的配送范围", "绿杨乡全域") {
                requireContext().toast("配送范围：绿杨乡全域")
            }
        )
        binding.menuCard1.addView(
            menuRow(R.drawable.ic_calendar, Design.PRIMARY, "每日签到", "连续7天") {
                CheckinActivity.start(requireContext())
            }
        )

        // 第二组：骑手等级 / 卡包管理（原资质认证）
        binding.menuCard2.removeAllViews()
        binding.menuCard2.addView(
            menuRow(R.drawable.ic_star, Design.PRIMARY, "骑手等级", "银牌 · 再跑12单升金牌") {
                requireContext().toast("银牌骑手，再跑12单升金牌")
            }
        )
        binding.menuCard2.addView(
            menuRow(
                R.drawable.ic_card,
                Design.PRIMARY,
                "卡包管理",
                "银行卡 ${CardWalletRepository.bankCards.size} 张",
            ) {
                CardWalletActivity.start(requireContext())
            }
        )

        // 第三组：骑手客服 / 设置
        binding.menuCard3.removeAllViews()
        binding.menuCard3.addView(
            menuRow(R.drawable.ic_chat, Design.ORANGE_DEEP, "骑手客服", null) {
                SupportActivity.start(requireContext())
            }
        )
        binding.menuCard3.addView(
            menuRow(R.drawable.ic_settings, Design.PRIMARY, "设置", null) {
                SettingsActivity.start(requireContext())
            }
        )
    }

    private fun menuRow(iconRes: Int, iconColor: Int, label: String, value: String?, onClick: () -> Unit): LinearLayout {
        val row = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, requireContext().dp(14), 0, requireContext().dp(14))
            isClickable = true
            isFocusable = true
        }
        val iconBox = LinearLayout(requireContext()).apply {
            layoutParams = LinearLayout.LayoutParams(requireContext().dp(24), requireContext().dp(24))
            gravity = Gravity.CENTER
        }
        val icon = ImageView(requireContext()).apply {
            setImageResource(iconRes)
            setColorFilter(iconColor)
            layoutParams = LinearLayout.LayoutParams(requireContext().dp(20), requireContext().dp(20))
        }
        iconBox.addView(icon)

        val labelTv = TextView(requireContext()).apply {
            text = label
            textSize = 14f
            setTextColor(Design.TEXT_PRIMARY)
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }

        row.addView(iconBox)
        row.addView(labelTv)

        if (value != null) {
            val valueTv = TextView(requireContext()).apply {
                text = value
                textSize = 12f
                setTextColor(Design.TEXT_TERTIARY)
                layoutParams = LinearLayout.LayoutParams(LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            }
            val arrow = ImageView(requireContext()).apply {
                setImageResource(R.drawable.ic_chevron_right)
                setColorFilter(Design.TEXT_FAINT)
                layoutParams = LinearLayout.LayoutParams(requireContext().dp(14), requireContext().dp(14))
            }
            row.addView(valueTv)
            row.addView(arrow)
        } else {
            val arrow = ImageView(requireContext()).apply {
                setImageResource(R.drawable.ic_chevron_right)
                setColorFilter(Design.TEXT_FAINT)
                layoutParams = LinearLayout.LayoutParams(requireContext().dp(14), requireContext().dp(14))
            }
            row.addView(arrow)
        }

        row.setOnClickListener { onClick() }
        return row
    }

    override fun onResume() {
        super.onResume()
        // 卡包增删后刷新银行卡数量（返回本页时）
        if (_binding != null) {
            CardWalletRepository.load(requireContext())
            buildMenuCards()
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
