package com.dawnerrand.store.ui.screens.menu

import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.StoreRepository

/**
 * 新增/编辑菜品：图片/名称/分类/售价/原价/库存/月售/描述/上架开关（edit-dish 稿）。
 * 传入 dishId 时为编辑态，预填已有数据。
 */
class EditDishViewModel(dishId: String?) : ViewModel() {

    private val existing = dishId?.let { id ->
        StoreRepository.dishCategories.flatMap { it.dishes }.find { it.id == id }
    }

    val title: String = if (existing != null) "编辑菜品" else "新增菜品"

    val imageRes = existing?.iconRes ?: com.dawnerrand.store.R.drawable.ic_image
    val imageBgRes = existing?.bgRes ?: com.dawnerrand.store.R.drawable.bg_dish_1
    val imageColor: Int = existing?.iconColor ?: 0xFF86909C.toInt()

    val name = MutableLiveData(existing?.name ?: "")
    val category = MutableLiveData(existing?.let { "招牌炒菜" } ?: "招牌炒菜")
    val price = MutableLiveData(existing?.let { "¥${it.price}" } ?: "")
    val originalPrice = MutableLiveData("")
    val stock = MutableLiveData(existing?.stockText ?: "不限")
    val monthlySales = MutableLiveData(existing?.monthlySales ?: "自动统计")
    val desc = MutableLiveData(existing?.let { "招牌土猪肉现炒，香辣下饭" } ?: "")
    val onSale = MutableLiveData(existing?.onSale ?: true)

    fun save(onDone: () -> Unit) {
        onDone()
    }
}
