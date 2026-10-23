package com.dawnerrand.driver.data.api

import com.dawnerrand.driver.BuildConfig
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL

/**
 * 轻量 API 客户端（HttpURLConnection，无第三方依赖）。
 * 对接后端统一返回 {code, message, body, request_id}（code=0 成功）。
 * 邮箱/密码走真实接口 /api/auth/login；短信接口（/api/auth/sms 系列）后端未提供时会自动进入演示模式（任意 6 位验证码可登录）。
 */
object ApiClient {

    var baseUrl: String = BuildConfig.API_BASE_URL

    /** 演示模式：后端不可达或未提供短信接口时置为 true */
    var demoMode: Boolean = false

    /** 最近一次真实登录成功后的访问令牌（邮箱/密码登录） */
    var accessToken: String? = null

    private const val TIMEOUT_MS = 4000

    enum class SmsStatus { OK, DEMO_FALLBACK }

    /** 邮箱/密码登录结果（真实接口 /api/auth/login） */
    data class PasswordLogin(val code: Int, val message: String, val token: String?)

    fun sendSmsCode(phone: String): SmsStatus {
        return try {
            val resp = post("/api/auth/sms/send", """{"phone":"$phone"}""")
            when {
                resp == null -> { demoMode = true; SmsStatus.DEMO_FALLBACK }
                resp.code == 0 -> SmsStatus.OK
                else -> { demoMode = true; SmsStatus.DEMO_FALLBACK }
            }
        } catch (e: Exception) {
            demoMode = true
            SmsStatus.DEMO_FALLBACK
        }
    }

    fun smsLogin(phone: String, code: String): Boolean {
        return try {
            val resp = post("/api/auth/sms/login", """{"phone":"$phone","code":"$code"}""")
            resp?.code == 0
        } catch (e: Exception) {
            false
        }
    }

    /**
     * 邮箱/密码登录：直连真实接口 POST /api/auth/login。
     * 请求体 {account, password}（account 支持邮箱或手机号）；成功 body 含 {token, refresh_token, user, permissions}。
     */
    fun loginWithPassword(account: String, password: String): PasswordLogin {
        return try {
            val body = """{"account":"$account","password":"$password"}"""
            val resp = post("/api/auth/login", body)
            if (resp == null) {
                PasswordLogin(-1, "无法连接服务器，请检查网络后重试", null)
            } else {
                val token = try {
                    JSONObject(resp.bodyRaw).optJSONObject("body")?.optString("token", "")
                } catch (e: Exception) {
                    null
                }?.takeIf { it.isNotEmpty() }
                if (resp.code == 0) accessToken = token
                PasswordLogin(resp.code, resp.message, token)
            }
        } catch (e: Exception) {
            PasswordLogin(-1, "无法连接服务器，请检查网络后重试", null)
        }
    }

    private data class Resp(val code: Int, val httpCode: Int, val message: String, val bodyRaw: String)

    private fun post(path: String, body: String): Resp? {
        var conn: HttpURLConnection? = null
        return try {
            conn = (URL(baseUrl + path).openConnection() as HttpURLConnection).apply {
                requestMethod = "POST"
                connectTimeout = TIMEOUT_MS
                readTimeout = TIMEOUT_MS
                doOutput = true
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Accept", "application/json")
            }
            conn!!.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
            val httpCode = conn!!.responseCode
            val stream = if (httpCode in 200..299) conn!!.inputStream else conn!!.errorStream
            val text = stream?.let {
                BufferedReader(InputStreamReader(it, Charsets.UTF_8)).use { reader -> reader.readText() }
            } ?: ""
            val code = try {
                JSONObject(text).optInt("code", -1)
            } catch (e: Exception) {
                -1
            }
            val message = try {
                JSONObject(text).optString("message", "")
            } catch (e: Exception) {
                ""
            }
            Resp(code, httpCode, message, text)
        } catch (e: Exception) {
            null
        } finally {
            conn?.disconnect()
        }
    }
}
