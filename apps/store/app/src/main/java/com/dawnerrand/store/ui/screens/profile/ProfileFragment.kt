package com.dawnerrand.store.ui.screens.profile

import android.content.Intent
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.LinearLayout
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import com.dawnerrand.store.R
import com.dawnerrand.store.data.ProfileMenuItem
import com.dawnerrand.store.data.ReviewStore
import com.dawnerrand.store.data.SettlementStore
import com.dawnerrand.store.databinding.FragmentProfileBinding
import com.dawnerrand.store.databinding.ItemProfileMenuBinding
import com.dawnerrand.store.ui.screens.qualification.QualificationActivity
import com.dawnerrand.store.ui.screens.reviews.ReviewsActivity
import com.dawnerrand.store.ui.screens.settlement.SettlementActivity
import com.dawnerrand.store.ui.screens.settings.SettingsActivity
import com.dawnerrand.store.ui.screens.statistics.StatisticsActivity
import com.dawnerrand.store.ui.screens.support.SupportActivity
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.padTopWithStatusBar
import com.dawnerrand.store.ui.util.toast

/**
 * 商家中心：店铺信息 + 资质/结算/经营/评价/客服/设置（profile 稿）。
 * 菜单按设计稿分三张卡片；结算账户、评价数量在 onResume 时刷新。
 */
class ProfileFragment : Fragment() {

    private var _b: FragmentProfileBinding? = null
    private val b get() = _b!!
    private val vm: ProfileViewModel by viewModels()

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View {
        _b = FragmentProfileBinding.inflate(inflater, container, false)
        return b.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        b.headerBox.padTopWithStatusBar()
        b.shopName.text = vm.shop.name
        b.shopDesc.text = vm.shop.desc
        renderMenu()
    }

    override fun onResume() {
        super.onResume()
        if (_b != null) renderMenu()
    }

    private fun renderMenu() {
        val ctx = requireContext()
        b.menuContainer.removeAllViews()
        val menus = menusWithValues()
        // 设计稿分组：资质/结算 · 经营/评价 · 客服/设置
        val groups = listOf(menus.subList(0, 2), menus.subList(2, 4), menus.subList(4, 6))
        groups.forEach { group ->
            val card = LinearLayout(ctx).apply {
                orientation = LinearLayout.VERTICAL
                setBackgroundResource(R.drawable.bg_card_white)
                setPadding(ctx.dp(16), 0, ctx.dp(16), 0)
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.MATCH_PARENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT
                ).apply { bottomMargin = ctx.dp(10) }
            }
            group.forEachIndexed { index, item ->
                val row = ItemProfileMenuBinding.inflate(layoutInflater, card, false)
                row.menuIcon.setImageResource(item.iconRes)
                row.menuIcon.setColorFilter(item.iconColor)
                row.menuLabel.text = item.label
                row.menuValue.text = item.value.orEmpty()
                row.menuValue.visibility = if (item.value.isNullOrBlank()) View.GONE else View.VISIBLE
                row.divider.visibility = if (index < group.size - 1) View.VISIBLE else View.GONE
                row.root.setOnClickListener { open(item) }
                card.addView(row.root)
            }
            b.menuContainer.addView(card)
        }
    }

    private fun menusWithValues(): List<ProfileMenuItem> {
        val ctx = requireContext()
        return vm.menus.map { item ->
            when (item.label) {
                "结算账户" -> item.copy(value = SettlementStore.summary(ctx))
                "顾客评价" -> item.copy(value = "${ReviewStore.rating}分 · ${ReviewStore.total}条")
                else -> item
            }
        }
    }

    private fun open(item: ProfileMenuItem) {
        val target = when (item.label) {
            "资质证照" -> QualificationActivity::class.java
            "结算账户" -> SettlementActivity::class.java
            "经营数据" -> StatisticsActivity::class.java
            "顾客评价" -> ReviewsActivity::class.java
            "商家客服" -> SupportActivity::class.java
            "设置" -> SettingsActivity::class.java
            else -> null
        }
        if (target == null) {
            requireContext().toast("「${item.label}」暂未开放")
        } else {
            startActivity(Intent(requireContext(), target))
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _b = null
    }
}
