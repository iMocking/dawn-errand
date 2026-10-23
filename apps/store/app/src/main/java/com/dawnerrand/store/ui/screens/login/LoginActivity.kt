package com.dawnerrand.store.ui.screens.login

import android.content.Intent
import android.os.Bundle
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.data.StoreRepository
import com.dawnerrand.store.databinding.ActivityLoginBinding
import com.dawnerrand.store.ui.screens.main.MainActivity
import com.dawnerrand.store.ui.util.TextWatcher
import com.dawnerrand.store.ui.util.padTopWithStatusBar
import com.dawnerrand.store.ui.util.toast

/**
 * 商家登录：手机号验证码 + 协议勾选（login 稿）。
 * 已登录直接进入主页面。
 */
class LoginActivity : AppCompatActivity() {

    private lateinit var b: ActivityLoginBinding
    private val vm: LoginViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        if (StoreRepository.isLoggedIn(this)) {
            goMain()
            return
        }

        b = ActivityLoginBinding.inflate(layoutInflater)
        setContentView(b.root)

        b.headerBox.padTopWithStatusBar()

        // 协议勾选
        var agreed = true
        b.agreementCheck.setOnClickListener {
            agreed = !agreed
            b.agreementCheck.setBackgroundResource(
                if (agreed) com.dawnerrand.store.R.drawable.bg_check_on
                else com.dawnerrand.store.R.drawable.bg_switch_big_off
            )
            b.agreementCheckIcon.alpha = if (agreed) 1f else 0f
        }

        b.getCode.setOnClickListener {
            toast("验证码已发送，演示环境请输入任意6位数字")
        }

        b.submitBtn.setOnClickListener {
            if (!agreed) {
                toast("请先阅读并同意《商家入驻协议》")
                return@setOnClickListener
            }
            vm.login(this)
        }

        b.footerTip.setOnClickListener { vm.login(this) }

        vm.phone.observe(this) { b.phoneInput.setText(it) }
        vm.code.observe(this) { b.codeInput.setText(it) }
        b.phoneInput.addTextChangedListener(TextWatcher { vm.onPhoneChange(it) })
        b.codeInput.addTextChangedListener(TextWatcher { vm.onCodeChange(it) })

        vm.loggedIn.observe(this) { ok ->
            if (ok) {
                toast("欢迎回来，${b.phoneInput.text}")
                goMain()
            }
        }
    }

    private fun goMain() {
        startActivity(Intent(this, MainActivity::class.java))
        finish()
    }
}
