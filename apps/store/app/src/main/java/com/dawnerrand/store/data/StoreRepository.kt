package com.dawnerrand.store.data

import android.graphics.Color
import com.dawnerrand.store.R

// ---------------- 店铺 ----------------
data class ShopInfo(
    val name: String,
    val desc: String,
    val statusText: String,          // 营业中 / 已歇业
    val open: Boolean,
    val todayRevenue: String,        // ¥1,286
    val todayOrders: String,         // 42
    val goodRate: String,            // 98%
    val monthlySales: String,        // 月售856单
    val goodScore: String,           // 4.8分
    val reviewCount: String,         // 326条
)

// ---------------- 订单 ----------------
const val ORDER_PENDING = "pending"      // 待接单
const val ORDER_COOKING = "cooking"      // 制作中
const val ORDER_DELIVERING = "delivering" // 配送中
const val ORDER_DONE = "done"           // 已完成

data class StoreOrder(
    val id: String,
    val status: String,
    val customerName: String,
    val customerPhone: String,
    val address: String,
    val items: String,
    val total: String,
    val actions: List<String>,
)

// ---------------- 菜品 ----------------
data class Dish(
    val id: String,
    val name: String,
    val monthlySales: String,
    val stockText: String,
    val price: String,
    val iconRes: Int,
    val iconColor: Int,
    val bgRes: Int,
    val onSale: Boolean,
    val desc: String = "",
)

data class DishCategory(
    val name: String,
    val dishes: List<Dish>,
)

// ---------------- 营业设置 ----------------
data class BusinessSetting(
    val open: Boolean,
    val hours: String,
    val autoAccept: Boolean,
    val minPrice: String,
    val deliveryRange: String,
    val deliveryFee: String,
    val announcement: String,
    val promotion: String,
)

// ---------------- 财务 ----------------
data class FinanceRecord(
    val title: String,
    val time: String,
    val amount: String,
    val incoming: Boolean,
)

// ---------------- 经营数据 ----------------
data class StatOverview(val revenue: String, val orders: String, val avgOrder: String)
data class StatTrend(val labels: List<String>, val values: List<Float>)
data class RankItem(val rank: Int, val name: String, val sales: String)

// ---------------- 工作台宫格 ----------------
data class GridItem(
    val label: String,
    val iconRes: Int,
    val iconColor: Int,
    val bgRes: Int,
    val badge: Int = 0,
)

// ---------------- 商家中心菜单 ----------------
data class ProfileMenuItem(
    val label: String,
    val value: String?,
    val iconRes: Int,
    val iconColor: Int,
)

/**
 * 演示业务数据仓库：店铺/订单/菜品/营业/财务/统计 + 登录会话。
 * 后端商家接口就绪后替换为 ApiClient 调用，视图层零改动。
 */
object StoreRepository {

    // ---------- 宫格/菜单图标色（初始化顺序靠前，供下方列表引用） ----------
    private val DesignOrange = 0xFFFF6B2C.toInt()
    private val DesignBlue = 0xFF2B7FFF.toInt()
    private val DesignGreen = 0xFF00B578.toInt()
    private val DesignGray = 0xFF86909C.toInt()
    private val DesignGold = 0xFFD48806.toInt()
    private val DesignRed = 0xFFF5222D.toInt()
    private val DesignPurple = 0xFF722ED1.toInt()

    // ---------- 店铺（登录后由真实 /api/app/store/me 覆盖） ----------
    var shop = ShopInfo(
        name = "村口老陈家常菜",
        desc = "中餐 · 炒菜 · 月售856单",
        statusText = "营业中",
        open = true,
        todayRevenue = "¥1,286",
        todayOrders = "42",
        goodRate = "98%",
        monthlySales = "月售856单",
        goodScore = "4.8分",
        reviewCount = "326条",
    )

    // ---------- 订单（orders / order-detail 稿） ----------
    val orders = listOf(
        StoreOrder(
            "20261022001", ORDER_PENDING, "王建国", "138****8888",
            "王家湾村三组26号 · 138****8888",
            "农家小炒肉×1、西红柿炒蛋×1、米饭×2", "¥32.3", listOf("拒单", "接单")),
        StoreOrder(
            "20261022002", ORDER_PENDING, "李秀英", "137****6666",
            "王家湾村五组 · 137****6666",
            "红烧土鸡块×1、麻婆豆腐×1、米饭×2", "¥42.5", listOf("拒单", "接单")),
        StoreOrder(
            "20261022000", ORDER_COOKING, "张大山", "135****9999",
            "绿杨乡中学 · 135****9999",
            "手撕包菜×1、西红柿炒蛋×1、米饭×3", "¥28.0", listOf("通知出餐")),
        StoreOrder(
            "20261022006", ORDER_COOKING, "刘春梅", "139****2222",
            "王家湾村一组 · 139****2222",
            "农家小炒肉×2、米饭×2", "¥37.6", listOf("通知出餐")),
        StoreOrder(
            "20261022007", ORDER_COOKING, "赵铁柱", "136****1111",
            "绿杨乡卫生院 · 136****1111",
            "红烧土鸡块×1、手撕包菜×1、米饭×2", "¥54.0", listOf("通知出餐")),
        StoreOrder(
            "20261021998", ORDER_DELIVERING, "孙秀兰", "138****3333",
            "兰溪中学门口 · 138****3333",
            "麻婆豆腐×1、西红柿炒蛋×1、米饭×3", "¥35.0", listOf("查看配送")),
        StoreOrder(
            "20261021997", ORDER_DELIVERING, "周建国", "137****7777",
            "镇农贸市场 · 137****7777",
            "红烧土鸡块×2、米饭×1", "¥58.0", listOf("查看配送")),
        StoreOrder(
            "20261021990", ORDER_DONE, "吴翠花", "135****5555",
            "王家湾村二组 · 135****5555",
            "农家小炒肉×1、米饭×2", "¥20.8", listOf("再来一单")),
        StoreOrder(
            "20261021985", ORDER_DONE, "郑有才", "139****8888",
            "绿杨乡政府 · 139****8888",
            "西红柿炒蛋×1、手撕包菜×1、米饭×2", "¥26.0", listOf("再来一单")),
        StoreOrder(
            "20261021980", ORDER_DONE, "钱大壮", "136****4444",
            "镇中心学校 · 136****4444",
            "麻婆豆腐×2、米饭×4", "¥30.0", listOf("再来一单")),
    )

    /** 订单详情：明细（order-detail 稿） */
    val orderItems = mapOf(
        "20261022001" to listOf(
            Triple("农家小炒肉 ×1", "¥16.8", "少辣，米饭要热一点"),
            Triple("西红柿炒蛋 ×1", "¥10.0", null),
            Triple("米饭 ×2", "¥4.0", null),
        ),
        "20261022002" to listOf(
            Triple("红烧土鸡块 ×1", "¥28.0", "多放点葱"),
            Triple("麻婆豆腐 ×1", "¥11.0", null),
            Triple("米饭 ×2", "¥4.0", null),
        ),
    )

    // ---------- 菜品（menu / edit-dish 稿） ----------
    private val hotDishes = listOf(
        Dish("d1", "农家小炒肉", "月售86", "库存充足", "16.8",
            R.drawable.ic_shop, Color.rgb(0xFF, 0x8A, 0x3D), R.drawable.bg_dish_1, true),
        Dish("d2", "红烧土鸡块", "月售64", "库存仅剩8份", "28.0",
            R.drawable.ic_smile, Color.rgb(0xFF, 0x6B, 0x2C), R.drawable.bg_dish_2, true),
        Dish("d3", "手撕包菜", "月售52", "已售罄", "12.0",
            R.drawable.ic_veg, Color.rgb(0x52, 0xC4, 0x1A), R.drawable.bg_dish_3, false),
        Dish("d4", "西红柿炒蛋", "月售78", "库存充足", "10.0",
            R.drawable.ic_egg, Color.rgb(0xFA, 0xAD, 0x14), R.drawable.bg_dish_4, true),
        Dish("d5", "麻婆豆腐", "月售45", "库存充足", "11.0",
            R.drawable.ic_flame, Color.rgb(0xF5, 0x22, 0x2D), R.drawable.bg_dish_5, true),
    )
    private val soupDishes = listOf(
        Dish("d6", "冬瓜排骨汤", "月售38", "库存充足", "18.0",
            R.drawable.ic_flame, Color.rgb(0xFF, 0x6B, 0x2C), R.drawable.bg_dish_1, true),
        Dish("d7", "紫菜蛋花汤", "月售26", "库存充足", "8.0",
            R.drawable.ic_egg, Color.rgb(0xFA, 0xAD, 0x14), R.drawable.bg_dish_4, true),
    )
    private val stapleDishes = listOf(
        Dish("d8", "白米饭", "月售420", "库存充足", "2.0",
            R.drawable.ic_veg, Color.rgb(0x52, 0xC4, 0x1A), R.drawable.bg_dish_3, true),
        Dish("d9", "手工馒头", "月售180", "库存充足", "1.5",
            R.drawable.ic_smile, Color.rgb(0xFF, 0x8A, 0x3D), R.drawable.bg_dish_2, true),
    )

    val dishCategories = listOf(
        DishCategory("热销爆款", hotDishes),
        DishCategory("招牌炒菜", listOf(hotDishes[0], hotDishes[1], hotDishes[4])),
        DishCategory("家常小炒", listOf(hotDishes[2], hotDishes[3])),
        DishCategory("汤类炖菜", soupDishes),
        DishCategory("主食", stapleDishes),
    )

    // ---------- 营业设置（business 稿） ----------
    val business = BusinessSetting(
        open = true,
        hours = "09:00 - 21:00",
        autoAccept = true,
        minPrice = "¥15",
        deliveryRange = "5公里",
        deliveryFee = "¥3/单",
        announcement = "今日辣椒用完，少辣请备注",
        promotion = "满25减3",
    )

    // ---------- 财务（finance 稿） ----------
    val balance = "3,682.50"
    val monthRevenue = "¥8,940"
    val monthOrders = "128单"

    val financeRecords = listOf(
        FinanceRecord("订单 #20261022001 收入", "今天 12:30", "+¥30.8", true),
        FinanceRecord("订单 #20261021156 收入", "今天 11:45", "+¥45.0", true),
        FinanceRecord("提现到微信", "10月20日 09:00", "-¥2,000.0", false),
        FinanceRecord("平台服务费扣款", "10月20日 00:00", "-¥89.4", false),
    )

    // ---------- 经营数据（statistics 稿） ----------
    val statOverview = StatOverview("¥1,286", "42", "¥30.6")
    val statTrend = StatTrend(
        labels = listOf("周三", "周四", "周五", "周六", "周日", "周一", "今天"),
        values = listOf(0.50f, 0.65f, 0.45f, 0.85f, 0.95f, 0.55f, 0.70f),
    )
    val statRanks = listOf(
        RankItem(1, "农家小炒肉", "86份"),
        RankItem(2, "西红柿炒蛋", "78份"),
        RankItem(3, "红烧土鸡块", "64份"),
        RankItem(4, "麻婆豆腐", "45份"),
        RankItem(5, "手撕包菜", "52份"),
    )

    // ---------- 工作台宫格（dashboard 稿） ----------
    val orderGrid = listOf(
        GridItem("待接单", R.drawable.ic_clock, DesignOrange, R.drawable.bg_icon_orange, 2),
        GridItem("制作中", R.drawable.ic_doc, DesignBlue, R.drawable.bg_icon_blue, 3),
        GridItem("配送中", R.drawable.ic_bike, DesignGreen, R.drawable.bg_icon_green),
        GridItem("已完成", R.drawable.ic_check_done, DesignGray, R.drawable.bg_icon_gray),
    )
    val shopGrid = listOf(
        GridItem("菜品管理", R.drawable.ic_layers, DesignGold, R.drawable.bg_icon_gold),
        GridItem("营业设置", R.drawable.ic_alert, DesignRed, R.drawable.bg_icon_red),
        GridItem("收入提现", R.drawable.ic_money, DesignGreen, R.drawable.bg_icon_green),
        GridItem("经营数据", R.drawable.ic_chart, DesignPurple, R.drawable.bg_icon_purple),
    )

    // ---------- 商家中心菜单（profile 稿） ----------
    val profileMenus = listOf(
        ProfileMenuItem("资质证照", "已认证", R.drawable.ic_award, 0xFF2B7FFF.toInt()),
        ProfileMenuItem("结算账户", "微信钱包", R.drawable.ic_card, 0xFF00B578.toInt()),
        ProfileMenuItem("经营数据", null, R.drawable.ic_chart, 0xFFFF6B2C.toInt()),
        ProfileMenuItem("顾客评价", "4.8分 · 326条", R.drawable.ic_star, 0xFF722ED1.toInt()),
        ProfileMenuItem("商家客服", null, R.drawable.ic_comment, 0xFFFA8C16.toInt()),
        ProfileMenuItem("设置", null, R.drawable.ic_settings, 0xFF86909C.toInt()),
    )
}
