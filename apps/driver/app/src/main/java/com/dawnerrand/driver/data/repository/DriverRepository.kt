package com.dawnerrand.driver.data.repository

import com.dawnerrand.driver.data.model.ActiveOrder
import com.dawnerrand.driver.data.model.FinanceRecord
import com.dawnerrand.driver.data.model.OrderItem
import com.dawnerrand.driver.data.model.OrderRecord
import com.dawnerrand.driver.data.model.RiderProfile

/**
 * 骑手端数据仓库：当前以演示数据还原设计稿（与 client 端约定一致）；
 * 后端骑手接口就绪后，可在各 ViewModel 中无缝替换为 ApiClient 调用，视图层零改动。
 */
object DriverRepository {

    var rider: RiderProfile = RiderProfile()

    var activeOrder = ActiveOrder(
        pickName = "村口老陈家常菜",
        pickAddr = "浠水县绿杨乡镇中心街 · 已出餐",
        pickContact = "陈老板 139****5555",
        dropName = "王建国",
        dropAddr = "王家湾村三组26号 · 红砖大铁门，门口有大樟树",
        dropContact = "138****8888",
        status = "待取货",
    )

    var orderPool: List<OrderItem> = listOf(
        OrderItem("外卖订单", false, "2.3km", "村口老陈家常菜（镇中心）", "王家湾村三组26号", "¥6.5", "配送费", 2.3f),
        OrderItem("帮我买", false, "1.8km", "镇中心供销社超市", "绿杨乡中学宿舍3栋201", "¥8.0", "跑腿费", 1.8f),
        OrderItem("帮我取", true, "3.5km", "镇邮政代办点（取快递）", "王家湾村五组李秀英家", "¥5.0", "取件费", 3.5f),
        OrderItem("外卖订单", false, "4.1km", "乡里乡亲水果店", "王家湾村七组·红砖大屋", "¥7.0", "配送费", 4.1f),
    )

    var orderHistory: List<OrderRecord> = listOf(
        OrderRecord("今天 12:35 完成", "已送达", "村口老陈家常菜", "王家湾村三组26号", "外卖配送 · 2.3km", "+¥7.0"),
        OrderRecord("今天 11:20 完成", "已送达", "镇中心供销社超市", "绿杨乡中学宿舍3栋", "帮我买 · 1.8km", "+¥8.5"),
        OrderRecord("今天 10:05 完成", "已送达", "乡里乡亲水果店", "王家湾村七组", "外卖配送 · 4.1km", "+¥7.0"),
        OrderRecord("昨天 18:40 完成", "已送达", "镇中心大药房", "王家湾村五组", "帮我送 · 3.5km", "+¥6.0"),
        OrderRecord("昨天 17:15 完成", "已送达", "镇邮政代办点", "王家湾村三组", "帮我取快递 · 2.8km", "+¥5.0"),
    )

    var financeRecords: List<FinanceRecord> = listOf(
        FinanceRecord("配送订单 #001 收入", "12:35 · 村口老陈→王家湾", "+¥7.0", true),
        FinanceRecord("帮我买订单 收入", "11:20 · 供销社超市→绿杨中学", "+¥8.5", true),
        FinanceRecord("配送订单 #000 收入", "10:05 · 乡里乡亲水果店→王家湾", "+¥7.0", true),
        FinanceRecord("提现到微信", "昨天 20:00", "-¥500.0", false),
    )
}
