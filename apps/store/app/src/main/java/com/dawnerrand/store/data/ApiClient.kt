package com.dawnerrand.store.data.api

import com.dawnerrand.store.BuildConfig
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL

/**
 * 轻量 API 客户端（HttpURLConnection，无第三方依赖）。
 * 对接后端统一返回 {code, message, body, request_id}（code=0 成功）。
 * 邮箱/密码登录直连真实接口 /api/auth/login；手机验证码接口后端尚未部署，走演示模式。
 */
object ApiClient {

    var baseUrl: String = BuildConfig.API_BASE_URL

    /** 最近一次真实登录成功后的访问令牌 */
    var accessToken: String? = null

    private const val TIMEOUT_MS = 5000

    data class LoginResult(
        val code: Int,
        val message: String,
        val token: String?,
        val name: String?,
        val role: String?,
    )

    /**
     * 邮箱/密码登录：POST /api/auth/login，请求体 {account, password}。
     * account 支持邮箱或手机号；成功 body 含 {token, refresh_token, user, permissions}。
     */
    fun login(account: String, password: String): LoginResult {
        val body = JSONObject().apply {
            put("account", account)
            put("password", password)
        }.toString()
        val resp = post("/api/auth/login", body)
            ?: return LoginResult(-1, "无法连接服务器，请检查网络后重试", null, null, null)
        if (resp.code != 0) {
            return LoginResult(resp.code, resp.message.ifBlank { "登录失败" }, null, null, null)
        }
        val token = resp.body?.optString("token")?.takeIf { it.isNotBlank() }
        val user = resp.body?.optJSONObject("user")
        val name = user?.optString("name")?.takeIf { it.isNotBlank() }
        val role = user?.optString("role")?.takeIf { it.isNotBlank() }
        accessToken = token
        return LoginResult(0, resp.message, token, name, role)
    }

    /** 登出：撤销服务端会话（失败不阻断本地登出） */
    fun logout(token: String?) {
        if (token.isNullOrBlank()) return
        val body = JSONObject().apply { put("refresh_token", "") }.toString()
        post("/api/auth/logout", body, token)
        accessToken = null
    }

    private data class Resp(
        val code: Int,
        val message: String,
        val body: JSONObject?,
    )

    private fun post(path: String, body: String, token: String? = null): Resp? {
        var conn: HttpURLConnection? = null
        return try {
            conn = (URL(baseUrl + path).openConnection() as HttpURLConnection).apply {
                requestMethod = "POST"
                connectTimeout = TIMEOUT_MS
                readTimeout = TIMEOUT_MS
                doOutput = true
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Accept", "application/json")
                if (!token.isNullOrBlank()) {
                    setRequestProperty("Authorization", "Bearer $token")
                }
            }
            conn.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
            val httpCode = conn.responseCode
            val stream = if (httpCode in 200..299) conn.inputStream else conn.errorStream
            val text = stream?.let {
                BufferedReader(InputStreamReader(it, Charsets.UTF_8)).use { r -> r.readText() }
            } ?: ""
            val json = try {
                JSONObject(text)
            } catch (e: Exception) {
                null
            }
            Resp(
                code = json?.optInt("code", -1) ?: -1,
                message = json?.optString("message").orEmpty(),
                body = json?.optJSONObject("body"),
            )
        } catch (e: Exception) {
            null
        } finally {
            conn?.disconnect()
        }
    }
}
