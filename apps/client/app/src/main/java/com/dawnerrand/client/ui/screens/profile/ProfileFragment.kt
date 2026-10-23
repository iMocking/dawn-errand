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
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.address.AddressActivity
import com.dawnerrand.client.ui.screens.coupons.CouponsActivity
import com.dawnerrand.client.ui.screens.orders.OrdersActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

class ProfileFragment : Fragment() {
    private val ctx: android.content.Context get() = requireContext()

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?,
    ): View {
        val scroll = ScrollView(requireContext()).apply {
            setBackgroundColor(Design.BACKGROUND)
            isVerticalScrollBarEnabled = false
        }
        val root = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
        }
        scroll.addView(root)

        root.addView(buildHeader())
        root.addView(buildAssets())
        root.addView(buildOrdersCard())
        root.addView(buildMenuCard(
            listOf(
                menuItem("收货地址", "3个地址", R.drawable.ic_pin, Color.rgb(0x2B, 0x7F, 0xFF)) {
                    startActivity(Intent(requireContext(), AddressActivity::class.java))
                },
                menuItem("优惠券", "2张可用", R.drawable.ic_ticket, Design.PRIMARY) {
                    startActivity(Intent(requireContext(), CouponsActivity::class.java))
                },
                menuItem("商家入驻", "开店赚钱", R.drawable.ic_food_box, Design.SUCCESS) {
                    ctx.toast("商家入驻暂未开放，敬请期待")
                },
                menuItem("成为骑手", "乡镇跑腿", R.drawable.ic_tab_errand, Design.PURPLE) {
                    ctx.toast("骑手招募功能建设中")
                },
            )
        ))
        root.addView(buildMenuCard(
            listOf(
                menuItem("联系客服", null, R.drawable.ic_chat, Color.rgb(0xFA, 0x8C, 0x16)) {
                    ctx.toast("客服热线 400-888-0000（演示）")
                },
                menuItem("帮助中心", null, R.drawable.ic_question, Design.TEXT_TERTIARY) {
                    ctx.toast("帮助中心建设中")
                },
                menuItem("设置", null, R.drawable.ic_settings, Design.TEXT_TERTIARY) {
                    ctx.toast("设置功能建设中")
                },
            )
        ))
        root.addView(View(requireContext()).apply {
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(24))
        })
        return scroll
    }

    private fun buildHeader(): LinearLayout {
        val header = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_home_top)
            setPadding(ctx.dp(20), ctx.dp(24), ctx.dp(20), ctx.dp(28))
            padTopWithStatusBar()
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
        }
        info.addView(TextView(requireContext()).apply {
            text = "王建国"
            setTextColor(Design.WHITE)
            textSize = 20f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        info.addView(TextView(requireContext()).apply {
            text = "138****8888 · 王家湾村"
            setTextColor(0xCCFFFFFF.toInt())
            textSize = 12f
            setPadding(0, ctx.dp(6), 0, 0)
        })
        header.addView(info)
        return header
    }

    private fun buildAssets(): LinearLayout {
        val card = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_card)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(72))
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.topMargin = ctx.dp(-16)
            }
        }
        val assets = listOf(
            Triple("2", "优惠券", true),
            Triple("¥15", "余额", false),
            Triple("128", "积分", false),
            Triple("★5", "会员等级", false),
        )
        assets.forEachIndexed { i, (num, lbl, clickable) ->
            val item = LinearLayout(requireContext()).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.MATCH_PARENT, 1f)
                if (clickable) {
                    isClickable = true
                    setOnClickListener {
                        startActivity(Intent(requireContext(), CouponsActivity::class.java))
                    }
                }
            }
            item.addView(TextView(requireContext()).apply {
                text = num
                setTextColor(Design.TEXT_PRIMARY)
                textSize = 20f
                typeface = android.graphics.Typeface.DEFAULT_BOLD
            })
            item.addView(TextView(requireContext()).apply {
                text = lbl
                setTextColor(Design.TEXT_TERTIARY)
                textSize = 11f
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
        head.setOnClickListener {
            startActivity(Intent(requireContext(), OrdersActivity::class.java))
        }
        card.addView(head)

        val icons = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(ctx.dp(8), ctx.dp(10), ctx.dp(8), ctx.dp(16))
        }
        val items = listOf(
            Pair("待支付", R.drawable.ic_tab_orders),
            Pair("待接单", R.drawable.ic_clock),
            Pair("配送中", R.drawable.ic_package),
            Pair("待评价", R.drawable.ic_star_filled),
        )
        items.forEach { (label, icon) ->
            val item = LinearLayout(requireContext()).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
                isClickable = true
                setOnClickListener {
                    startActivity(Intent(requireContext(), OrdersActivity::class.java))
                }
            }
            item.addView(ImageView(requireContext()).apply {
                setImageResource(icon)
                setColorFilter(Design.PRIMARY)
                layoutParams = LinearLayout.LayoutParams(ctx.dp(26), ctx.dp(26))
            })
            item.addView(TextView(requireContext()).apply {
                text = label
                setTextColor(Design.TEXT_SECONDARY)
                textSize = 11f
                setPadding(0, ctx.dp(6), 0, 0)
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
