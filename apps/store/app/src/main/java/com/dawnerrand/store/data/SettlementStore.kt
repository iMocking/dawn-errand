package com.dawnerrand.store.data

import android.content.Context
import android.content.SharedPreferences
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData

/**
 * 结算账户：微信钱包（默认）或银行卡（可绑定/更换/解绑）。
 * 提现按钮文案与提现记录跟随当前结算账户。
 */
object SettlementStore {

    private const val PREF = "store_settlement"
    private lateinit var prefs: SharedPreferences

    /** "wechat" 或 "bank" */
    var type: String = "wechat"
        private set
    var bankName: String = ""
        private set
    var cardNo: String = ""
        private set
    var holder: String = ""
        private set

    private val _revision = MutableLiveData(0)
    val revision: LiveData<Int> = _revision

    fun init(ctx: Context) {
        if (::prefs.isInitialized) return
        prefs = ctx.applicationContext.getSharedPreferences(PREF, Context.MODE_PRIVATE)
        type = prefs.getString("type", "wechat") ?: "wechat"
        bankName = prefs.getString("bankName", "") ?: ""
        cardNo = prefs.getString("cardNo", "") ?: ""
        holder = prefs.getString("holder", "") ?: ""
        if (type == "bank" && bankName.isBlank()) type = "wechat"
    }

    fun hasBank(): Boolean = bankName.isNotBlank() && cardNo.isNotBlank()

    fun summary(ctx: Context): String {
        init(ctx)
        return if (type == "bank" && hasBank()) "$bankName · 尾号${last4()}" else "微信钱包"
    }

    /** 提现按钮文案：提现到微信钱包 / 提现到招商银行 · 尾号8888 */
    fun withdrawTarget(ctx: Context): String {
        init(ctx)
        return if (type == "bank" && hasBank()) "提现到$bankName · 尾号${last4()}" else "提现到微信钱包"
    }

    fun maskedCard(): String {
        val digits = cardNo.filter { it.isDigit() }
        if (digits.length < 4) return "**** **** **** ****"
        return "**** **** **** ${digits.takeLast(4)}"
    }

    fun last4(): String = cardNo.filter { it.isDigit() }.takeLast(4).ifBlank { "8888" }

    fun saveBank(ctx: Context, bankName: String, cardNo: String, holder: String): String? {
        init(ctx)
        val digits = cardNo.filter { it.isDigit() }
        if (holder.isBlank()) return "请填写持卡人姓名"
        if (bankName.isBlank()) return "请填写开户银行"
        if (digits.length < 12) return "银行卡号至少 12 位"
        this.bankName = bankName.trim()
        this.cardNo = digits
        this.holder = holder.trim()
        this.type = "bank"
        persist()
        _revision.value = (_revision.value ?: 0) + 1
        return null
    }

    fun useWechat(ctx: Context) {
        init(ctx)
        type = "wechat"
        persist()
        _revision.value = (_revision.value ?: 0) + 1
    }

    fun useBank(ctx: Context) {
        init(ctx)
        if (!hasBank()) return
        type = "bank"
        persist()
        _revision.value = (_revision.value ?: 0) + 1
    }

    fun removeBank(ctx: Context) {
        init(ctx)
        bankName = ""
        cardNo = ""
        holder = ""
        type = "wechat"
        persist()
        _revision.value = (_revision.value ?: 0) + 1
    }

    fun revisionNow(): Int = _revision.value ?: 0

    private fun persist() {
        if (!::prefs.isInitialized) return
        prefs.edit()
            .putString("type", type)
            .putString("bankName", bankName)
            .putString("cardNo", cardNo)
            .putString("holder", holder)
            .apply()
    }
}
