package com.dawnerrand.driver.data.api

import com.dawnerrand.driver.BuildConfig
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL

/**
 * 骑手端 API 客户端（HttpURLConnection，无第三方依赖）。
 * 统一返回 {code, message, body, request_id}（code=0 成功）。
 */
object ApiClient {

    var baseUrl: String = BuildConfig.API_BASE_URL

    /** 登录成功后的访问令牌 */
    var accessToken: String? = null

    private const val TIMEOUT_MS = 6000

    enum class SmsStatus { OK, DEMO_FALLBACK }

    data class PasswordLogin(val code: Int, val message: String, val token: String?)

    /** GET，返回 body 对象；失败返回 null。 */
    fun getObject(path: String): JSONObject? = request("GET", path, null)?.body

    /** POST，返回 body 对象；失败返回 null。 */
    fun postObject(path: String, body: JSONObject = JSONObject()): JSONObject? =
        request("POST", path, body.toString())?.body

    fun postOk(path: String, body: JSONObject = JSONObject()): Boolean =
        request("POST", path, body.toString())?.code == 0

    fun sendSmsCode(phone: String): SmsStatus {
        val resp = request("POST", "/api/auth/sms/send", """{"phone":"$phone"}""")
        return if (resp?.code == 0) SmsStatus.OK else SmsStatus.DEMO_FALLBACK
    }

    fun smsLogin(phone: String, code: String): Boolean =
        request("POST", "/api/auth/sms/login", """{"phone":"$phone","code":"$code"}""")?.code == 0

    /** 邮箱/密码登录：POST /api/auth/login（account 支持邮箱或手机号）。 */
    fun loginWithPassword(account: String, password: String): PasswordLogin {
        val body = JSONObject().put("account", account).put("password", password).toString()
        val resp = request("POST", "/api/auth/login", body)
            ?: return PasswordLogin(-1, "无法连接服务器，请检查网络后重试", null)
        val token = resp.body?.optString("token")?.takeIf { it.isNotBlank() }
        if (resp.code == 0) accessToken = token
        return PasswordLogin(resp.code, resp.message, token)
    }

    data class Resp(val code: Int, val message: String, val body: JSONObject?)

    private fun request(method: String, path: String, body: String?): Resp? {
        var conn: HttpURLConnection? = null
        return try {
            conn = (URL(baseUrl + path).openConnection() as HttpURLConnection).apply {
                requestMethod = method
                connectTimeout = TIMEOUT_MS
                readTimeout = TIMEOUT_MS
                doOutput = body != null
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Accept", "application/json")
                accessToken?.takeIf { it.isNotBlank() }?.let {
                    setRequestProperty("Authorization", "Bearer $it")
                }
            }
            if (body != null) conn.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
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
