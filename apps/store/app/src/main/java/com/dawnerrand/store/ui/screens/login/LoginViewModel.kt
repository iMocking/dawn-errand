package com.dawnerrand.store.ui.screens.login

import android.content.Context
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.StoreRepository

/**
 * 登录页：手机号 + 6 位验证码（演示态：任意 11 位手机号与 6 位验证码即可登录）。
 */
class LoginViewModel : ViewModel() {

    val phone = MutableLiveData("138 8888 8888")
    val code = MutableLiveData("")

    private val _loggedIn = MutableLiveData(false)
    val loggedIn: LiveData<Boolean> = _loggedIn

    fun onPhoneChange(v: String) = phone.postValue(v)

    fun onCodeChange(v: String) = code.postValue(v)

    fun login(ctx: Context) {
        val p = (phone.value ?: "").replace(" ", "")
        val c = code.value ?: ""
        if (p.length != 11 || c.length != 6) return
        StoreRepository.saveLogin(ctx, "demo-token", p)
        _loggedIn.value = true
    }
}
