package com.dawnerrand.client.ui.screens.merchant

import android.content.Intent
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.Dish
import com.dawnerrand.client.data.Merchant
import com.dawnerrand.client.ui.screens.orderconfirm.OrderConfirmActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.confirmDialog
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padBottomWithNavBar
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.toast

/** 商家详情：渐变 Hero + 信息卡 + 分类 chips + 菜品 + 深色购物车栏。 */
class MerchantDetailActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private lateinit var merchant: Merchant
    private val catTabs = mutableListOf<LinearLayout>()
    private var currentCat = 0
    private lateinit var tvTotal: TextView
    private lateinit var tvBadge: TextView
    private lateinit var btnCheckout: TextView
    private lateinit var dishList: LinearLayout
    private var categories: List<String> = emptyList()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_merchant)
        window.decorView.padTopWithStatusBar()
        window.decorView.padBottomWithNavBar()

        val id = intent.getStringExtra("merchant_id") ?: "m1"
        merchant = DemoRepository.merchantById(id)

        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }
        tvTotal = findViewById(R.id.cartTotal)
        tvBadge = findViewById(R.id.cartBadge)
        btnCheckout = findViewById(R.id.btnCheckout)
        dishList = findViewById(R.id.dishList)
        btnCheckout.setOnClickListener { checkout() }
        findViewById<View>(R.id.cartBar).setOnLongClickListener {
            if (DemoRepository.cartCountFor(merchant.id) > 0) {
                ctx.confirmDialog("清空购物车", "确定清空当前商家的购物车吗？") {
                    DemoRepository.clearCart(merchant.id)
                    renderCart()
                    renderDishes()
                    ctx.toast("购物车已清空")
                }
            }
            true
        }

        // Hero 渐变
        val hero = findViewById<View>(R.id.hero)
        val grad = GradientDrawable(
            GradientDrawable.Orientation.TL_BR,
            intArrayOf(merchant.bannerColorStart, merchant.bannerColorEnd)
        )
        hero.background = grad
        findViewById<TextView>(R.id.heroBadge).text =
            "🔥 ${merchant.tags.firstOrNull()?.removePrefix("🔥 ") ?: "热销"} · 已售${merchant.monthlySales}份"

        findViewById<TextView>(R.id.shopName).text = merchant.name
        findViewById<TextView>(R.id.metaScore).text = "⭐ ${merchant.score}分"
        findViewById<TextView>(R.id.metaDist).text = "📏 ${merchant.distance}"
        findViewById<TextView>(R.id.metaTime).text = "⏱ 约${merchant.deliveryMinutes}分钟"

        val tags = findViewById<LinearLayout>(R.id.shopTags)
        tags.addView(stag(merchant.freeTag ?: "满${merchant.freeDeliveryThreshold}免配送", Design.PRIMARY))
        tags.addView(stag("新用户立减3元", Design.SUCCESS))

        // 分类 chips
        categories = DemoRepository.categoriesFor(merchant.id)
        val catRow = findViewById<LinearLayout>(R.id.catRow)
        categories.forEachIndexed { i, c ->
            catTabs.add(buildCatChip(i, c))
            catRow.addView(catTabs[i])
        }
        renderCats()
        renderDishes()
        renderCart()
    }

    override fun onResume() {
        super.onResume()
        if (::dishList.isInitialized) {
            renderDishes()
            renderCart()
        }
    }

    private fun stag(text: String, color: Int): TextView = TextView(this).apply {
        this.text = text
        textSize = 11f
        setTextColor(color)
        background = ContextCompat.getDrawable(this@MerchantDetailActivity,
            if (color == Design.PRIMARY) R.drawable.bg_tag_food else R.drawable.bg_tag_takeout)
        setPadding(ctx.dp(8), ctx.dp(3), ctx.dp(8), ctx.dp(3))
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
        (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(8)
    }

    private fun buildCatChip(index: Int, name: String): LinearLayout {
        val chip = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT, ctx.dp(32))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(8)
            isClickable = true
            setOnClickListener {
                currentCat = index
                renderCats()
                renderDishes()
            }
        }
        chip.addView(TextView(this).apply {
            text = name
            textSize = 13f
            setPadding(ctx.dp(12), ctx.dp(6), ctx.dp(12), ctx.dp(6))
            gravity = Gravity.CENTER
        })
        return chip
    }

    private fun renderCats() {
        catTabs.forEachIndexed { i, chip ->
            val tv = chip.getChildAt(0) as TextView
            val active = i == currentCat
            tv.setTextColor(if (active) Design.WHITE else Design.TEXT_SECONDARY)
            tv.setBackgroundResource(if (active) R.drawable.bg_chip_active else R.drawable.bg_chip)
            tv.typeface = android.graphics.Typeface.create(
                android.graphics.Typeface.DEFAULT,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
        }
    }

    private fun renderDishes() {
        dishList.removeAllViews()
        val category = categories.getOrElse(currentCat) { categories.firstOrNull() ?: "🔥 热销爆款" }
        val dishes = DemoRepository.dishesForCategory(merchant.id, category)
        if (dishes.isEmpty()) {
            dishList.addView(TextView(this).apply {
                text = "该分类暂无商品，看看其他分类吧"
                setTextColor(Design.TEXT_FAINT)
                textSize = 13f
                gravity = Gravity.CENTER
                setPadding(0, ctx.dp(60), 0, ctx.dp(60))
            })
            return
        }
        dishes.forEach { dishList.addView(buildDishCard(it)) }
    }

    private fun buildDishCard(dish: Dish): LinearLayout {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            background = ContextCompat.getDrawable(this@MerchantDetailActivity, R.drawable.bg_card)
            setPadding(ctx.dp(12), ctx.dp(12), ctx.dp(12), ctx.dp(12))
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT)
            (layoutParams as LinearLayout.LayoutParams).bottomMargin = ctx.dp(10)
        }
        val img = LinearLayout(this).apply {
            gravity = Gravity.CENTER
            background = android.graphics.drawable.ColorDrawable(dish.bgColor)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(56), ctx.dp(56))
            (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(12)
        }
        img.addView(ImageView(this).apply {
            setImageResource(dish.iconRes)
            setColorFilter(dish.iconColor)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(40), ctx.dp(40))
        })
        card.addView(img)

        val info = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        info.addView(TextView(this).apply {
            text = dish.name
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
        })
        info.addView(TextView(this).apply {
            text = dish.desc
            setTextColor(Design.TEXT_TERTIARY)
            textSize = 11f
            setPadding(0, ctx.dp(4), 0, 0)
        })
        val bottom = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, ctx.dp(8), 0, 0)
        }
        bottom.addView(TextView(this).apply {
            text = "¥${dish.price}"
            setTextColor(Design.REWARD)
            textSize = 16f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        })
        bottom.addView(buildQtyControl(dish))
        info.addView(bottom)
        card.addView(info)
        return card
    }

    private fun buildQtyControl(dish: Dish): LinearLayout {
        val count = DemoRepository.cartCountOf(dish.id)
        val box = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
        }
        if (count > 0) {
            box.addView(TextView(this).apply {
                text = "−"
                textSize = 18f
                gravity = Gravity.CENTER
                setTextColor(Design.PRIMARY)
                background = ContextCompat.getDrawable(this@MerchantDetailActivity, R.drawable.bg_chip)
                layoutParams = LinearLayout.LayoutParams(ctx.dp(26), ctx.dp(26))
                isClickable = true
                setOnClickListener {
                    DemoRepository.removeDish(dish)
                    renderDishes()
                    renderCart()
                }
            })
            box.addView(TextView(this).apply {
                text = count.toString()
                textSize = 14f
                gravity = Gravity.CENTER
                setTextColor(Design.TEXT_PRIMARY)
                typeface = android.graphics.Typeface.DEFAULT_BOLD
                layoutParams = LinearLayout.LayoutParams(ctx.dp(30), ctx.dp(26))
            })
        }
        box.addView(ImageView(this).apply {
            setImageResource(R.drawable.ic_add)
            background = ContextCompat.getDrawable(this@MerchantDetailActivity, R.drawable.bg_btn)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(26), ctx.dp(26))
            setColorFilter(Design.WHITE)
            scaleType = ImageView.ScaleType.CENTER_INSIDE
            isClickable = true
            setOnClickListener { addDish(dish) }
        })
        return box
    }

    private fun addDish(dish: Dish) {
        if (DemoRepository.addDish(merchant.id, dish)) {
            renderDishes()
            renderCart()
            return
        }
        AlertDialog.Builder(this)
            .setTitle("购物车已有其他商家的商品")
            .setMessage("添加当前商家的商品将清空原购物车，是否继续？")
            .setNegativeButton("取消", null)
            .setPositiveButton("清空并添加") { _, _ ->
                DemoRepository.clearCart(merchant.id)
                DemoRepository.addDish(merchant.id, dish)
                renderDishes()
                renderCart()
                ctx.toast("已加入购物车：${dish.name}")
            }
            .show()
    }

    private fun renderCart() {
        val count = DemoRepository.cartCountFor(merchant.id)
        val total = DemoRepository.cartTotalFor(merchant.id)
        tvTotal.text = "¥${if (count > 0) "%.1f".format(total) else "0"}"
        if (count > 0) {
            tvBadge.visibility = View.VISIBLE
            tvBadge.text = count.toString()
        } else {
            tvBadge.visibility = View.GONE
        }
    }

    private fun checkout() {
        if (DemoRepository.cartCountFor(merchant.id) == 0) {
            ctx.toast("购物车为空，先加几道菜吧")
            return
        }
        startActivity(Intent(this, OrderConfirmActivity::class.java)
            .putExtra("merchant_id", merchant.id))
    }
}
