package com.dawnerrand.store.ui.screens.menu

import android.content.Context
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.Dish
import com.dawnerrand.store.data.DishStore

/**
 * 新增/编辑菜品：图片/名称/分类/售价/原价/库存/月售/描述/上架开关（edit-dish 稿）。
 * 传入 dishId 时为编辑态，预填已有数据；保存写入 DishStore 并本地持久化。
 */
class EditDishViewModel(dishId: String?, defaultCategory: String?) : ViewModel() {

    private val existing: Dish? = dishId?.let { DishStore.findDish(it) }

    val title: String = if (existing != null) "编辑菜品" else "新增菜品"

    val imageRes = existing?.iconRes ?: com.dawnerrand.store.R.drawable.ic_image
    val imageBgRes = existing?.bgRes ?: com.dawnerrand.store.R.drawable.bg_dish_1
    val imageColor: Int = existing?.iconColor ?: 0xFF86909C.toInt()

    val name = MutableLiveData(existing?.name ?: "")
    val category = MutableLiveData(
        existing?.let { DishStore.categoryOf(it.id) } ?: defaultCategory ?: "热销爆款"
    )
    val price = MutableLiveData(existing?.price ?: "")
    val originalPrice = MutableLiveData("")
    val stock = MutableLiveData(existing?.stockText ?: "库存充足")
    val monthlySales = MutableLiveData(existing?.monthlySales ?: "自动统计")
    val desc = MutableLiveData(existing?.desc ?: "")
    val onSale = MutableLiveData(existing?.onSale ?: true)

    fun save(ctx: Context, onDone: (Dish?) -> Unit) {
        val saved = DishStore.save(
            ctx = ctx,
            existingId = existing?.id,
            categoryName = category.value.orEmpty(),
            name = name.value.orEmpty(),
            price = price.value.orEmpty(),
            stockText = stock.value.orEmpty(),
            monthlySales = if (existing != null) monthlySales.value.orEmpty() else "月售0",
            onSale = onSale.value ?: true,
            desc = desc.value.orEmpty(),
        )
        onDone(saved)
    }
}
