package com.dawnerrand.client.ui.screens.profile

import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import com.dawnerrand.client.BuildConfig
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.OrderStatus
import com.dawnerrand.client.ui.screens.address.AddressActivity
import com.dawnerrand.client.ui.screens.coupons.CouponsActivity
import com.dawnerrand.client.ui.screens.customer.CustomerServiceActivity
import com.dawnerrand.client.ui.screens.editor.ProfileEditActivity
import com.dawnerrand.client.ui.screens.help.HelpCenterActivity
import com.dawnerrand.client.ui.screens.join.JoinActivity
import com.dawnerrand.client.ui.screens.main.MainActivity
import com.dawnerrand.client.ui.screens.orders.OrdersActivity
import com.dawnerrand.client.ui.screens.settings.SettingsActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 我的：用户资料 + 资产 + 订单入口 + 地址/优惠券/入驻/客服/帮助/设置。 */
class ProfileFragment : Fragment() {
    private val ctx: android.content.Context get() = requireContext()
    private lateinit var root: LinearLayout

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?,
    ): View {
        val scroll = ScrollView(requireContext()).apply {
            setBackgroundColor(Design.BACKGROUND)
            isVerticalScrollBarEnabled = false
        }
        root = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
        }
        scroll.addView(root)
        render()
        return scroll
    }

    override fun onResume() {
        super.onResume()
        if (::root.isInitialized) render()
    }

    override fun onHiddenChanged(hidden: Boolean) {
        super.onHiddenChanged(hidden)
        if (!hidden && ::root.isInitialized) render()
    }

    private fun render() {
        root.removeAllViews()
        root.addView(buildHeader())
        root.addView(buildAssets())
        root.addView(buildOrdersCard())
        root.addView(buildMenuCard(
            listOf(
                menuItem("收货地址", "${DemoRepository.addresses.size}个地址", R.drawable.ic_pin, Color.rgb(0x2B, 0x7F, 0xFF)) {
                    startActivity(Intent(requireContext(), AddressActivity::class.java))
                },
                menuItem("优惠券", "${DemoRepository.availableCouponCount()}张可用", R.drawable.ic_ticket, Design.PRIMARY) {
                    startActivity(Intent(requireContext(), CouponsActivity::class.java))
                },
                menuItem(
                    "商家入驻",
                    DemoRepository.applicationOf("merchant")?.status ?: "开店赚钱",
                    R.drawable.ic_food_box,
                    Design.SUCCESS,
                ) {
                    startActivity(Intent(requireContext(), JoinActivity::class.java)
                        .putExtra(JoinActivity.EXTRA_TYPE, JoinActivity.TYPE_MERCHANT))
                },
                menuItem(
                    "成为骑手",
                    DemoRepository.applicationOf("rider")?.status ?: "乡镇跑腿",
                    R.drawable.ic_tab_errand,
                    Design.PURPLE,
                ) {
                    startActivity(Intent(requireContext(), JoinActivity::class.java)
                        .putExtra(JoinActivity.EXTRA_TYPE, JoinActivity.TYPE_RIDER))
                },
            )
        ))
        root.addView(buildMenuCard(
            listOf(
                menuItem("联系客服", null, R.drawable.ic_chat, Color.rgb(0xFA, 0x8C, 0x16)) {
                    startActivity(Intent(requireContext(), CustomerServiceActivity::class.java))
                },
                menuItem("帮助中心", null, R.drawable.ic_question, Design.TEXT_TERTIARY) {
                    startActivity(Intent(requireContext(), HelpCenterActivity::class.java))
                },
                menuItem("设置", null, R.drawable.ic_settings, Design.TEXT_TERTIARY) {
                    startActivity(Intent(requireContext(), SettingsActivity::class.java))
                },
            )
        ))
        root.addView(TextView(requireContext()).apply {
            text = "曙光跑腿 v${BuildConfig.VERSION_NAME} · 乡镇跑腿，随叫随到"
            setTextColor(Design.TEXT_FAINT)
            textSize = 11f
            gravity = Gravity.CENTER
            setPadding(0, ctx.dp(8), 0, ctx.dp(80))
        })
    }

    private fun buildHeader(): LinearLayout {
        val header = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_home_top)
            setPadding(ctx.dp(20), ctx.dp(24), ctx.dp(20), ctx.dp(32))
            padTopWithStatusBar()
            isClickable = true
            setOnClickListener {
                startActivity(Intent(requireContext(), ProfileEditActivity::class.java))
            }
        }
        val avatar = LinearLayout(requireContext()).apply {
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_avatar)
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(ctx.dp(56), ctx.dp(56))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(14)
        }
        avatar.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_user)
            setColorFilter(Design.WHITE)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(32), ctx.dp(32))
        })
        header.addView(avatar)
        val info = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        info.addView(TextView(requireContext()).apply {
            text = DemoRepository.currentUser.name
            setTextColor(Design.WHITE)
            textSize = 20f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        info.addView(TextView(requireContext()).apply {
            text = "${DemoRepository.displayPhone()} · ${DemoRepository.shortLocation()}"
            setTextColor(0xCCFFFFFF.toInt())
            textSize = 12f
            setPadding(0, ctx.dp(6), 0, 0)
        })
        header.addView(info)
        header.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_chevron_right)
            setColorFilter(Design.WHITE)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(16), ctx.dp(16))
        })
        return header
    }

    private fun buildAssets(): LinearLayout {
        val card = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_card)
            setPadding(ctx.dp(6), 0, ctx.dp(6), 0)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(72))
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.topMargin = ctx.dp(-18)
            }
            elevation = ctx.dp(3).toFloat()
        }
        val balance = DemoRepository.currentUser.balance
        val balanceText = if (balance % 1.0 == 0.0) "¥${balance.toInt()}" else "¥%.1f".format(balance)
        val assets = listOf(
            Triple("${DemoRepository.availableCouponCount()}", "优惠券", true),
            Triple(balanceText, "余额", false),
            Triple("${DemoRepository.currentUser.points}", "积分", false),
            Triple(DemoRepository.currentUser.level, "会员等级", false),
        )
        assets.forEachIndexed { i, (num, lbl, clickable) ->
            val item = LinearLayout(requireContext()).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.MATCH_PARENT, 1f)
                isClickable = true
                setOnClickListener {
                    when (i) {
                        0 -> startActivity(Intent(requireContext(), CouponsActivity::class.java))
                        1 -> ctx.toast("账户余额 ${balanceText}，下单可直接抵扣")
                        2 -> ctx.toast("当前积分 ${DemoRepository.currentUser.points}，可兑换优惠券")
                        else -> ctx.toast("会员等级 ${DemoRepository.currentUser.level}，继续下单可升级")
                    }
                }
            }
            item.addView(TextView(requireContext()).apply {
                text = num
                setTextColor(Design.TEXT_PRIMARY)
                textSize = 20f
                typeface = android.graphics.Typeface.DEFAULT_BOLD
                gravity = Gravity.CENTER
            })
            item.addView(TextView(requireContext()).apply {
                text = lbl
                setTextColor(Design.TEXT_TERTIARY)
                textSize = 11f
                gravity = Gravity.CENTER
                setPadding(0, ctx.dp(4), 0, 0)
            })
            card.addView(item)
        }
        return card
    }

    private fun buildOrdersCard(): LinearLayout {
        val card = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_card)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.topMargin = ctx.dp(16)
            }
        }
        val head = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(ctx.dp(16), ctx.dp(14), ctx.dp(16), ctx.dp(6))
            isClickable = true
            setOnClickListener {
                (activity as? MainActivity)?.selectTab(2)
                    ?: startActivity(Intent(requireContext(), OrdersActivity::class.java))
            }
        }
        head.addView(TextView(requireContext()).apply {
            text = "我的订单"
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 16f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        })
        head.addView(TextView(requireContext()).apply {
            text = "全部订单"
            setTextColor(Design.TEXT_TERTIARY)
            textSize = 12f
        })
        head.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_chevron_right)
            setColorFilter(Design.TEXT_FAINT)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(14), ctx.dp(14))
        })
        card.addView(head)

        val icons = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(ctx.dp(8), ctx.dp(10), ctx.dp(8), ctx.dp(16))
        }
        val items = listOf(
            Triple("待支付", OrderStatus.PENDING_PAY, R.drawable.ic_tab_orders),
            Triple("待接单", OrderStatus.PENDING_ACCEPT, R.drawable.ic_clock),
            Triple("配送中", OrderStatus.DELIVERING, R.drawable.ic_package),
            Triple("待评价", OrderStatus.PENDING_REVIEW, R.drawable.ic_star_filled),
        )
        items.forEach { (label, status, icon) ->
            val count = DemoRepository.orders.count { it.status == status }
            val item = LinearLayout(requireContext()).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
                isClickable = true
                setOnClickListener {
                    val main = activity as? MainActivity
                    if (main != null) {
                        main.selectTab(2, status)
                    } else {
                        startActivity(Intent(requireContext(), OrdersActivity::class.java)
                            .putExtra(OrdersActivity.EXTRA_FILTER_STATUS, status))
                    }
                }
            }
            val iconBox = android.widget.FrameLayout(requireContext())
            iconBox.layoutParams = LinearLayout.LayoutParams(ctx.dp(34), ctx.dp(32))
            iconBox.addView(ImageView(requireContext()).apply {
                setImageResource(icon)
                setColorFilter(Design.PRIMARY)
                layoutParams = android.widget.FrameLayout.LayoutParams(ctx.dp(26), ctx.dp(26)).apply {
                    gravity = Gravity.CENTER
                }
            })
            if (count > 0) {
                iconBox.addView(TextView(requireContext()).apply {
                    text = if (count > 99) "99+" else count.toString()
                    setTextColor(Design.WHITE)
                    textSize = 9f
                    gravity = Gravity.CENTER
                    background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_badge_red)
                    layoutParams = android.widget.FrameLayout.LayoutParams(ctx.dp(16), ctx.dp(16)).apply {
                        gravity = Gravity.TOP or Gravity.END
                    }
                })
            }
            item.addView(iconBox)
            item.addView(TextView(requireContext()).apply {
                text = label
                setTextColor(Design.TEXT_SECONDARY)
                textSize = 11f
                gravity = Gravity.CENTER
                setPadding(0, ctx.dp(6), 0, 0)
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.MATCH_PARENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                )
            })
            icons.addView(item)
        }
        card.addView(icons)
        return card
    }

    private data class MenuDef(
        val label: String,
        val value: String?,
        val icon: Int,
        val color: Int,
        val onClick: () -> Unit,
    )

    private fun menuItem(
        label: String, value: String?, icon: Int, color: Int, onClick: () -> Unit,
    ) = MenuDef(label, value, icon, color, onClick)

    private fun buildMenuCard(items: List<MenuDef>): LinearLayout {
        val card = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_card)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.topMargin = ctx.dp(16)
            }
        }
        items.forEachIndexed { index, def ->
            card.addView(buildRow(def))
            if (index < items.size - 1) {
                card.addView(divider())
            }
        }
        return card
    }

    private fun divider(): View = View(requireContext()).apply {
        setBackgroundColor(Design.DIVIDER)
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.MATCH_PARENT, 1)
    }

    private fun buildRow(def: MenuDef): LinearLayout {
        val row = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(ctx.dp(16), ctx.dp(14), ctx.dp(16), ctx.dp(14))
            isClickable = true
            setOnClickListener { def.onClick() }
        }
        row.addView(ImageView(requireContext()).apply {
            setImageResource(def.icon)
            setColorFilter(def.color)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(20), ctx.dp(20))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(12)
        })
        row.addView(TextView(requireContext()).apply {
            text = def.label
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 14f
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        })
        def.value?.let {
            row.addView(TextView(requireContext()).apply {
                text = it
                setTextColor(Design.TEXT_TERTIARY)
                textSize = 12f
            })
        }
        row.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_chevron_right)
            setColorFilter(Design.TEXT_FAINT)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(14), ctx.dp(14))
            (layoutParams as LinearLayout.LayoutParams).leftMargin = ctx.dp(6)
        })
        return row
    }
}
