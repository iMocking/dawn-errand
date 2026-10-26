package com.dawnerrand.store.data.api

import android.util.Log
import com.dawnerrand.store.BuildConfig
import com.dawnerrand.store.data.StoreSession
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL
import java.util.UUID

/**
 * 商家端 API 客户端（HttpURLConnection，无第三方依赖）。
 * 对接后端统一返回 {code, message, body, request_id}（code=0 成功）。
 */
object ApiClient {

    private const val TAG = "ApiClient"

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

    /** 一次调用的完整结果：失败也返回，不再用 null 混淆「网络不通 / 业务拒绝 / 响应异常」。 */
    data class Resp(
        val code: Int,
        val message: String,
        val body: JSONObject?,
        val httpCode: Int,
        val requestId: String,
    ) {
        val ok: Boolean get() = code == 0 && httpCode in 200..299
    }

    /** 邮箱/密码登录：POST /api/auth/login，请求体 {account, password}。 */
    fun login(account: String, password: String): LoginResult =
        loginResult(request("POST", "/api/auth/login", accountPasswordBody(account, password), null))

    /** 登出：撤销服务端会话（失败不阻断本地登出） */
    fun logout(token: String?) {
        val active = token ?: accessToken
        if (active.isNullOrBlank()) return
        request("POST", "/api/auth/logout", """{"refresh_token":""}""", active)
        accessToken = null
    }

    /** 发送短信验证码：连同 code/message 一起返回，避免把「服务不可用」显示成「已发送」。 */
    fun sendSmsCode(phone: String): Resp =
        request("POST", "/api/auth/sms/send", """{"phone":"$phone"}""", null)

    /** 手机验证码登录。 */
    fun loginBySms(phone: String, code: String): LoginResult = loginResult(
        request(
            "POST", "/api/auth/sms/login",
            JSONObject().put("phone", phone).put("code", code).toString(), null,
        ),
    )

    /** GET 请求，返回 body 对象；失败返回 null（原因已落日志，供联调排查）。 */
    fun getObject(path: String): JSONObject? =
        request("GET", path, null, accessToken ?: StoreSession.token).takeIf { it.ok }?.body

    /** POST 请求，返回 body 对象；失败返回 null（原因已落日志）。 */
    fun postObject(path: String, body: JSONObject): JSONObject? = postResp(path, body).takeIf { it.ok }?.body

    /** POST 请求，仅关心是否成功。 */
    fun postOk(path: String, body: JSONObject): Boolean = postResp(path, body).ok

    /** 需要把后端 code/message 透给用户（弹窗、toast）的调用走这两个。 */
    fun postResp(path: String, body: JSONObject): Resp =
        request("POST", path, body.toString(), accessToken ?: StoreSession.token)

    /** 营业状态、菜品编辑等后端是 PUT 路由，用 POST 会拿到 405。 */
    fun putResp(path: String, body: JSONObject): Resp =
        request("PUT", path, body.toString(), accessToken ?: StoreSession.token)

    private fun accountPasswordBody(account: String, password: String): String =
        JSONObject().put("account", account).put("password", password).toString()

    private fun loginResult(resp: Resp): LoginResult {
        if (!resp.ok) {
            return LoginResult(resp.code, resp.message.ifBlank { "登录失败" }, null, null, null)
        }
        val token = resp.body?.optString("token")?.takeIf { it.isNotBlank() }
        val user = resp.body?.optJSONObject("user")
        accessToken = token
        return LoginResult(
            code = 0,
            message = resp.message,
            token = token,
            name = user?.optString("name")?.takeIf { it.isNotBlank() },
            role = user?.optString("role")?.takeIf { it.isNotBlank() },
        )
    }

    private fun request(method: String, path: String, body: String?, token: String?): Resp {
        var conn: HttpURLConnection? = null
        val requestId = UUID.randomUUID().toString().replace("-", "").uppercase()
        return try {
            conn = (URL(baseUrl + path).openConnection() as HttpURLConnection).apply {
                requestMethod = method
                connectTimeout = TIMEOUT_MS
                readTimeout = TIMEOUT_MS
                doOutput = body != null
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Accept", "application/json")
                setRequestProperty("X-Request-ID", requestId)
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
            if (json == null) {
                // 响应不是统一信封（网关页、5xx HTML、405 纯文本等），给可判读文案而不是空 message。
                val msg = if (httpCode in 200..299) "服务响应格式异常" else "服务暂不可用（HTTP $httpCode）"
                Log.w(TAG, "$method $path -> $msg http=$httpCode body=${text.take(200)}")
                Resp(-1, msg, null, httpCode, requestId)
            } else {
                val code = json.optInt("code", -1)
                val serverRequestId = json.optString("request_id").takeIf { it.isNotBlank() } ?: requestId
                if (code != 0) {
                    Log.w(
                        TAG,
                        "$method $path -> code=$code http=$httpCode ${json.optString("message")} req=$serverRequestId",
                    )
                }
                Resp(
                    code = code,
                    message = json.optString("message"),
                    body = json.optJSONObject("body"),
                    httpCode = httpCode,
                    requestId = serverRequestId,
                )
            }
        } catch (e: Exception) {
            Log.w(TAG, "$method $path -> 请求失败 ${e.javaClass.simpleName}: ${e.message} req=$requestId")
            Resp(-1, "无法连接服务器，请检查网络后重试", null, 0, requestId)
        } finally {
            conn?.disconnect()
        }
    }
}
