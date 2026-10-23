package com.dawnerrand.store.ui.screens.menu

import android.content.Intent
import android.graphics.Typeface
import android.os.Bundle
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.R
import com.dawnerrand.store.data.Dish
import com.dawnerrand.store.data.DishCategory
import com.dawnerrand.store.databinding.ActivityMenuBinding
import com.dawnerrand.store.ui.adapter.DishAdapter
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.setLightStatusBars

/**
 * 菜品管理：分类 Tab + 菜品上下架/编辑/新增（menu 稿）。
 * 分类 Tab 为静态布局（5 个），文案按 DishStore 分类填充，避免动态测量的宽度塌陷问题。
 */
class MenuActivity : AppCompatActivity() {

    private lateinit var b: ActivityMenuBinding
    private val vm: MenuViewModel by viewModels()

    private val tabs = listOf(
        Triple(R.id.catTab0, R.id.catText0, R.id.catLine0),
        Triple(R.id.catTab1, R.id.catText1, R.id.catLine1),
        Triple(R.id.catTab2, R.id.catText2, R.id.catLine2),
        Triple(R.id.catTab3, R.id.catText3, R.id.catLine3),
        Triple(R.id.catTab4, R.id.catText4, R.id.catLine4),
    )

    private val adapter = DishAdapter(
        onToggleSale = { dish -> vm.toggleSale(this, dish) },
        onEdit = { dish -> openEdit(dish) },
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityMenuBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)
        b.root.setBackgroundColor(getColor(R.color.bg))

        b.dishList.layoutManager = LinearLayoutManager(this)
        b.dishList.adapter = adapter

        b.addBtn.setOnClickListener { openEdit(null) }

        tabs.forEachIndexed { index, tab ->
            findViewById<View>(tab.first).setOnClickListener { vm.selectCategory(index) }
        }

        vm.categories.observe(this) { cats ->
            bindTabLabels(cats)
            refreshCatStyles()
            adapter.submit(vm.dishes(cats))
        }
        vm.currentCat.observe(this) {
            refreshCatStyles()
            adapter.submit(vm.dishes())
        }
    }

    private fun bindTabLabels(cats: List<DishCategory>) {
        tabs.forEachIndexed { index, ids ->
            val tab = findViewById<View>(ids.first)
            val text = findViewById<TextView>(ids.second)
            val cat = cats.getOrNull(index)
            tab.visibility = if (cat == null) View.GONE else View.VISIBLE
            if (cat != null) text.text = cat.name
        }
    }

    private fun refreshCatStyles() {
        val current = vm.currentIndex()
        tabs.forEachIndexed { index, ids ->
            val tab = findViewById<LinearLayout>(ids.first)
            if (tab.visibility != View.VISIBLE) return@forEachIndexed
            val text = findViewById<TextView>(ids.second)
            val line = findViewById<View>(ids.third)
            val active = index == current
            text.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            text.setTypeface(text.typeface, if (active) Typeface.BOLD else Typeface.NORMAL)
            line.visibility = if (active) View.VISIBLE else View.INVISIBLE
        }
    }

    private fun openEdit(dish: Dish?) {
        startActivity(
            Intent(this, EditDishActivity::class.java)
                .putExtra("dishId", dish?.id)
                .putExtra("category", vm.currentCategoryName())
        )
    }
}
