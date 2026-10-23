package com.dawnerrand.store.ui.screens.menu

import android.content.Context
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.Dish
import com.dawnerrand.store.data.DishCategory
import com.dawnerrand.store.data.DishStore

/**
 * 菜品管理：分类 Tab + 分类下菜品列表 + 上下架切换（menu 稿）。
 * 数据来自 DishStore（本地持久化），切换上下架立即写回。
 */
class MenuViewModel : ViewModel() {

    val categories: LiveData<List<DishCategory>> = DishStore.categories

    private val _currentCat = MutableLiveData(0)
    val currentCat: LiveData<Int> = _currentCat

    fun selectCategory(index: Int) {
        _currentCat.value = index
    }

    fun currentIndex(): Int = _currentCat.value ?: 0

    fun dishes(cats: List<DishCategory>? = categories.value): List<Dish> =
        cats?.getOrNull(currentIndex())?.dishes ?: emptyList()

    fun currentCategoryName(): String =
        categories.value?.getOrNull(currentIndex())?.name ?: "热销爆款"

    fun toggleSale(ctx: Context, dish: Dish) {
        DishStore.toggleSale(ctx, dish.id)
    }
}
