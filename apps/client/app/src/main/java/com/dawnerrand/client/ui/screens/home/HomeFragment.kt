package com.dawnerrand.client.ui.screens.home

import android.content.Intent
import android.graphics.Color
import android.graphics.drawable.ColorDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.HorizontalScrollView
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.Merchant
import com.dawnerrand.client.ui.screens.address.AddressActivity
import com.dawnerrand.client.ui.screens.errand.ErrandActivity
import com.dawnerrand.client.ui.screens.main.MainActivity
import com.dawnerrand.client.ui.screens.merchant.MerchantDetailActivity
import com.dawnerrand.client.ui.screens.search.SearchActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar

class HomeFragment : Fragment() {
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
        root.addView(buildSceneRow())
        root.addView(buildSectionHead("🔥 村里人都在点", "查看全部 ›"))
        root.addView(buildRecScroll())
        root.addView(buildSectionHead("📍 离你最近", null))
        DemoRepository.merchants.forEach { root.addView(buildNearbyCard(it)) }
        root.addView(View(requireContext()).apply {
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(24))
        })
        return scroll
    }

    override fun onResume() {
        super.onResume()
        refreshHeader()
    }

    override fun onHiddenChanged(hidden: Boolean) {
        super.onHiddenChanged(hidden)
        if (!hidden) refreshHeader()
    }

    private fun refreshHeader() {
        view?.findViewById<TextView>(R.id.tvGreeting)?.text = DemoRepository.greeting() + "，"
        view?.findViewById<TextView>(R.id.tvName)?.text = DemoRepository.currentUser.name + " 👋"
        view?.findViewById<TextView>(R.id.tvLocation)?.text = DemoRepository.shortLocation()
    }

    // ---------- 渐变问候区 ----------
    private fun buildHeader(): LinearLayout {
        val header = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_home_top)
            setPadding(ctx.dp(20), ctx.dp(24), ctx.dp(20), ctx.dp(20))
            padTopWithStatusBar()
        }
        val helloRow = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
        }
        helloRow.addView(TextView(requireContext()).apply {
            id = R.id.tvGreeting
            text = DemoRepository.greeting() + "，"
            setTextColor(Design.WHITE)
            textSize = 28f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        helloRow.addView(TextView(requireContext()).apply {
            id = R.id.tvName
            text = DemoRepository.currentUser.name + " 👋"
            setTextColor(Design.WHITE)
            textSize = 28f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        header.addView(helloRow)

        val locRow = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, ctx.dp(8), 0, 0)
            isClickable = true
            setOnClickListener {
                startActivity(Intent(requireContext(), AddressActivity::class.java))
            }
        }
        locRow.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_pin)
            setColorFilter(Design.WHITE)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(12), ctx.dp(12))
        })
        locRow.addView(TextView(requireContext()).apply {
            id = R.id.tvLocation
            text = DemoRepository.shortLocation()
            setTextColor(Design.WHITE)
            textSize = 13f
            setPadding(ctx.dp(6), 0, 0, 0)
        })
        locRow.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_chevron_down)
            setColorFilter(Design.WHITE)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(10), ctx.dp(10))
        })
        header.addView(locRow)

        val search = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_search_bar)
            setPadding(ctx.dp(12), 0, ctx.dp(12), 0)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(40))
            (layoutParams as LinearLayout.LayoutParams).topMargin = ctx.dp(16)
            isClickable = true
            setOnClickListener {
                startActivity(Intent(requireContext(), SearchActivity::class.java))
            }
        }
        search.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_search)
            setColorFilter(Design.TEXT_TERTIARY)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(16), ctx.dp(16))
        })
        search.addView(TextView(requireContext()).apply {
            text = "想吃点啥？搜\"小炒肉\"\"大米\""
            setTextColor(Design.TEXT_TERTIARY)
            textSize = 13f
            setPadding(ctx.dp(8), 0, 0, 0)
        })
        header.addView(search)
        return header
    }

    // ---------- 场景卡 ----------
    private fun buildSceneRow(): LinearLayout {
        val row = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(ctx.dp(16), ctx.dp(16), ctx.dp(16), 0)
        }
        row.addView(buildSceneCard("🍚 点个饭", "村里餐馆送到家", R.drawable.bg_scene_eat) {
            startActivity(Intent(requireContext(), MerchantDetailActivity::class.java)
                .putExtra("merchant_id", "m1"))
        })
        row.addView(buildSceneCard("🏃 跑个腿", "代买代取代送", R.drawable.bg_scene_run) {
            (activity as? MainActivity)?.selectTab(1)
                ?: startActivity(Intent(requireContext(), ErrandActivity::class.java))
        })
        return row
    }

    private fun buildSceneCard(
        title: String, sub: String, bg: Int, onClick: () -> Unit,
    ): LinearLayout {
        val card = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            background = ContextCompat.getDrawable(requireContext(), bg)
            setPadding(ctx.dp(16), 0, ctx.dp(14), 0)
            layoutParams = LinearLayout.LayoutParams(0, ctx.dp(76), 1f)
            (layoutParams as LinearLayout.LayoutParams).let {
                if (title.contains("跑")) it.leftMargin = ctx.dp(10) else it.rightMargin = ctx.dp(10)
            }
            isClickable = true
            setOnClickListener { onClick() }
        }
        val left = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
        }
        left.addView(TextView(requireContext()).apply {
            text = title
            setTextColor(Design.WHITE)
            textSize = 17f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        left.addView(TextView(requireContext()).apply {
            text = sub
            setTextColor(0xCCFFFFFF.toInt())
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        card.addView(left)
        card.addView(ImageView(requireContext()).apply {
            setImageResource(R.drawable.ic_chevron_right)
            setColorFilter(Design.WHITE)
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_circle_white25)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(28), ctx.dp(28))
            (layoutParams as LinearLayout.LayoutParams).leftMargin = ctx.dp(8)
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(0)
            scaleType = ImageView.ScaleType.CENTER_INSIDE
        })
        return card
    }

    // ---------- 区块标题 ----------
    private fun buildSectionHead(title: String, more: String?): LinearLayout {
        val head = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(ctx.dp(16), ctx.dp(20), ctx.dp(16), ctx.dp(10))
            if (more != null) {
                isClickable = true
                setOnClickListener {
                    startActivity(Intent(requireContext(), SearchActivity::class.java))
                }
            }
        }
        head.addView(TextView(requireContext()).apply {
            text = title
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 17f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        if (more != null) {
            val spacer = View(requireContext()).apply {
                layoutParams = LinearLayout.LayoutParams(0, 1, 1f)
            }
            head.addView(spacer)
            head.addView(TextView(requireContext()).apply {
                text = more
                setTextColor(Design.TEXT_TERTIARY)
                textSize = 12f
            })
        }
        return head
    }

    // ---------- 推荐横滑 ----------
    private fun buildRecScroll(): HorizontalScrollView {
        val hsv = HorizontalScrollView(requireContext()).apply {
            isHorizontalScrollBarEnabled = false
            setPadding(ctx.dp(16), 0, ctx.dp(16), 0)
        }
        val row = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
        }
        DemoRepository.merchants.forEachIndexed { i, m ->
            row.addView(buildRecCard(m, i))
        }
        hsv.addView(row)
        return hsv
    }

    private fun buildRecCard(m: Merchant, index: Int): LinearLayout {
        val card = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(0, 0, ctx.dp(12), 0)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(140), LinearLayout.LayoutParams.WRAP_CONTENT)
            isClickable = true
            setOnClickListener {
                startActivity(Intent(requireContext(), MerchantDetailActivity::class.java)
                    .putExtra("merchant_id", m.id))
            }
        }
        val img = LinearLayout(requireContext()).apply {
            background = ColorDrawable(m.bannerColorStart)
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, ctx.dp(88))
            (layoutParams as LinearLayout.LayoutParams).bottomMargin = ctx.dp(8)
        }
        img.addView(ImageView(requireContext()).apply {
            setImageResource(m.iconRes)
            setColorFilter(m.iconColor)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(40), ctx.dp(40))
        })
        card.addView(img)
        card.addView(TextView(requireContext()).apply {
            text = m.recName
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 14f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            maxLines = 1
        })
        card.addView(TextView(requireContext()).apply {
            text = "${m.score}分 · ${m.deliveryMinutes}分钟"
            setTextColor(Design.TEXT_SECONDARY)
            textSize = 12f
            setPadding(0, ctx.dp(2), 0, 0)
        })
        return card
    }

    // ---------- 附近商家卡 ----------
    private fun buildNearbyCard(m: Merchant): LinearLayout {
        val card = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            background = ContextCompat.getDrawable(requireContext(), R.drawable.bg_card)
            setPadding(ctx.dp(12), ctx.dp(12), ctx.dp(12), ctx.dp(12))
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).let {
                it.leftMargin = ctx.dp(16); it.rightMargin = ctx.dp(16); it.bottomMargin = ctx.dp(10)
            }
            isClickable = true
            setOnClickListener {
                startActivity(Intent(requireContext(), MerchantDetailActivity::class.java)
                    .putExtra("merchant_id", m.id))
            }
        }
        val banner = LinearLayout(requireContext()).apply {
            background = ColorDrawable(m.bannerColorStart)
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(ctx.dp(56), ctx.dp(56))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(12)
        }
        banner.addView(ImageView(requireContext()).apply {
            setImageResource(m.iconRes)
            setColorFilter(m.iconColor)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(28), ctx.dp(28))
        })
        card.addView(banner)

        val info = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        val top = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
        }
        top.addView(TextView(requireContext()).apply {
            text = m.name
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 16f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            maxLines = 1
        })
        top.addView(TextView(requireContext()).apply {
            text = m.distance
            setTextColor(Design.TEXT_TERTIARY)
            textSize = 12f
        })
        info.addView(top)
        info.addView(TextView(requireContext()).apply {
            text = "⭐ ${m.score} · 月售${m.monthlySales}单 · ${m.deliveryMinutes}分钟送到"
            setTextColor(Design.TEXT_SECONDARY)
            textSize = 12f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        val tags = LinearLayout(requireContext()).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(0, ctx.dp(6), 0, 0)
        }
        m.tags.forEach { tag ->
            tags.addView(tagView(tag, R.drawable.bg_tag_food, Design.PRIMARY))
        }
        m.freeTag?.let {
            tags.addView(tagView(it, R.drawable.bg_tag_takeout, Design.SUCCESS))
        }
        info.addView(tags)
        card.addView(info)
        return card
    }

    private fun tagView(text: String, bg: Int, color: Int): TextView {
        return TextView(requireContext()).apply {
            this.text = text
            setTextColor(color)
            textSize = 11f
            background = ContextCompat.getDrawable(requireContext(), bg)
            setPadding(ctx.dp(6), ctx.dp(2), ctx.dp(6), ctx.dp(2))
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(6)
        }
    }
}
