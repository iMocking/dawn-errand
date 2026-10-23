package com.dawnerrand.store.ui.screens.menu

import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.R
import com.dawnerrand.store.data.Dish
import com.dawnerrand.store.databinding.ActivityMenuBinding
import com.dawnerrand.store.ui.adapter.DishAdapter
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp

/**
 * 菜品管理：分类 Tab + 菜品上下架/编辑/新增（menu 稿）。
 */
class MenuActivity : AppCompatActivity() {

    private lateinit var b: ActivityMenuBinding
    private val vm: MenuViewModel by viewModels()

    private val adapter = DishAdapter(
        onToggleSale = { dish -> vm.toggleSale(dish) },
        onEdit = { dish -> openEdit(dish) },
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityMenuBinding.inflate(layoutInflater)
        setContentView(b.root)

        b.dishList.layoutManager = LinearLayoutManager(this)
        b.dishList.adapter = adapter

        b.addBtn.setOnClickListener { openEdit(null) }

        vm.dishes.observe(this) { adapter.submit(it) }
        vm.currentCat.observe(this) { refreshCatTabs() }

        buildCatTabs()
    }

    private fun buildCatTabs() {
        b.catTabs.removeAllViews()
        vm.categories.forEachIndexed { index, cat ->
            val tab = LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
                gravity = Gravity.CENTER_HORIZONTAL
                layoutParams = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT
                )
                isClickable = true
                setOnClickListener { vm.selectCategory(index) }
            }
            val text = TextView(this).apply {
                text = cat.name
                textSize = 14f
                setTextColor(Design.TEXT_SECONDARY)
                setPadding(dp(16), dp(12), dp(16), dp(12))
            }
            val line = View(this).apply {
                setBackgroundColor(Design.PRIMARY)
                layoutParams = LinearLayout.LayoutParams(dp(0), dp(2))
            }
            tab.addView(text)
            tab.addView(line)
            b.catTabs.addView(tab)
        }
        refreshCatTabs()
    }

    private fun refreshCatTabs() {
        val current = vm.currentCat.value ?: 0
        for (i in 0 until b.catTabs.childCount) {
            val tab = b.catTabs.getChildAt(i) as LinearLayout
            val text = tab.getChildAt(0) as TextView
            val line = tab.getChildAt(1) as View
            val active = i == current
            text.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            text.setTypeface(
                text.typeface,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL
            )
            line.layoutParams = LinearLayout.LayoutParams(
                if (active) dp(24) else dp(0),
                dp(2)
            )
        }
    }

    private fun openEdit(dish: Dish?) {
        startActivity(
            Intent(this, EditDishActivity::class.java)
                .putExtra("dishId", dish?.id)
        )
    }
}
