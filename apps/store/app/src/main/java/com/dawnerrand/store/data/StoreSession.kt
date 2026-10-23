package com.dawnerrand.store.data

import android.content.Context
import android.content.SharedPreferences

/**
 * 商家登录会话：真实后端 token（邮箱密码登录）或演示 token（手机验证码登录）。
 * 仅保存登录态与展示信息，业务数据仍来自演示仓库。
 */
object StoreSession {

    private const val PREF = "store_session"

    private lateinit var prefs: SharedPreferences

    var token: String? = null
        private set
    var name: String = "陈老板"
        private set
    var account: String = ""
        private set
    var role: String = ""
        private set
    var phone: String = "138 8888 8888"
        private set
    /** true 表示通过真实 /api/auth/login 登录 */
    var realBackend: Boolean = false
        private set

    fun init(ctx: Context) {
        if (::prefs.isInitialized) return
        prefs = ctx.applicationContext.getSharedPreferences(PREF, Context.MODE_PRIVATE)
        token = prefs.getString("token", null)
        name = prefs.getString("name", null) ?: "陈老板"
        account = prefs.getString("account", null) ?: ""
        role = prefs.getString("role", null) ?: ""
        phone = prefs.getString("phone", null) ?: "138 8888 8888"
        realBackend = prefs.getBoolean("real", false)
    }

    fun isLoggedIn(): Boolean = !token.isNullOrEmpty()

    fun save(ctx: Context, token: String, name: String, account: String, role: String, phone: String, real: Boolean) {
        init(ctx)
        this.token = token
        this.name = name.ifBlank { "商家店主" }
        this.account = account
        this.role = role
        this.phone = phone
        this.realBackend = real
        prefs.edit()
            .putString("token", token)
            .putString("name", this.name)
            .putString("account", account)
            .putString("role", role)
            .putString("phone", phone)
            .putBoolean("real", real)
            .apply()
    }

    fun logout(ctx: Context) {
        init(ctx)
        prefs.edit().clear().apply()
        token = null
        name = "陈老板"
        account = ""
        role = ""
        phone = "138 8888 8888"
        realBackend = false
    }

    /** 账号展示：真实登录用邮箱/手机号，演示登录用手机号 */
    fun accountLabel(): String = account.ifBlank { phone }

    fun roleLabel(): String = when (role) {
        "merchant" -> "商家店主"
        "merchant_admin" -> "商家管理员"
        "super_admin" -> "平台管理员"
        "driver" -> "配送骑手"
        "" -> if (realBackend) "商家账号" else "演示账号"
        else -> "商家账号"
    }
}
