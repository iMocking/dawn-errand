package com.dawnerrand.client.data

import android.content.Context
import android.graphics.Color
import com.dawnerrand.client.R

/** 商家 */
data class Merchant(
    val id: String,
    val name: String,
    val recName: String, // 推荐横滑卡片用短名
    val score: String,
    val distance: String,
    val monthlySales: String,
    val deliveryMinutes: String,
    val tags: List<String>,
    val freeTag: String?,
    val bannerColorStart: Int,
    val bannerColorEnd: Int,
    val iconRes: Int,
    val iconColor: Int,
    val freeDeliveryThreshold: Int = 25,
)

/** 菜品 */
data class Dish(
    val id: String,
    val name: String,
    val desc: String,
    val price: Double,
    val monthlySales: String,
    val iconRes: Int,
    val iconColor: Int,
    val bgColor: Int,
)

/** 购物车项 */
data class CartItem(val dish: Dish, var count: Int)

/** 订单 */
data class Order(
    val shop: String,
    val status: String,
    val ongoing: Boolean,
    val items: String,
    val time: String,
    val total: String,
    val iconRes: Int,
    val iconColor: Int,
    val iconBg: Int,
    val actions: List<String>,
)

/** 地址 */
data class Address(
    val name: String,
    val phone: String,
    val detail: String,
    val tags: List<String>,
    val isDefault: Boolean = false,
)

/** 优惠券 */
data class Coupon(
    val amount: String,
    val cond: String,
    val title: String,
    val desc: String,
    val expire: String,
    val disabled: Boolean = false,
)

/**
 * 演示业务数据仓库：商家/菜品/订单/地址/优惠券 + 购物车 + 会话。
 * 后端 C 端接口就绪后替换为 ApiClient 调用，视图层零改动。
 */
object DemoRepository {

    // ---------- 商家 ----------
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

    // ---------- 菜品（商家 m1 菜单，merchant-detail 稿） ----------
    val dishes = listOf(
        Dish("d1", "农家小炒肉", "招牌土猪肉现炒，香辣下饭 · 月售86份", 16.8, "86",
            R.drawable.ic_food_bag, Color.rgb(0xFF, 0x8A, 0x3D), Color.rgb(0xFF, 0xF0, 0xE0)),
        Dish("d2", "红烧土鸡块", "散养土鸡慢炖，入味软烂 · 月售64份", 28.0, "64",
            R.drawable.ic_food_meat, Color.rgb(0xFF, 0x6B, 0x2C), Color.rgb(0xFF, 0xE0, 0xD6)),
        Dish("d3", "手撕包菜", "本地包菜大火快炒，脆嫩爽口 · 月售52份", 12.0, "52",
            R.drawable.ic_food_veg, Color.rgb(0x52, 0xC4, 0x1A), Color.rgb(0xE8, 0xF5, 0xE9)),
        Dish("d4", "西红柿炒蛋", "农家土鸡蛋，酸甜可口 · 月售78份", 10.0, "78",
            R.drawable.ic_food_egg, Color.rgb(0xFA, 0xAD, 0x14), Color.rgb(0xFF, 0xF8, 0xE1)),
    )

    // ---------- 购物车 ----------
    val cart = mutableListOf<CartItem>()

    fun cartTotal(): Double = cart.sumOf { it.dish.price * it.count }

    fun cartCount(): Int = cart.sumOf { it.count }

    fun addDish(dish: Dish) {
        val it = cart.find { c -> c.dish.id == dish.id }
        if (it != null) it.count++ else cart.add(CartItem(dish, 1))
    }

    fun removeDish(dish: Dish) {
        val it = cart.find { c -> c.dish.id == dish.id } ?: return
        it.count--
        if (it.count <= 0) cart.remove(it)
    }

    // ---------- 订单（orders 稿） ----------
    val orders = listOf(
        Order("村口老陈家常菜", "配送中", true, "农家小炒肉、西红柿炒蛋、米饭×2",
            "今天 12:30 · 预计12:45送达", "¥32.3",
            R.drawable.ic_food_bag, Color.rgb(0xFF, 0x8A, 0x3D), Color.rgb(0xFF, 0xF0, 0xE0),
            listOf("取消订单", "查看配送")),
        Order("王家湾综合超市", "已完成", false, "纯牛奶×2箱、洗衣粉×1、抽纸×3",
            "昨天 18:20 · 已送达", "¥86.5",
            R.drawable.ic_food_box, Color.rgb(0x52, 0xC4, 0x1A), Color.rgb(0xE8, 0xF5, 0xE9),
            listOf("再来一单", "评价")),
        Order("跑腿代办 · 镇中心", "已完成", false, "帮取快递（镇邮政所→王家湾村）",
            "10月20日 15:10 · 已送达", "¥6.0",
            R.drawable.ic_tab_errand, Color.rgb(0x00, 0xB5, 0x78), Color.rgb(0xE8, 0xF7, 0xF0),
            listOf("再次下单")),
        Order("乡里乡亲水果店", "已完成", false, "红富士苹果5斤、香蕉1把、葡萄2斤",
            "10月18日 10:00 · 已送达", "¥45.8",
            R.drawable.ic_food_fruit, Color.rgb(0xEB, 0x2F, 0x96), Color.rgb(0xFC, 0xE4, 0xEC),
            listOf("再来一单", "评价")),
    )

    // ---------- 地址（address 稿） ----------
    val addresses = listOf(
        Address("王建国", "138****8888",
            "湖北省黄冈市浠水县\n王家湾村三组26号 · 红砖大铁门，门口有棵樟树",
            listOf("默认", "家"), true),
        Address("王秀兰", "137****6666",
            "湖北省黄冈市浠水县\n绿杨乡卫生院住院部 2楼3床", listOf("医院")),
        Address("王强", "135****9999",
            "湖北省黄冈市浠水县\n兰溪中学 高三(2)班宿舍", listOf("学校")),
    )

    // ---------- 优惠券（coupons 稿） ----------
    val coupons = listOf(
        Coupon("5", "满20可用", "新用户立减券", "全场通用 · 外卖跑腿均可使用", "有效期至 2026-10-29"),
        Coupon("3", "无门槛", "骑手好评奖励券", "完成订单评价后获得", "有效期至 2026-10-25"),
        Coupon("8", "满30可用", "周末专享券", "仅限周六周日使用", "已过期 2026-10-20", disabled = true),
    )

    // ---------- 会话 ----------
    private const val SESSION = "client_session"

    fun isLoggedIn(ctx: Context): Boolean =
        ctx.getSharedPreferences(SESSION, Context.MODE_PRIVATE).getString("token", null) != null

    fun saveLogin(ctx: Context, token: String, phone: String) {
        ctx.getSharedPreferences(SESSION, Context.MODE_PRIVATE)
            .edit().putString("token", token).putString("phone", phone).apply()
    }

    fun logout(ctx: Context) {
        ctx.getSharedPreferences(SESSION, Context.MODE_PRIVATE).edit().clear().apply()
    }
}
