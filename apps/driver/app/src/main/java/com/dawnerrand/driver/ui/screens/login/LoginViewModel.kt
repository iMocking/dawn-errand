package com.dawnerrand.driver.ui.screens.login

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dawnerrand.driver.data.api.ApiClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

sealed class LoginUiState {
    object Idle : LoginUiState()
    data class Countdown(val seconds: Int) : LoginUiState()
    object Sending : LoginUiState()
    data class Success(val demoMode: Boolean) : LoginUiState()
    data class Error(val message: String) : LoginUiState()
}

class LoginViewModel : ViewModel() {

    private val _state = MutableLiveData<LoginUiState>(LoginUiState.Idle)
    val state: LiveData<LoginUiState> = _state

    var agreementChecked: Boolean = true
        private set

    private var countdownJob: Job? = null

    val isDemoMode: Boolean
        get() = ApiClient.demoMode

    fun toggleAgreement() {
        agreementChecked = !agreementChecked
    }

    fun sendCode(phoneRaw: String) {
        val phone = phoneRaw.replace(" ", "")
        if (phone.length != 11) {
            _state.value = LoginUiState.Error("请输入正确的11位手机号")
            return
        }
        startCountdown()
        viewModelScope.launch {
            val status = withContext(Dispatchers.IO) { ApiClient.sendSmsCode(phone) }
            when (status) {
                ApiClient.SmsStatus.OK -> _state.value = LoginUiState.Error("验证码已发送")
                ApiClient.SmsStatus.DEMO_FALLBACK ->
                    _state.value = LoginUiState.Error("后端未提供验证码接口，已进入演示模式：任意6位验证码可登录")
            }
        }
    }

    fun login(phoneRaw: String, code: String) {
        val phone = phoneRaw.replace(" ", "")
        if (phone.length != 11) {
            _state.value = LoginUiState.Error("请输入正确的11位手机号")
            return
        }
        if (code.length != 6) {
            _state.value = LoginUiState.Error("请输入6位验证码")
            return
        }
        if (!agreementChecked) {
            _state.value = LoginUiState.Error("请先阅读并同意《骑手入驻协议》")
            return
        }
        if (ApiClient.demoMode) {
            _state.value = LoginUiState.Success(true)
            return
        }
        _state.value = LoginUiState.Sending
        viewModelScope.launch {
            val ok = withContext(Dispatchers.IO) { ApiClient.smsLogin(phone, code) }
            if (ok) {
                _state.value = LoginUiState.Success(false)
            } else {
                _state.value = LoginUiState.Error("验证码错误，请重试")
            }
        }
    }

    /**
     * 邮箱/密码登录：直连真实接口 POST /api/auth/login。
     * 成功进入主界面；失败透传后端 message（如"账号或密码错误"）。
     */
    fun loginByEmail(email: String, password: String) {
        val account = email.trim()
        if (account.isEmpty() || !account.contains("@")) {
            _state.value = LoginUiState.Error("请输入正确的邮箱地址")
            return
        }
        if (password.isEmpty()) {
            _state.value = LoginUiState.Error("请输入密码")
            return
        }
        if (!agreementChecked) {
            _state.value = LoginUiState.Error("请先阅读并同意《骑手入驻协议》")
            return
        }
        _state.value = LoginUiState.Sending
        viewModelScope.launch {
            val result = withContext(Dispatchers.IO) { ApiClient.loginWithPassword(account, password) }
            if (result.code == 0) {
                _state.value = LoginUiState.Success(false)
            } else {
                _state.value = LoginUiState.Error(result.message.ifBlank { "登录失败，请重试" })
            }
        }
    }

    private fun startCountdown() {
        countdownJob?.cancel()
        countdownJob = viewModelScope.launch {
            for (i in 60 downTo 1) {
                _state.value = LoginUiState.Countdown(i)
                delay(1000)
            }
            _state.value = LoginUiState.Idle
        }
    }

    override fun onCleared() {
        countdownJob?.cancel()
        super.onCleared()
    }
}
