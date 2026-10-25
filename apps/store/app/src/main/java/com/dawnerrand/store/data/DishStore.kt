package com.dawnerrand.store.data

import android.content.Context
import android.content.SharedPreferences
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import com.dawnerrand.store.R
import org.json.JSONArray
import org.json.JSONObject

/**
 * 菜品仓库：新增/编辑/上下架写入共享状态并持久化到 SharedPreferences。
 * 分类与菜品初始值来自 StoreRepository（演示数据），首次启动后以本地数据为准。
 */
object DishStore {

    private const val PREF = "store_dishes"
    private lateinit var prefs: SharedPreferences

    private val _categories = MutableLiveData(StoreRepository.dishCategories)
    val categories: LiveData<List<DishCategory>> = _categories

    private val palette = listOf(
        Triple(R.drawable.ic_shop, 0xFFFF8A3D.toInt(), R.drawable.bg_dish_1),
        Triple(R.drawable.ic_smile, 0xFFFF6B2C.toInt(), R.drawable.bg_dish_2),
        Triple(R.drawable.ic_veg, 0x52C41A.toInt(), R.drawable.bg_dish_3),
        Triple(R.drawable.ic_egg, 0xFAAD14.toInt(), R.drawable.bg_dish_4),
        Triple(R.drawable.ic_flame, 0xF5222D.toInt(), R.drawable.bg_dish_5),
    )

    fun init(ctx: Context) {
        if (::prefs.isInitialized) return
        prefs = ctx.applicationContext.getSharedPreferences(PREF, Context.MODE_PRIVATE)
        val raw = prefs.getString("data", null) ?: return
        try {
            val arr = JSONArray(raw)
            val cats = mutableListOf<DishCategory>()
            for (i in 0 until arr.length()) {
                val c = arr.getJSONObject(i)
                val dishes = mutableListOf<Dish>()
                val da = c.optJSONArray("dishes") ?: JSONArray()
                for (j in 0 until da.length()) {
                    val d = da.getJSONObject(j)
                    dishes += Dish(
                        id = d.optString("id"),
                        name = d.optString("name"),
                        monthlySales = d.optString("monthlySales"),
                        stockText = d.optString("stockText"),
                        price = d.optString("price"),
                        iconRes = d.optInt("iconRes", R.drawable.ic_shop),
                        iconColor = d.optInt("iconColor", 0xFFFF6B2C.toInt()),
                        bgRes = d.optInt("bgRes", R.drawable.bg_dish_1),
                        onSale = d.optBoolean("onSale", true),
                        desc = d.optString("desc"),
                    )
                }
                cats += DishCategory(c.optString("name"), dishes)
            }
            if (cats.isNotEmpty()) _categories.value = cats
        } catch (e: Exception) {
            // 本地数据损坏时回退演示数据
        }
    }

    fun findDish(id: String): Dish? =
        _categories.value?.flatMap { it.dishes }?.firstOrNull { it.id == id }

    fun categoryOf(dishId: String): String? =
        _categories.value?.firstOrNull { c -> c.dishes.any { it.id == dishId } }?.name

    fun toggleSale(ctx: Context, dishId: String) {
        init(ctx)
        val cats = _categories.value ?: return
        _categories.value = cats.map { c ->
            c.copy(dishes = c.dishes.map { d ->
                if (d.id == dishId) d.copy(onSale = !d.onSale) else d
            })
        }
        persist()
    }

    /**
     * 保存菜品（新增或编辑）。
     * @return 保存后的菜品，参数非法时返回 null。
     */
    fun save(
        ctx: Context,
        existingId: String?,
        categoryName: String,
        name: String,
        price: String,
        stockText: String,
        monthlySales: String,
        onSale: Boolean,
        desc: String,
    ): Dish? {
        init(ctx)
        val cleanName = name.trim()
        val cleanPrice = price.trim().removePrefix("¥")
        val priceValue = cleanPrice.toDoubleOrNull() ?: return null
        if (cleanName.isEmpty() || priceValue <= 0) return null
        val target = categoryName.trim().ifBlank { "热销爆款" }
        val cats = (_categories.value ?: StoreRepository.dishCategories).toMutableList()
        val existing = existingId?.let { findDish(it) }

        // 编辑时先从原分类移除
        if (existing != null) {
            for (i in cats.indices) {
                cats[i] = cats[i].copy(dishes = cats[i].dishes.filterNot { it.id == existing.id })
            }
        }

        val index = cats.indexOfFirst { it.name == target }
        if (index < 0) cats += DishCategory(target, emptyList())
        val ci = cats.indexOfFirst { it.name == target }
        val count = cats[ci].dishes.size
        val style = palette[count % palette.size]
        val dish = existing?.copy(
            name = cleanName,
            price = formatPrice(priceValue),
            stockText = stockText.ifBlank { "库存充足" },
            monthlySales = monthlySales.ifBlank { existing.monthlySales },
            onSale = onSale,
            desc = desc,
        ) ?: Dish(
            id = "d" + System.currentTimeMillis(),
            name = cleanName,
            monthlySales = monthlySales.ifBlank { "月售0" },
            stockText = stockText.ifBlank { "库存充足" },
            price = formatPrice(priceValue),
            iconRes = style.first,
            iconColor = style.second,
            bgRes = style.third,
            onSale = onSale,
            desc = desc,
        )
        cats[ci] = cats[ci].copy(dishes = cats[ci].dishes + dish)
        _categories.value = cats
        persist()
        return dish
    }

    fun categoriesSnapshot(): List<DishCategory> = _categories.value ?: StoreRepository.dishCategories

    /** 用后端 /api/app/store/menu 返回的分类/菜品覆盖本地缓存。 */
    fun replaceFromApi(categories: List<DishCategory>) {
        if (categories.isEmpty()) return
        _categories.postValue(categories)
        persist()
    }

    private fun formatPrice(v: Double): String =
        if (v % 1.0 == 0.0) v.toInt().toString() else String.format("%.1f", v)

    private fun persist() {
        if (!::prefs.isInitialized) return
        val arr = JSONArray()
        (_categories.value ?: emptyList()).forEach { c ->
            val co = JSONObject()
            co.put("name", c.name)
            val da = JSONArray()
            c.dishes.forEach { d ->
                da.put(JSONObject().apply {
                    put("id", d.id)
                    put("name", d.name)
                    put("monthlySales", d.monthlySales)
                    put("stockText", d.stockText)
                    put("price", d.price)
                    put("iconRes", d.iconRes)
                    put("iconColor", d.iconColor)
                    put("bgRes", d.bgRes)
                    put("onSale", d.onSale)
                    put("desc", d.desc)
                })
            }
            co.put("dishes", da)
            arr.put(co)
        }
        prefs.edit().putString("data", arr.toString()).apply()
    }
}
