package com.dawnerrand.driver.data.api

import android.util.Log
import com.dawnerrand.driver.BuildConfig
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL
import java.util.UUID

/**
 * 骑手端 API 客户端（HttpURLConnection，无第三方依赖）。
 * 统一返回 {code, message, body, request_id}（code=0 成功）。
 */
object ApiClient {

    private const val TAG = "ApiClient"

    var baseUrl: String = BuildConfig.API_BASE_URL

    /** 登录成功后的访问令牌 */
    var accessToken: String? = null

    private const val TIMEOUT_MS = 6000

    data class PasswordLogin(val code: Int, val message: String, val token: String?)

    /** 一次调用的完整结果：失败也返回，不再用 null 混淆「网络不通/业务拒绝/响应异常」。 */
    data class Resp(
        val code: Int,
        val message: String,
        val body: JSONObject?,
        val httpCode: Int,
        val requestId: String,
    ) {
        val ok: Boolean get() = code == 0 && httpCode in 200..299
    }

    /** GET，返回 body 对象；失败返回 null（原因已写日志，供联调排查）。 */
    fun getObject(path: String): JSONObject? = request("GET", path, null).takeIf { it.ok }?.body

    /** POST，返回 body 对象；失败返回 null（原因已写日志）。 */
    fun postObject(path: String, body: JSONObject = JSONObject()): JSONObject? =
        request("POST", path, body.toString()).takeIf { it.ok }?.body

    fun postOk(path: String, body: JSONObject = JSONObject()): Boolean =
        request("POST", path, body.toString()).ok

    /** 需要把后端 message 透给用户的调用走这个。 */
    fun postResp(path: String, body: JSONObject = JSONObject()): Resp =
        request("POST", path, body.toString())

    /** 上下线、营业状态等后端是 PUT 路由，用 POST 会拿到 405。 */
    fun putResp(path: String, body: JSONObject = JSONObject()): Resp =
        request("PUT", path, body.toString())

    /** 发送验证码：成功与否连同原因一起返回，避免把「服务不可用」报成「短信服务异常」。 */
    fun sendSmsCode(phone: String): Resp = request(
        "POST", "/api/auth/sms/send", """{"phone":"$phone"}""",
    )

    fun smsLogin(phone: String, code: String): Resp = request(
        "POST", "/api/auth/sms/login", """{"phone":"$phone","code":"$code"}""",
    ).also { if (it.ok) accessToken = it.body?.optString("token")?.takeIf { t -> t.isNotBlank() } }

    /** 邮箱/密码登录：POST /api/auth/login（account 支持邮箱或手机号）。 */
    fun loginWithPassword(account: String, password: String): PasswordLogin {
        val body = JSONObject().put("account", account).put("password", password).toString()
        val resp = request("POST", "/api/auth/login", body)
        val token = resp.body?.optString("token")?.takeIf { it.isNotBlank() }
        if (resp.ok) accessToken = token
        return PasswordLogin(resp.code, resp.message, token)
    }

    private fun request(method: String, path: String, body: String?): Resp {
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
            if (json == null) {
                // 响应不是统一信封（网关页、5xx HTML 等），给出可判读的文案而不是空 message。
                val msg = if (httpCode in 200..299) "服务响应格式异常" else "服务暂不可用（HTTP $httpCode）"
                Log.w(TAG, "$method $path -> $msg http=$httpCode body=${text.take(200)}")
                Resp(-1, msg, null, httpCode, requestId)
            } else {
                val code = json.optInt("code", -1)
                val serverRequestId = json.optString("request_id").takeIf { it.isNotBlank() } ?: requestId
                if (code != 0) {
                    Log.w(TAG, "$method $path -> code=$code http=$httpCode ${json.optString("message")} req=$serverRequestId")
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
