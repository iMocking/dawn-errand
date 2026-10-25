package com.dawnerrand.store.data

import android.content.Context
import android.content.SharedPreferences
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import org.json.JSONArray
import org.json.JSONObject
import java.util.Calendar
import java.util.Locale

/**
 * 收入与提现：余额、近期收支记录、提现闭环（扣减余额 + 追加记录）。
 * 本地持久化，重启后保留。
 */
object FinanceStore {

    private const val PREF = "store_finance"
    private const val FEE_RATE = 0.006

    private lateinit var prefs: SharedPreferences

    private val _balance = MutableLiveData(3682.50)
    val balance: LiveData<Double> = _balance

    private val _records = MutableLiveData(StoreRepository.financeRecords)
    val records: LiveData<List<FinanceRecord>> = _records

    data class WithdrawResult(val ok: Boolean, val message: String)

    fun init(ctx: Context) {
        if (::prefs.isInitialized) return
        prefs = ctx.applicationContext.getSharedPreferences(PREF, Context.MODE_PRIVATE)
        _balance.value = prefs.getString("balance", null)?.toDoubleOrNull() ?: 3682.50
        val raw = prefs.getString("records", null) ?: return
        try {
            val arr = JSONArray(raw)
            val list = mutableListOf<FinanceRecord>()
            for (i in 0 until arr.length()) {
                val o = arr.getJSONObject(i)
                list += FinanceRecord(
                    title = o.optString("title"),
                    time = o.optString("time"),
                    amount = o.optString("amount"),
                    incoming = o.optBoolean("incoming"),
                )
            }
            if (list.isNotEmpty()) _records.value = list
        } catch (e: Exception) {
            // 解析失败保留演示数据
        }
    }

    fun balanceText(): String = formatMoney(_balance.value ?: 0.0)

    /** 用后端 summary/records 覆盖余额与流水。 */
    fun replaceFromApi(balance: Double, records: List<FinanceRecord>?) {
        _balance.postValue(balance)
        if (records != null) _records.postValue(records)
        persist()
    }

    fun feeRateText(): String = "0.6%"

    /** 手续费（0.6%） */
    fun feeOf(amount: Double): Double = amount * FEE_RATE

    /**
     * 发起提现：金额需 > 0 且不超过可提现余额。
     */
    fun withdraw(ctx: Context, amountText: String): WithdrawResult {
        init(ctx)
        val amount = amountText.trim().removePrefix("¥").toDoubleOrNull()
            ?: return WithdrawResult(false, "请输入正确的提现金额")
        val balance = _balance.value ?: 0.0
        if (amount < 1.0) return WithdrawResult(false, "单次提现金额不能低于 ¥1.00")
        if (amount > balance) return WithdrawResult(false, "提现金额超过可提现余额")
        _balance.value = balance - amount
        val now = Calendar.getInstance()
        val time = String.format(
            Locale.CHINA, "今天 %02d:%02d", now.get(Calendar.HOUR_OF_DAY), now.get(Calendar.MINUTE)
        )
        val account = SettlementStore.summary(ctx)
        val record = FinanceRecord(
            title = "提现到$account",
            time = time,
            amount = "-¥${formatMoney(amount)}",
            incoming = false,
        )
        _records.value = listOf(record) + (_records.value ?: emptyList())
        persist()
        val arrive = amount - feeOf(amount)
        return WithdrawResult(
            true,
            "提现申请已提交，扣除 ${feeRateText()} 手续费后预计到账 ¥${formatMoney(arrive)}",
        )
    }

    /** 千分位 + 两位小数，与设计稿 ¥3,682.50 一致 */
    fun formatMoney(v: Double): String {
        val s = String.format(Locale.CHINA, "%,.2f", v)
        return s
    }

    private fun persist() {
        if (!::prefs.isInitialized) return
        val arr = JSONArray()
        (_records.value ?: emptyList()).forEach { r ->
            arr.put(JSONObject().apply {
                put("title", r.title)
                put("time", r.time)
                put("amount", r.amount)
                put("incoming", r.incoming)
            })
        }
        prefs.edit()
            .putString("balance", (_balance.value ?: 0.0).toString())
            .putString("records", arr.toString())
            .apply()
    }
}
