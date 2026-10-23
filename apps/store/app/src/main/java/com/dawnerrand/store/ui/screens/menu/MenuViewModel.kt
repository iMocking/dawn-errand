package com.dawnerrand.store.ui.screens.menu

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.Dish
import com.dawnerrand.store.data.DishCategory
import com.dawnerrand.store.data.StoreRepository

/**
 * 菜品管理：分类 Tab + 分类下菜品列表 + 上下架切换（menu 稿）。
 */
class MenuViewModel : ViewModel() {

    val categories: List<DishCategory> = StoreRepository.dishCategories

    private val _currentCat = MutableLiveData(0)
    val currentCat: LiveData<Int> = _currentCat

    private val _dishes = MutableLiveData<List<Dish>>(emptyList())
    val dishes: LiveData<List<Dish>> = _dishes

    init {
        selectCategory(0)
    }

    fun selectCategory(index: Int) {
        _currentCat.value = index
        _dishes.value = categories.getOrNull(index)?.dishes ?: emptyList()
    }

    fun toggleSale(dish: Dish) {
        val updated = dish.copy(onSale = !dish.onSale)
        val cats = categories.map { c ->
            c.copy(dishes = c.dishes.map { if (it.id == dish.id) updated else it })
        }
        _dishes.value = cats.getOrNull(_currentCat.value ?: 0)?.dishes ?: emptyList()
    }
}
