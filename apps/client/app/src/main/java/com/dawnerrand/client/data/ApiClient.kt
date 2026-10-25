package com.dawnerrand.client.data

import com.dawnerrand.client.BuildConfig
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL
import java.util.UUID

/** 后端返回非 0 业务码（或 HTTP 非 2xx）时抛出 */
class ApiException(val code: Int, override val message: String) : Exception(message)

data class SmsLoginResult(
    val token: String,
    val refreshToken: String,
    val userId: String,
    val name: String,
    val role: String,
)

/**
 * 客户端轻量 API 封装：仅依赖 HttpURLConnection + org.json，不引入额外网络库。
 * 契约与后端统一返回一致：{message, body, request_id, code}，code == 0 表示成功。
 */
object ApiClient {

    private val baseUrl: String = BuildConfig.API_BASE_URL.trimEnd('/')

    /** 邮箱/密码登录：POST /api/auth/login（account 支持邮箱或手机号） */
    suspend fun loginByPassword(account: String, password: String): SmsLoginResult =
        withContext(Dispatchers.IO) {
            val body = JSONObject().put("account", account).put("password", password)
            val env = post("/api/auth/login", body)
            val payload = env.body ?: throw ApiException(-1, "登录响应为空")
            val user = payload.optJSONObject("user")
            val token = payload.optString("token")
            if (token.isBlank()) throw ApiException(-1, "登录响应缺少 token")
            SmsLoginResult(
                token = token,
                refreshToken = payload.optString("refresh_token"),
                userId = user?.optString("id").orEmpty(),
                name = user?.optString("name").orEmpty(),
                role = user?.optString("role").orEmpty(),
            )
        }

    /** GET 请求（携带当前会话 token），返回 body 对象；失败返回 null */
    suspend fun getObject(path: String): JSONObject? = withContext(Dispatchers.IO) {
        val token = DemoRepository.session()?.token
        if (token.isNullOrBlank()) return@withContext null
        val url = URL(baseUrl + path)
        val conn = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "GET"
            connectTimeout = 8_000
            readTimeout = 10_000
            useCaches = false
            setRequestProperty("Accept", "application/json")
            setRequestProperty("Authorization", "Bearer $token")
            setRequestProperty(
                "X-Request-ID",
                UUID.randomUUID().toString().replace("-", "").uppercase(),
            )
        }
        try {
            val httpCode = conn.responseCode
            val stream = if (httpCode in 200..299) conn.inputStream else conn.errorStream
            val text = stream?.bufferedReader(Charsets.UTF_8)?.use { it.readText() }.orEmpty()
            val json = runCatching { JSONObject(text) }.getOrNull() ?: return@withContext null
            if (json.optInt("code", -1) == 0) json.optJSONObject("body") else null
        } catch (e: Exception) {
            null
        } finally {
            conn.disconnect()
        }
    }

    /** 发送短信验证码；dev 环境返回自动回显的验证码，正式环境返回 null */
    suspend fun sendSmsCode(phone: String): String? = withContext(Dispatchers.IO) {
        val body = JSONObject().put("phone", phone)
        val env = post("/api/auth/sms/send", body)
        env.body?.optString("dev_code")?.takeIf { it.isNotBlank() }
    }

    /** 手机验证码登录（后端未注册手机号会自动创建会员账号） */
    suspend fun loginBySms(phone: String, code: String): SmsLoginResult =
        withContext(Dispatchers.IO) {
            val body = JSONObject().put("phone", phone).put("code", code)
            val env = post("/api/auth/sms/login", body)
            val payload = env.body ?: throw ApiException(-1, "登录响应为空")
            val user = payload.optJSONObject("user")
            val token = payload.optString("token")
            if (token.isBlank()) throw ApiException(-1, "登录响应缺少 token")
            SmsLoginResult(
                token = token,
                refreshToken = payload.optString("refresh_token"),
                userId = user?.optString("id").orEmpty(),
                name = user?.optString("name").orEmpty(),
                role = user?.optString("role").orEmpty(),
            )
        }

    private fun post(path: String, body: JSONObject): ApiEnvelope {
        val url = URL(baseUrl + path)
        val conn = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "POST"
            connectTimeout = 8_000
            readTimeout = 10_000
            doOutput = true
            useCaches = false
            setRequestProperty("Content-Type", "application/json; charset=utf-8")
            setRequestProperty("Accept", "application/json")
            setRequestProperty(
                "X-Request-ID",
                UUID.randomUUID().toString().replace("-", "").uppercase(),
            )
        }
        try {
            conn.outputStream.use { it.write(body.toString().toByteArray(Charsets.UTF_8)) }
            val httpCode = conn.responseCode
            val stream = if (httpCode in 200..299) conn.inputStream else conn.errorStream
            val text = stream?.bufferedReader(Charsets.UTF_8)?.use { it.readText() }.orEmpty()
            val json = runCatching { JSONObject(text) }.getOrNull()
            val code = json?.optInt("code", -1) ?: -1
            val message = json?.optString("message").orEmpty().ifBlank {
                if (httpCode in 200..299) "请求失败" else "服务暂不可用（HTTP $httpCode）"
            }
            if (httpCode !in 200..299 || code != 0) {
                throw ApiException(if (code > 0) code else -1, message)
            }
            return ApiEnvelope(
                code = code,
                message = message,
                body = json?.optJSONObject("body"),
                requestId = json?.optString("request_id").orEmpty(),
            )
        } catch (e: ApiException) {
            throw e
        } catch (e: IOException) {
            throw e
        } catch (e: Exception) {
            throw IOException(e.message ?: "网络请求失败", e)
        } finally {
            conn.disconnect()
        }
    }

    private data class ApiEnvelope(
        val code: Int,
        val message: String,
        val body: JSONObject?,
        val requestId: String,
    )
}
