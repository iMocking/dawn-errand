package com.dawnerrand.store.ui.screens.login

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.viewModelScope
import com.dawnerrand.store.data.StoreSession
import com.dawnerrand.store.data.api.ApiClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

enum class LoginMode { PHONE, EMAIL }

/**
 * 登录页：
 * - 手机验证码：后端短信接口尚未部署，演示模式（任意 11 位手机号 + 6 位验证码）。
 * - 邮箱密码：直连后端真实接口 POST /api/auth/login。
 */
class LoginViewModel(app: Application) : AndroidViewModel(app) {

    val mode = MutableLiveData(LoginMode.PHONE)

    val phone = MutableLiveData("138 8888 8888")
    val code = MutableLiveData("")

    val email = MutableLiveData("")
    val password = MutableLiveData("")

    private val _loading = MutableLiveData(false)
    val loading: LiveData<Boolean> = _loading

    private val _message = MutableLiveData<String?>(null)
    val message: LiveData<String?> = _message

    private val _loggedIn = MutableLiveData(false)
    val loggedIn: LiveData<Boolean> = _loggedIn

    fun switchMode(value: LoginMode) {
        mode.value = value
    }

    fun onPhoneChange(v: String) = phone.postValue(v)

    fun onCodeChange(v: String) = code.postValue(v)

    fun onEmailChange(v: String) = email.postValue(v)

    fun onPasswordChange(v: String) = password.postValue(v)

    fun sendCode() {
        if ((mode.value ?: LoginMode.PHONE) != LoginMode.PHONE) {
            _message.value = "邮箱密码登录无需验证码"
            return
        }
        val p = (phone.value ?: "").replace(" ", "")
        if (!p.matches(Regex("^1[3-9]\\d{9}$"))) {
            _message.value = "请输入正确的手机号"
            return
        }
        _message.value = "验证码已发送（演示环境请输入任意 6 位数字）"
    }

    fun submit() {
        if (_loading.value == true) return
        when (mode.value ?: LoginMode.PHONE) {
            LoginMode.PHONE -> loginDemo()
            LoginMode.EMAIL -> loginReal()
        }
    }

    /** 演示登录：后端短信接口未部署，任意手机号 + 6 位验证码放行 */
    private fun loginDemo() {
        val p = (phone.value ?: "").replace(" ", "")
        val c = code.value ?: ""
        if (!p.matches(Regex("^1[3-9]\\d{9}$"))) {
            _message.value = "请输入正确的手机号"
            return
        }
        if (!c.matches(Regex("^\\d{6}$"))) {
            _message.value = "请输入 6 位数字验证码"
            return
        }
        StoreSession.save(
            getApplication(), "demo-token-$p", "陈老板", p, "", p, real = false
        )
        _loggedIn.value = true
    }

    /** 邮箱密码登录：真实后端接口 */
    private fun loginReal() {
        val account = (email.value ?: "").trim()
        val pwd = password.value ?: ""
        if (account.isEmpty()) {
            _message.value = "请输入商家账号"
            return
        }
        if (pwd.length < 6) {
            _message.value = "密码至少 6 位"
            return
        }
        _loading.value = true
        viewModelScope.launch {
            val result = withContext(Dispatchers.IO) { ApiClient.login(account, pwd) }
            _loading.value = false
            val token = result.token
            if (result.code == 0 && !token.isNullOrBlank()) {
                StoreSession.save(
                    getApplication(),
                    token,
                    result.name ?: "商家店主",
                    account,
                    result.role ?: "",
                    phone = if (account.matches(Regex("^1\\d{10}$"))) account else "138 8888 8888",
                    real = true,
                )
                _loggedIn.value = true
            } else {
                _message.value = result.message.ifBlank { "登录失败，请稍后重试" }
            }
        }
    }

    fun clearMessage() {
        _message.value = null
    }
}
