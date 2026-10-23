package com.dawnerrand.driver.data.repository

import android.content.Context
import org.json.JSONArray
import org.json.JSONObject

/**
 * 骑手卡包：健康证 / 身份证（唯一，模拟证件卡样式 + 有效期起止）+ 银行卡（多卡，含有效期）。
 * 本地持久化于 SharedPreferences("driver_cards")，收入提现闭环依赖银行卡数量。
 */
data class CertCard(
    val name: String,
    val certNo: String,
    val validFrom: String,
    val validUntil: String,
)

data class BankCard(
    val holder: String,
    val bank: String,
    val cardNo: String,
    val expiry: String,
)

object CardWalletRepository {

    private const val PREFS = "driver_cards"
    private const val KEY_HEALTH = "health_cert"
    private const val KEY_ID = "id_cert"
    private const val KEY_BANKS = "bank_cards"

    // 唯一证件：默认已认证（与 rider 画像一致），有效期起止
    var healthCert: CertCard = CertCard("张师傅", "NO.2024-8821", "2024-03-01", "2027-02-28")
    var idCert: CertCard = CertCard("张师傅", "4211 **** **** 5821", "2020-01-15", "2040-01-15")

    // 银行卡：默认 1 张，支持增删，含有效期（MM/YY）
    val bankCards: MutableList<BankCard> = mutableListOf(
        BankCard("张师傅", "招商银行", "6225 8801 **** 8888", "08/28"),
    )

    fun load(context: Context) {
        val sp = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
        sp.getString(KEY_HEALTH, null)?.let {
            runCatching { parseCert(it, healthCert) }.getOrNull()?.let { c -> healthCert = c }
        }
        sp.getString(KEY_ID, null)?.let {
            runCatching { parseCert(it, idCert) }.getOrNull()?.let { c -> idCert = c }
        }
        bankCards.clear()
        val banksJson = sp.getString(KEY_BANKS, null)
        if (banksJson != null) {
            runCatching {
                val arr = JSONArray(banksJson)
                for (i in 0 until arr.length()) {
                    val o = arr.getJSONObject(i)
                    bankCards.add(
                        BankCard(
                            o.getString("holder"),
                            o.getString("bank"),
                            o.getString("cardNo"),
                            o.optString("expiry", ""),
                        )
                    )
                }
            }.getOrNull()
        } else {
            // 首次启动：恢复默认卡（保证提现闭环可走通）
            bankCards.add(BankCard("张师傅", "招商银行", "6225 8801 **** 8888", "08/28"))
        }
    }

    fun save(context: Context) {
        val editor = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE).edit()
        editor.putString(KEY_HEALTH, certToJson(healthCert).toString())
        editor.putString(KEY_ID, certToJson(idCert).toString())
        val arr = JSONArray()
        bankCards.forEach { c ->
            arr.put(
                JSONObject()
                    .put("holder", c.holder)
                    .put("bank", c.bank)
                    .put("cardNo", c.cardNo)
                    .put("expiry", c.expiry)
            )
        }
        editor.putString(KEY_BANKS, arr.toString())
        editor.apply()
    }

    fun addBank(context: Context, card: BankCard) {
        bankCards.add(card)
        save(context)
    }

    fun removeBank(context: Context, index: Int) {
        if (index in bankCards.indices) {
            bankCards.removeAt(index)
            save(context)
        }
    }

    private fun certToJson(c: CertCard) =
        JSONObject()
            .put("name", c.name)
            .put("certNo", c.certNo)
            .put("validFrom", c.validFrom)
            .put("validUntil", c.validUntil)

    private fun parseCert(json: String, default: CertCard): CertCard {
        val o = JSONObject(json)
        // 兼容旧版本持久化数据（无有效期字段）：缺失时按各自证件类型回退默认值
        val from = o.optString("validFrom", "")
        val until = o.optString("validUntil", "")
        return CertCard(
            o.getString("name"),
            o.getString("certNo"),
            if (from.isEmpty()) default.validFrom else from,
            if (until.isEmpty()) default.validUntil else until,
        )
    }
}
