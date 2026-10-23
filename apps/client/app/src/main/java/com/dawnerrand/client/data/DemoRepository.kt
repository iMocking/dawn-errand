package com.dawnerrand.client.data

import android.content.Context
import android.graphics.Color
import com.dawnerrand.client.R
import org.json.JSONArray
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Calendar
import java.util.Date
import java.util.Locale
import kotlin.math.max

/** 登录会话（SharedPreferences 持久化） */
data class SessionInfo(
    val token: String,
    val refreshToken: String,
    val phone: String,
    val name: String,
    val userId: String,
    val demo: Boolean,
)

/**
 * 客户端演示业务数据仓库：商家/菜品/购物车/订单/地址/优惠券/会话/设置。
 *
 * - 首次启动写入设计稿种子数据，之后所有增删改均持久化到 SharedPreferences；
 * - 后端 C 端业务接口就绪后，仅需把本仓库的同名方法替换为 ApiClient 调用，视图层零改动；
 * - 登录会话单独存于 `client_session`，短信登录成功后写入真实 token（演示降级时 demo=true）。
 */
object DemoRepository {

    private const val PREFS = "client_demo"
    private const val KEY = "snapshot_v3"
    private const val SESSION = "client_session"

    private lateinit var appContext: Context
    private var initialized = false

    // ---------------- 会话 ----------------

    private fun sessionPrefs(context: Context = appContext) =
        context.getSharedPreferences(SESSION, Context.MODE_PRIVATE)

    fun session(): SessionInfo? {
        val p = sessionPrefs()
        val token = p.getString("token", null) ?: return null
        return SessionInfo(
            token = token,
            refreshToken = p.getString("refresh_token", "").orEmpty(),
            phone = p.getString("phone", "").orEmpty(),
            name = p.getString("name", "").orEmpty(),
            userId = p.getString("user_id", "").orEmpty(),
            demo = p.getBoolean("demo", true),
        )
    }

    fun isLoggedIn(ctx: Context): Boolean = session()?.token != null

    fun login(
        token: String,
        refreshToken: String?,
        phone: String,
        name: String?,
        userId: String?,
        demo: Boolean,
    ) {
        val displayName = name?.takeIf { it.isNotBlank() } ?: defaultNameFor(phone)
        sessionPrefs().edit()
            .putString("token", token)
            .putString("refresh_token", refreshToken.orEmpty())
            .putString("phone", phone)
            .putString("name", displayName)
            .putString("user_id", userId.orEmpty())
            .putBoolean("demo", demo)
            .apply()
        currentUser = currentUser.copy(name = displayName, phone = phone)
        persist()
    }

    fun logout(ctx: Context) {
        sessionPrefs(ctx).edit().clear().apply()
    }

    fun defaultNameFor(phone: String): String {
        val digits = phone.filter { it.isDigit() }
        return if (digits == "13888888888") "王建国"
        else "用户" + digits.takeLast(4).ifBlank { "0000" }
    }

    fun maskPhone(phone: String): String {
        val digits = phone.filter { it.isDigit() }
        return if (digits.length == 11) {
            digits.substring(0, 3) + "****" + digits.substring(7)
        } else phone
    }

    // ---------------- 静态商家 / 菜品 ----------------

    val merchants = listOf(
        Merchant(
            "m1", "村口老陈家常菜", "老陈家常菜", "4.8", "2.1km", "326", "25",
            listOf("🔥 本周热销"), "满25免配送费",
            Color.rgb(0xFF, 0xE8, 0xD6), Color.rgb(0xFF, 0xD4, 0xB8),
            R.drawable.ic_food_bag, Color.rgb(0xFF, 0x8A, 0x3D),
        ),
        Merchant(
            "m2", "王家湾综合超市", "王家湾超市", "4.9", "0.8km", "512", "15",
            listOf("生鲜日配"), "烟酒饮料",
            Color.rgb(0xE8, 0xF5, 0xE9), Color.rgb(0xC8, 0xE6, 0xC9),
            R.drawable.ic_food_box, Color.rgb(0x52, 0xC4, 0x1A),
        ),
        Merchant(
            "m3", "乡里乡亲水果店", "乡亲水果店", "4.7", "3.5km", "189", "35",
            listOf("产地直发"), null,
            Color.rgb(0xFC, 0xE4, 0xEC), Color.rgb(0xF8, 0xBB, 0xD0),
            R.drawable.ic_food_fruit, Color.rgb(0xEB, 0x2F, 0x96),
        ),
        Merchant(
            "m4", "李记卤味", "李记卤味", "4.6", "2.0km", "240", "20",
            listOf("🔥 卤味熟食"), "满20免配送费",
            Color.rgb(0xFF, 0xF3, 0xE0), Color.rgb(0xFF, 0xE0, 0xB2),
            R.drawable.ic_food_cal, Color.rgb(0xFA, 0x8C, 0x16),
        ),
    )

    private val allDishes = listOf(
        // m1 老陈家常菜
        Dish("d1", "m1", "农家小炒肉", "招牌炒菜",
            "招牌土猪肉现炒，香辣下饭 · 月售86份", 16.8, "86",
            R.drawable.ic_food_bag, Color.rgb(0xFF, 0x8A, 0x3D), Color.rgb(0xFF, 0xF0, 0xE0)),
        Dish("d2", "m1", "红烧土鸡块", "汤类炖菜",
            "散养土鸡慢炖，入味软烂 · 月售64份", 28.0, "64",
            R.drawable.ic_food_meat, Color.rgb(0xFF, 0x6B, 0x2C), Color.rgb(0xFF, 0xE0, 0xD6)),
        Dish("d3", "m1", "手撕包菜", "家常小炒",
            "本地包菜大火快炒，脆嫩爽口 · 月售52份", 12.0, "52",
            R.drawable.ic_food_veg, Color.rgb(0x52, 0xC4, 0x1A), Color.rgb(0xE8, 0xF5, 0xE9)),
        Dish("d4", "m1", "西红柿炒蛋", "家常小炒",
            "农家土鸡蛋，酸甜可口 · 月售78份", 10.0, "78",
            R.drawable.ic_food_egg, Color.rgb(0xFA, 0xAD, 0x14), Color.rgb(0xFF, 0xF8, 0xE1)),
        // m2 王家湾综合超市
        Dish("d5", "m2", "纯牛奶 250ml×24", "乳品饮料",
            "本地牧场当日鲜奶 · 月售132箱", 49.9, "132",
            R.drawable.ic_food_box, Color.rgb(0x2B, 0x7F, 0xFF), Color.rgb(0xE8, 0xF2, 0xFF)),
        Dish("d6", "m2", "洗衣粉 2kg", "日用百货",
            "去渍留香，机洗手洗均可用 · 月售98袋", 19.9, "98",
            R.drawable.ic_package, Color.rgb(0x52, 0xC4, 0x1A), Color.rgb(0xE8, 0xF5, 0xE9)),
        Dish("d7", "m2", "原木抽纸 3层", "日用百货",
            "柔韧不掉屑，家庭装更划算 · 月售210提", 15.9, "210",
            R.drawable.ic_package, Color.rgb(0xFA, 0x8C, 0x16), Color.rgb(0xFF, 0xF3, 0xE0)),
        Dish("d8", "m2", "农家土鸡蛋 30枚", "粮油副食",
            "散养土鸡蛋，蛋黄饱满 · 月售76盒", 32.8, "76",
            R.drawable.ic_food_egg, Color.rgb(0xFA, 0xAD, 0x14), Color.rgb(0xFF, 0xF8, 0xE1)),
        // m3 乡里乡亲水果店
        Dish("d9", "m3", "红富士苹果", "时令水果",
            "脆甜多汁，产地直发 · 月售189份", 8.8, "189",
            R.drawable.ic_food_fruit, Color.rgb(0xEB, 0x2F, 0x96), Color.rgb(0xFC, 0xE4, 0xEC)),
        Dish("d10", "m3", "香蕉", "时令水果",
            "自然熟，软糯香甜 · 月售156份", 5.8, "156",
            R.drawable.ic_food_fruit, Color.rgb(0xFA, 0xAD, 0x14), Color.rgb(0xFF, 0xF8, 0xE1)),
        Dish("d11", "m3", "阳光玫瑰葡萄", "时令水果",
            "颗颗饱满，无籽脆甜 · 月售88份", 16.8, "88",
            R.drawable.ic_food_fruit, Color.rgb(0x52, 0xC4, 0x1A), Color.rgb(0xE8, 0xF5, 0xE9)),
        Dish("d12", "m3", "砂糖橘", "时令水果",
            "皮薄易剥，一口爆汁 · 月售142份", 6.8, "142",
            R.drawable.ic_food_fruit, Color.rgb(0xFA, 0x8C, 0x16), Color.rgb(0xFF, 0xF3, 0xE0)),
        // m4 李记卤味
        Dish("d13", "m4", "卤鸭脖", "卤味熟食",
            "老卤慢入味，越啃越香 · 月售120份", 12.8, "120",
            R.drawable.ic_food_meat, Color.rgb(0xFF, 0x6B, 0x2C), Color.rgb(0xFF, 0xE0, 0xD6)),
        Dish("d14", "m4", "卤鸡爪", "卤味熟食",
            "软糯脱骨，胶质满满 · 月售95份", 18.8, "95",
            R.drawable.ic_food_cal, Color.rgb(0xFA, 0x8C, 0x16), Color.rgb(0xFF, 0xF3, 0xE0)),
        Dish("d15", "m4", "卤牛肉", "卤味熟食",
            "黄牛肉现卤切片 · 月售66份", 32.8, "66",
            R.drawable.ic_food_meat, Color.rgb(0xEB, 0x2F, 0x96), Color.rgb(0xFC, 0xE4, 0xEC)),
        Dish("d16", "m4", "卤藕片", "卤味熟食",
            "脆爽入味，素卤首选 · 月售58份", 8.8, "58",
            R.drawable.ic_food_veg, Color.rgb(0x52, 0xC4, 0x1A), Color.rgb(0xE8, 0xF5, 0xE9)),
    )

    // ---------------- 可持久化状态 ----------------

    var currentUser = UserProfile("王建国", "13888888888")
    var addresses = mutableListOf<Address>()
    var orders = mutableListOf<Order>()
    var coupons = mutableListOf<Coupon>()
    var cart = mutableListOf<CartItem>()
    var cartMerchantId: String? = null
    var settings = ClientSettings()
    var searchHistory = mutableListOf<String>()
    var applications = mutableListOf<JoinApplication>()
    var feedbacks = mutableListOf<Feedback>()

    // ---------------- 初始化 / 持久化 ----------------

    fun init(context: Context) {
        if (initialized) return
        appContext = context.applicationContext
        val raw = prefs().getString(KEY, null)
        val restored = raw != null && runCatching { restore(raw) }.getOrDefault(false)
        if (!restored) {
            seed()
            persist()
        }
        session()?.let { s ->
            currentUser = currentUser.copy(
                name = s.name.ifBlank { defaultNameFor(s.phone) },
                phone = s.phone.ifBlank { currentUser.phone },
            )
        }
        initialized = true
    }

    private fun prefs() = appContext.getSharedPreferences(PREFS, Context.MODE_PRIVATE)

    private fun seed() {
        currentUser = UserProfile("王建国", "13888888888")
        addresses = mutableListOf(
            Address("a1", "王建国", "13888888888",
                "湖北省黄冈市浠水县\n王家湾村三组26号 · 红砖大铁门，门口有棵樟树", "家", true),
            Address("a2", "王秀兰", "13766666666",
                "湖北省黄冈市浠水县\n绿杨乡卫生院住院部 2楼3床", "医院"),
            Address("a3", "王强", "13599999999",
                "湖北省黄冈市浠水县\n兰溪中学 高三(2)班宿舍", "学校"),
        )
        coupons = mutableListOf(
            Coupon("c1", "5", "满20可用", "新用户立减券",
                "全场通用 · 外卖跑腿均可使用", "有效期至 2026-10-29", minSpend = 20.0),
            Coupon("c2", "3", "无门槛", "骑手好评奖励券",
                "完成订单评价后获得", "有效期至 2026-10-25", minSpend = 0.0),
            Coupon("c3", "8", "满30可用", "周末专享券",
                "仅限周六周日使用", "已过期 2026-10-20", disabled = true, minSpend = 30.0),
        )
        val now = System.currentTimeMillis()
        orders = mutableListOf(
            Order(
                id = "o1", orderNo = "DD202610240001", shop = "村口老陈家常菜",
                status = OrderStatus.DELIVERING,
                items = "农家小炒肉、西红柿炒蛋、米饭×2",
                time = "今天 12:30 · 预计12:45送达", total = "¥32.3",
                iconRes = R.drawable.ic_food_bag, iconColor = Color.rgb(0xFF, 0x8A, 0x3D),
                iconBg = Color.rgb(0xFF, 0xF0, 0xE0), type = "外卖",
                address = "王建国 138****8888 · 王家湾村三组26号",
                remark = "少辣，米饭要热的",
                lines = listOf(
                    OrderLine("农家小炒肉", 16.8, 1),
                    OrderLine("西红柿炒蛋", 10.0, 1),
                    OrderLine("米饭", 2.5, 2),
                ),
                subtotal = 31.8, deliveryFee = 3.0, packingFee = 1.5, discount = 4.0,
                createdAt = now - 12 * 60_000L, merchantId = "m1",
            ),
            Order(
                id = "o2", orderNo = "DD202610240002", shop = "李记卤味",
                status = OrderStatus.PENDING_ACCEPT,
                items = "卤鸭脖、卤藕片",
                time = "今天 12:18 · 等待商家接单", total = "¥23.1",
                iconRes = R.drawable.ic_food_cal, iconColor = Color.rgb(0xFA, 0x8C, 0x16),
                iconBg = Color.rgb(0xFF, 0xF3, 0xE0), type = "外卖",
                address = "王建国 138****8888 · 王家湾村三组26号",
                lines = listOf(OrderLine("卤鸭脖", 12.8, 1), OrderLine("卤藕片", 8.8, 1)),
                subtotal = 21.6, deliveryFee = 3.0, packingFee = 1.5, discount = 3.0,
                createdAt = now - 25 * 60_000L, merchantId = "m4",
            ),
            Order(
                id = "o3", orderNo = "DD202610230001", shop = "王家湾综合超市",
                status = OrderStatus.DONE, items = "纯牛奶×2箱、洗衣粉×1、抽纸×3",
                time = "昨天 18:20 · 已送达", total = "¥86.5",
                iconRes = R.drawable.ic_food_box, iconColor = Color.rgb(0x52, 0xC4, 0x1A),
                iconBg = Color.rgb(0xE8, 0xF5, 0xE9), type = "外卖",
                address = "王建国 138****8888 · 王家湾村三组26号",
                lines = listOf(
                    OrderLine("纯牛奶 250ml×24", 49.9, 1),
                    OrderLine("洗衣粉 2kg", 19.9, 1),
                    OrderLine("原木抽纸 3层", 15.9, 1),
                ),
                subtotal = 85.7, deliveryFee = 3.0, packingFee = 1.5, discount = 3.7,
                createdAt = now - 18 * 60 * 60_000L, merchantId = "m2",
            ),
            Order(
                id = "o4", orderNo = "DD202610200001", shop = "跑腿代办 · 镇中心",
                status = OrderStatus.DONE, items = "帮取快递（镇邮政所→王家湾村）",
                time = "10月20日 15:10 · 已送达", total = "¥6.0",
                iconRes = R.drawable.ic_tab_errand, iconColor = Color.rgb(0x00, 0xB5, 0x78),
                iconBg = Color.rgb(0xE8, 0xF7, 0xF0), type = "跑腿",
                address = "王家湾村三组26号",
                lines = listOf(OrderLine("跑腿费", 6.0, 1)),
                subtotal = 6.0, createdAt = now - 3 * 24 * 60 * 60_000L,
            ),
            Order(
                id = "o5", orderNo = "DD202610180001", shop = "乡里乡亲水果店",
                status = OrderStatus.DONE, items = "红富士苹果5斤、香蕉1把、葡萄2斤",
                time = "10月18日 10:00 · 已送达", total = "¥45.8",
                iconRes = R.drawable.ic_food_fruit, iconColor = Color.rgb(0xEB, 0x2F, 0x96),
                iconBg = Color.rgb(0xFC, 0xE4, 0xEC), type = "外卖",
                address = "王建国 138****8888 · 王家湾村三组26号",
                lines = listOf(
                    OrderLine("红富士苹果", 8.8, 2),
                    OrderLine("香蕉", 5.8, 2),
                    OrderLine("阳光玫瑰葡萄", 16.8, 1),
                ),
                subtotal = 46.0, deliveryFee = 3.0, packingFee = 1.5, discount = 4.7,
                createdAt = now - 6 * 24 * 60 * 60_000L, merchantId = "m3",
            ),
            Order(
                id = "o6", orderNo = "DD202610160001", shop = "村口老陈家常菜",
                status = OrderStatus.CANCELED, items = "手撕包菜、米饭×1",
                time = "10月16日 09:20 · 已取消", total = "¥18.0",
                iconRes = R.drawable.ic_food_veg, iconColor = Color.rgb(0x52, 0xC4, 0x1A),
                iconBg = Color.rgb(0xE8, 0xF5, 0xE9), type = "外卖",
                address = "王建国 138****8888 · 王家湾村三组26号",
                lines = listOf(OrderLine("手撕包菜", 12.0, 1), OrderLine("米饭", 2.5, 1)),
                subtotal = 14.5, deliveryFee = 3.0, packingFee = 1.5, discount = 1.0,
                createdAt = now - 8 * 24 * 60 * 60_000L, merchantId = "m1",
            ),
            Order(
                id = "o7", orderNo = "DD202610120001", shop = "乡里乡亲水果店",
                status = OrderStatus.REFUND, items = "阳光玫瑰葡萄×2（部分坏果）",
                time = "10月12日 16:40 · 退款处理中", total = "¥32.0",
                iconRes = R.drawable.ic_food_fruit, iconColor = Color.rgb(0xEB, 0x2F, 0x96),
                iconBg = Color.rgb(0xFC, 0xE4, 0xEC), type = "外卖",
                address = "王建国 138****8888 · 王家湾村三组26号",
                lines = listOf(OrderLine("阳光玫瑰葡萄", 16.8, 2)),
                subtotal = 33.6, deliveryFee = 3.0, packingFee = 1.5, discount = 6.1,
                createdAt = now - 12 * 24 * 60 * 60_000L, merchantId = "m3",
            ),
        )
        cart = mutableListOf()
        cartMerchantId = null
        settings = ClientSettings()
        searchHistory = mutableListOf("农家小炒肉", "王家湾综合超市", "西红柿炒蛋")
        applications = mutableListOf()
        feedbacks = mutableListOf()
    }

    private fun persist() {
        if (!::appContext.isInitialized) return
        val o = JSONObject()
        o.put("user", currentUser.toJson())
        o.put("addresses", JSONArray().apply { addresses.forEach { put(it.toJson()) } })
        o.put("orders", JSONArray().apply { orders.forEach { put(it.toJson()) } })
        o.put("coupons", JSONArray().apply { coupons.forEach { put(it.toJson()) } })
        o.put("cart", JSONArray().apply { cart.forEach { put(it.toJson()) } })
        o.put("cartMerchantId", cartMerchantId ?: "")
        o.put("settings", settings.toJson())
        o.put("searchHistory", JSONArray(searchHistory))
        o.put("applications", JSONArray().apply { applications.forEach { put(it.toJson()) } })
        o.put("feedbacks", JSONArray().apply { feedbacks.forEach { put(it.toJson()) } })
        prefs().edit().putString(KEY, o.toString()).apply()
    }

    private fun restore(raw: String): Boolean {
        val o = JSONObject(raw)
        currentUser = userFromJson(o.optJSONObject("user") ?: JSONObject())
        addresses = o.array("addresses").mapObjects { addressFromJson(it) }.toMutableList()
        orders = o.array("orders").mapObjects { normalizeOrder(orderFromJson(it)) }.toMutableList()
        coupons = o.array("coupons").mapObjects { couponFromJson(it) }.toMutableList()
        cart = mutableListOf<CartItem>().apply {
            o.array("cart").mapObjects { it }.forEach { item ->
                val dish = allDishes.firstOrNull { it.id == item.str("dishId") } ?: return@forEach
                add(CartItem(dish, max(1, item.int("count", 1))))
            }
        }
        cartMerchantId = o.str("cartMerchantId").takeIf { it.isNotBlank() }
        settings = settingsFromJson(o.optJSONObject("settings") ?: JSONObject())
        searchHistory = o.array("searchHistory").toStringList().toMutableList()
        applications = o.array("applications").mapObjects { applicationFromJson(it) }.toMutableList()
        feedbacks = o.array("feedbacks").mapObjects { feedbackFromJson(it) }.toMutableList()
        return true
    }

    // ---------------- 商家 / 菜品 ----------------

    /** 资源 id 不跨构建稳定，恢复订单时按商家/类型重新绑定图标 */
    private fun normalizeOrder(order: Order): Order {
        if (order.type == "跑腿") {
            return order.copy(
                iconRes = R.drawable.ic_tab_errand,
                iconColor = Color.rgb(0x00, 0xB5, 0x78),
                iconBg = Color.rgb(0xE8, 0xF7, 0xF0),
            )
        }
        val merchant = order.merchantId?.let { id -> merchants.firstOrNull { it.id == id } }
        return if (merchant != null) {
            order.copy(
                iconRes = merchant.iconRes,
                iconColor = merchant.iconColor,
                iconBg = merchant.bannerColorStart,
            )
        } else order
    }

    fun merchantById(id: String): Merchant = merchants.firstOrNull { it.id == id } ?: merchants.first()

    fun dishesFor(merchantId: String): List<Dish> = allDishes.filter { it.merchantId == merchantId }

    fun dishById(id: String): Dish? = allDishes.firstOrNull { it.id == id }

    fun categoriesFor(merchantId: String): List<String> =
        listOf("🔥 热销爆款") + dishesFor(merchantId).map { it.category }.distinct()

    fun dishesForCategory(merchantId: String, category: String): List<Dish> {
        val dishes = dishesFor(merchantId)
        return if (category.startsWith("🔥") || category == "全部") dishes
        else dishes.filter { it.category == category }
    }

    // ---------------- 购物车 ----------------

    /** 返回 false 表示购物车已有其他商家的商品 */
    fun addDish(merchantId: String, dish: Dish, count: Int = 1): Boolean {
        if (cartMerchantId != null && cartMerchantId != merchantId && cart.isNotEmpty()) return false
        cartMerchantId = merchantId
        val item = cart.firstOrNull { it.dish.id == dish.id }
        if (item != null) item.count += count else cart.add(CartItem(dish, count))
        persist()
        return true
    }

    fun removeDish(dish: Dish) {
        val item = cart.firstOrNull { it.dish.id == dish.id } ?: return
        item.count--
        if (item.count <= 0) cart.remove(item)
        if (cart.isEmpty()) cartMerchantId = null
        persist()
    }

    fun clearCart(merchantId: String? = null) {
        cart.clear()
        cartMerchantId = merchantId
        persist()
    }

    fun cartCountFor(merchantId: String): Int =
        if (cartMerchantId == merchantId) cart.sumOf { it.count } else 0

    fun cartTotalFor(merchantId: String): Double =
        if (cartMerchantId == merchantId) cart.sumOf { it.dish.price * it.count } else 0.0

    fun cartCountOf(dishId: String): Int =
        cart.firstOrNull { it.dish.id == dishId }?.count ?: 0

    fun cartCount(): Int = cart.sumOf { it.count }

    fun cartTotal(): Double = cart.sumOf { it.dish.price * it.count }

    fun cartSummary(): String =
        cart.joinToString("、") { "${it.dish.name}×${it.count}" }.ifBlank { "空购物车" }

    // ---------------- 地址 ----------------

    fun defaultAddress(): Address? =
        addresses.firstOrNull { it.isDefault } ?: addresses.firstOrNull()

    fun addressById(id: String): Address? = addresses.firstOrNull { it.id == id }

    fun saveAddress(
        id: String?,
        name: String,
        phone: String,
        detail: String,
        tag: String,
        isDefault: Boolean,
    ): Address {
        val existing = id?.let { addressById(it) }
        val result = if (existing != null) {
            val updated = existing.copy(
                name = name, phone = phone, detail = detail, tag = tag, isDefault = isDefault,
            )
            addresses[addresses.indexOf(existing)] = updated
            updated
        } else {
            val created = Address(
                id = "a" + System.currentTimeMillis(),
                name = name, phone = phone, detail = detail, tag = tag,
                isDefault = isDefault || addresses.isEmpty(),
            )
            addresses.add(created)
            created
        }
        if (result.isDefault) {
            for (i in addresses.indices) {
                addresses[i] = addresses[i].copy(isDefault = addresses[i].id == result.id)
            }
        }
        if (addresses.isNotEmpty() && addresses.none { it.isDefault }) {
            addresses[0] = addresses[0].copy(isDefault = true)
        }
        persist()
        return result
    }

    fun deleteAddress(id: String) {
        val wasDefault = addresses.firstOrNull { it.id == id }?.isDefault == true
        addresses.removeAll { it.id == id }
        if (wasDefault && addresses.isNotEmpty()) {
            addresses[0] = addresses[0].copy(isDefault = true)
        }
        persist()
    }

    fun setDefaultAddress(id: String) {
        for (i in addresses.indices) {
            addresses[i] = addresses[i].copy(isDefault = addresses[i].id == id)
        }
        persist()
    }

    // ---------------- 优惠券 ----------------

    fun availableCoupons(): List<Coupon> =
        coupons.filter { it.usable }.sortedByDescending { it.value }

    fun availableCouponCount(): Int = availableCoupons().size

    fun couponById(id: String): Coupon? = coupons.firstOrNull { it.id == id }

    fun couponUsableFor(coupon: Coupon?, subtotal: Double): Boolean =
        coupon != null && coupon.usable && subtotal >= coupon.minSpend

    fun markCouponUsed(id: String) {
        val index = coupons.indexOfFirst { it.id == id }
        if (index >= 0) {
            coupons[index] = coupons[index].copy(used = true)
            persist()
        }
    }

    // ---------------- 订单 ----------------

    fun orderById(id: String): Order? = orders.firstOrNull { it.id == id }

    fun ordersForTab(index: Int): List<Order> = when (index) {
        0 -> orders.filter { it.ongoing }
        1 -> orders.filter { it.finished }
        2 -> orders.filter { it.status == OrderStatus.CANCELED }
        3 -> orders.filter { it.status == OrderStatus.REFUND }
        else -> orders
    }

    fun orderCountForTab(index: Int): Int = ordersForTab(index).size

    fun actionsFor(order: Order): List<String> = when (order.status) {
        OrderStatus.PENDING_PAY -> listOf("取消订单", "去支付")
        OrderStatus.PENDING_ACCEPT -> listOf("取消订单", "查看配送")
        OrderStatus.DELIVERING -> listOf("查看配送", "确认收货")
        OrderStatus.PENDING_REVIEW -> listOf("去评价", "再来一单")
        OrderStatus.DONE -> if (order.reviewed) listOf("再来一单") else listOf("再来一单", "评价")
        OrderStatus.CANCELED -> listOf("删除订单", "再来一单")
        OrderStatus.REFUND -> listOf("查看进度")
        else -> emptyList()
    }

    fun createFoodOrder(
        merchant: Merchant,
        address: String,
        lines: List<OrderLine>,
        subtotal: Double,
        deliveryFee: Double,
        packingFee: Double,
        discount: Double,
        coupon: Coupon?,
        remark: String,
        payMethod: String,
    ): Order {
        val now = System.currentTimeMillis()
        val total = max(0.0, subtotal + deliveryFee + packingFee - discount - (coupon?.value ?: 0.0))
        val order = Order(
            id = "o$now",
            orderNo = "DD" + SimpleDateFormat("yyyyMMddHHmmss", Locale.CHINA).format(Date(now)),
            shop = merchant.name,
            status = OrderStatus.PENDING_ACCEPT,
            items = lines.filter { it.name != "配送费" && it.name != "打包费" }
                .joinToString("、") { "${it.name}×${it.count}" },
            time = "今天 " + SimpleDateFormat("HH:mm", Locale.CHINA).format(Date(now)) + " · 等待商家接单",
            total = "¥%.1f".format(total),
            iconRes = merchant.iconRes,
            iconColor = merchant.iconColor,
            iconBg = merchant.bannerColorStart,
            type = "外卖",
            address = address,
            remark = remark,
            lines = lines,
            subtotal = subtotal,
            deliveryFee = deliveryFee,
            packingFee = packingFee,
            discount = discount,
            couponTitle = coupon?.title.orEmpty(),
            couponAmount = coupon?.value ?: 0.0,
            createdAt = now,
            merchantId = merchant.id,
            payMethod = payMethod,
        )
        orders.add(0, order)
        coupon?.let { markCouponUsed(it.id) }
        cart.clear()
        cartMerchantId = null
        persist()
        return order
    }

    fun createErrandOrder(
        type: String,
        description: String,
        pickAddress: String,
        dropAddress: String,
        fee: Double,
    ): Order {
        val now = System.currentTimeMillis()
        val order = Order(
            id = "o$now",
            orderNo = "DD" + SimpleDateFormat("yyyyMMddHHmmss", Locale.CHINA).format(Date(now)),
            shop = "跑腿代办 · 镇中心",
            status = OrderStatus.PENDING_ACCEPT,
            items = "$type：$description",
            time = "今天 " + SimpleDateFormat("HH:mm", Locale.CHINA).format(Date(now)) + " · 等待骑手接单",
            total = "¥%.1f".format(fee),
            iconRes = R.drawable.ic_tab_errand,
            iconColor = Color.rgb(0x00, 0xB5, 0x78),
            iconBg = Color.rgb(0xE8, 0xF7, 0xF0),
            type = "跑腿",
            address = "$pickAddress → $dropAddress",
            remark = description,
            lines = listOf(OrderLine("跑腿费", fee, 1)),
            subtotal = fee,
            createdAt = now,
            payMethod = "微信支付",
        )
        orders.add(0, order)
        persist()
        return order
    }

    fun cancelOrder(id: String) {
        updateOrder(id) {
            it.copy(
                status = OrderStatus.CANCELED,
                time = todayStamp() + " · 订单已取消",
            )
        }
    }

    fun payOrder(id: String) {
        updateOrder(id) {
            it.copy(
                status = OrderStatus.PENDING_ACCEPT,
                time = todayStamp() + " · 等待商家接单",
            )
        }
    }

    fun confirmReceipt(id: String) {
        updateOrder(id) {
            it.copy(
                status = OrderStatus.PENDING_REVIEW,
                time = todayStamp() + " · 已送达，待评价",
            )
        }
    }

    fun reviewOrder(id: String, rating: Int, tags: List<String>, text: String) {
        updateOrder(id) {
            it.copy(
                status = OrderStatus.DONE,
                reviewed = true,
                rating = rating,
                reviewTags = tags.joinToString("、"),
                reviewText = text,
                time = todayStamp() + " · 已评价",
            )
        }
    }

    fun applyRefund(id: String) {
        updateOrder(id) {
            it.copy(
                status = OrderStatus.REFUND,
                time = todayStamp() + " · 退款处理中",
            )
        }
    }

    fun deleteOrder(id: String) {
        orders.removeAll { it.id == id }
        persist()
    }

    /** 再来一单：把订单中的菜品重新加入购物车；跑腿订单返回 false */
    fun reorder(order: Order): Boolean {
        val merchantId = order.merchantId ?: return false
        val dishes = order.lines.mapNotNull { line ->
            allDishes.firstOrNull { it.merchantId == merchantId && it.name == line.name }
        }
        if (dishes.isEmpty()) return false
        clearCart(merchantId)
        dishes.forEach { addDish(merchantId, it) }
        return true
    }

    private fun updateOrder(id: String, block: (Order) -> Order) {
        val index = orders.indexOfFirst { it.id == id }
        if (index < 0) return
        orders[index] = block(orders[index])
        persist()
    }

    // ---------------- 搜索 ----------------

    fun addSearchHistory(keyword: String) {
        val kw = keyword.trim()
        if (kw.isEmpty()) return
        searchHistory.remove(kw)
        searchHistory.add(0, kw)
        while (searchHistory.size > 8) searchHistory.removeAt(searchHistory.lastIndex)
        persist()
    }

    fun removeSearchHistory(keyword: String) {
        searchHistory.remove(keyword)
        persist()
    }

    fun clearSearchHistory() {
        searchHistory.clear()
        persist()
    }

    // ---------------- 设置 / 入驻 / 反馈 ----------------

    fun updateSettings(
        orderNotify: Boolean? = null,
        promoNotify: Boolean? = null,
        sound: Boolean? = null,
    ) {
        settings = settings.copy(
            orderNotify = orderNotify ?: settings.orderNotify,
            promoNotify = promoNotify ?: settings.promoNotify,
            sound = sound ?: settings.sound,
        )
        persist()
    }

    fun applicationOf(type: String): JoinApplication? =
        applications.lastOrNull { it.type == type }

    fun submitApplication(type: String, summary: String) {
        applications.removeAll { it.type == type }
        applications.add(
            JoinApplication(type, "审核中", System.currentTimeMillis(), summary)
        )
        persist()
    }

    fun submitFeedback(content: String) {
        feedbacks.add(Feedback(content, System.currentTimeMillis()))
        persist()
    }

    // ---------------- 展示辅助 ----------------

    fun greeting(): String {
        val hour = Calendar.getInstance().get(Calendar.HOUR_OF_DAY)
        return when {
            hour < 6 -> "凌晨好"
            hour < 11 -> "早上好"
            hour < 14 -> "中午好"
            hour < 18 -> "下午好"
            else -> "晚上好"
        }
    }

    fun displayPhone(): String = maskPhone(currentUser.phone)

    fun updateUserName(name: String) {
        currentUser = currentUser.copy(name = name)
        if (session() != null) {
            sessionPrefs().edit().putString("name", name).apply()
        }
        persist()
    }

    fun shortLocation(): String {
        val address = defaultAddress() ?: return "王家湾村"
        val first = address.detail.split("\n").firstOrNull().orEmpty()
        return when {
            address.detail.contains("王家湾村") -> "王家湾村三组 · 红砖大铁门"
            first.contains("浠水县") && address.detail.contains("\n") ->
                address.detail.split("\n").getOrElse(1) { first }.take(18)
            first.isNotBlank() -> first
            else -> "王家湾村"
        }
    }

    private fun todayStamp(): String =
        "今天 " + SimpleDateFormat("HH:mm", Locale.CHINA).format(Date())
}
