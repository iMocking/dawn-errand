package com.dawnerrand.store.data.api

import com.dawnerrand.store.BuildConfig
import com.dawnerrand.store.data.StoreSession
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL

/**
 * 商家端 API 客户端（HttpURLConnection，无第三方依赖）。
 * 对接后端统一返回 {code, message, body, request_id}（code=0 成功）。
 */
object ApiClient {

    var baseUrl: String = BuildConfig.API_BASE_URL

    /** 登录成功后的访问令牌；App 启动时从 StoreSession 恢复 */
    var accessToken: String? = null

    private const val TIMEOUT_MS = 6000

    data class LoginResult(
        val code: Int,
        val message: String,
        val token: String?,
        val name: String?,
        val role: String?,
    )

    /** 邮箱/密码登录：POST /api/auth/login，请求体 {account, password}。 */
    fun login(account: String, password: String): LoginResult {
        val body = JSONObject().apply {
            put("account", account)
            put("password", password)
        }.toString()
        val resp = request("POST", "/api/auth/login", body, null)
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
        val active = token ?: accessToken
        if (active.isNullOrBlank()) return
        request("POST", "/api/auth/logout", """{"refresh_token":""}""", active)
        accessToken = null
    }

    /** 发送短信验证码，dev 环境返回自动回显的验证码。 */
    fun sendSmsCode(phone: String): String? {
        val resp = request("POST", "/api/auth/sms/send", """{"phone":"$phone"}""", null)
        if (resp?.code != 0) return null
        return resp.body?.optString("dev_code")?.takeIf { it.isNotBlank() }
    }

    /** 手机验证码登录。 */
    fun loginBySms(phone: String, code: String): LoginResult {
        val resp = request(
            "POST", "/api/auth/sms/login",
            JSONObject().put("phone", phone).put("code", code).toString(), null,
        ) ?: return LoginResult(-1, "无法连接服务器，请检查网络后重试", null, null, null)
        if (resp.code != 0) return LoginResult(resp.code, resp.message.ifBlank { "登录失败" }, null, null, null)
        val token = resp.body?.optString("token")?.takeIf { it.isNotBlank() }
        val user = resp.body?.optJSONObject("user")
        accessToken = token
        return LoginResult(0, resp.message, token, user?.optString("name"), user?.optString("role"))
    }

    /** GET 请求，返回 body 对象；失败返回 null。 */
    fun getObject(path: String): JSONObject? =
        request("GET", path, null, accessToken ?: StoreSession.token)?.body

    /** POST 请求，返回 body 对象；失败返回 null。 */
    fun postObject(path: String, body: JSONObject): JSONObject? =
        request("POST", path, body.toString(), accessToken ?: StoreSession.token)?.body

    /** POST 请求，仅关心是否成功。 */
    fun postOk(path: String, body: JSONObject): Boolean =
        request("POST", path, body.toString(), accessToken ?: StoreSession.token)?.code == 0

    data class Resp(
        val code: Int,
        val message: String,
        val body: JSONObject?,
        val raw: String,
    )

    private fun request(method: String, path: String, body: String?, token: String?): Resp? {
        var conn: HttpURLConnection? = null
        return try {
            conn = (URL(baseUrl + path).openConnection() as HttpURLConnection).apply {
                requestMethod = method
                connectTimeout = TIMEOUT_MS
                readTimeout = TIMEOUT_MS
                doOutput = body != null
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Accept", "application/json")
                if (!token.isNullOrBlank()) {
                    setRequestProperty("Authorization", "Bearer $token")
                }
            }
            if (body != null) {
                conn.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
            }
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
                raw = text,
            )
        } catch (e: Exception) {
            null
        } finally {
            conn?.disconnect()
        }
    }
}
