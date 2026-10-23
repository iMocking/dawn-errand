package com.dawnerrand.store.data

import android.content.Context
import android.content.SharedPreferences
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData

/**
 * 营业设置（可编辑 + 本地持久化）：营业开关/自动接单/营业时间/起送价/配送范围/配送费/公告/满减。
 * 工作台、营业设置页共享同一份数据，修改后通过 revision 联动刷新。
 */
object StoreSettings {

    private const val PREF = "store_settings"
    private lateinit var prefs: SharedPreferences

    var open: Boolean = true
        private set
    var autoAccept: Boolean = true
        private set
    var hours: String = "09:00 - 21:00"
        private set
    var minPrice: String = "¥15"
        private set
    var deliveryRange: String = "5公里"
        private set
    var deliveryFee: String = "¥3/单"
        private set
    var announcement: String = "今日辣椒用完，少辣请备注"
        private set
    var promotion: String = "满25减3"
        private set

    /** 订单通知 / 系统通知（设置页） */
    var orderNotify: Boolean = true
        private set
    var systemNotify: Boolean = true
        private set

    private val _revision = MutableLiveData(0)
    val revision: LiveData<Int> = _revision

    fun init(ctx: Context) {
        if (::prefs.isInitialized) return
        prefs = ctx.applicationContext.getSharedPreferences(PREF, Context.MODE_PRIVATE)
        val d = StoreRepository.business
        open = prefs.getBoolean("open", d.open)
        autoAccept = prefs.getBoolean("autoAccept", d.autoAccept)
        hours = prefs.getString("hours", d.hours) ?: d.hours
        minPrice = prefs.getString("minPrice", d.minPrice) ?: d.minPrice
        deliveryRange = prefs.getString("deliveryRange", d.deliveryRange) ?: d.deliveryRange
        deliveryFee = prefs.getString("deliveryFee", d.deliveryFee) ?: d.deliveryFee
        announcement = prefs.getString("announcement", d.announcement) ?: d.announcement
        promotion = prefs.getString("promotion", d.promotion) ?: d.promotion
        orderNotify = prefs.getBoolean("orderNotify", true)
        systemNotify = prefs.getBoolean("systemNotify", true)
    }

    fun setOpen(ctx: Context, value: Boolean) = update(ctx, "open") {
        open = value
        prefs.edit().putBoolean("open", value).apply()
    }

    fun setAutoAccept(ctx: Context, value: Boolean) = update(ctx, "autoAccept") {
        autoAccept = value
        prefs.edit().putBoolean("autoAccept", value).apply()
    }

    fun setHours(ctx: Context, value: String) = update(ctx, "hours") {
        hours = value
        prefs.edit().putString("hours", value).apply()
    }

    fun setMinPrice(ctx: Context, value: String) = update(ctx, "minPrice") {
        minPrice = value
        prefs.edit().putString("minPrice", value).apply()
    }

    fun setDeliveryRange(ctx: Context, value: String) = update(ctx, "deliveryRange") {
        deliveryRange = value
        prefs.edit().putString("deliveryRange", value).apply()
    }

    fun setDeliveryFee(ctx: Context, value: String) = update(ctx, "deliveryFee") {
        deliveryFee = value
        prefs.edit().putString("deliveryFee", value).apply()
    }

    fun setAnnouncement(ctx: Context, value: String) = update(ctx, "announcement") {
        announcement = value
        prefs.edit().putString("announcement", value).apply()
    }

    fun setPromotion(ctx: Context, value: String) = update(ctx, "promotion") {
        promotion = value
        prefs.edit().putString("promotion", value).apply()
    }

    fun setOrderNotify(ctx: Context, value: Boolean) {
        orderNotify = value
        prefs.edit().putBoolean("orderNotify", value).apply()
    }

    fun setSystemNotify(ctx: Context, value: Boolean) {
        systemNotify = value
        prefs.edit().putBoolean("systemNotify", value).apply()
    }

    /** 营业结束时间（用于工作台状态文案），从 hours 中解析 */
    fun closeTime(): String = hours.split("-").lastOrNull()?.trim().orEmpty().ifBlank { "21:00" }

    fun toBusinessSetting(): BusinessSetting = BusinessSetting(
        open = open,
        hours = hours,
        autoAccept = autoAccept,
        minPrice = minPrice,
        deliveryRange = deliveryRange,
        deliveryFee = deliveryFee,
        announcement = announcement,
        promotion = promotion,
    )

    private inline fun update(ctx: Context, key: String, block: () -> Unit) {
        init(ctx)
        block()
        _revision.value = (_revision.value ?: 0) + 1
    }
}
