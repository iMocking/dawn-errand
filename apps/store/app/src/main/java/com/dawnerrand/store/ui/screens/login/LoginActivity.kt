package com.dawnerrand.store.ui.screens.login

import android.content.Intent
import android.graphics.Typeface
import android.os.Bundle
import android.text.InputType
import android.view.View
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.R
import com.dawnerrand.store.data.StoreSession
import com.dawnerrand.store.databinding.ActivityLoginBinding
import com.dawnerrand.store.ui.screens.main.MainActivity
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.TextWatcher
import com.dawnerrand.store.ui.util.padTopWithStatusBar
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast

/**
 * 商家登录：默认邮箱密码（真实后端），手机号验证码作为备用方式；协议勾选（login 稿）。
 * 已登录直接进入主页面。
 */
class LoginActivity : AppCompatActivity() {

    private lateinit var b: ActivityLoginBinding
    private val vm: LoginViewModel by viewModels()

    private var agreed = true
    private var passwordVisible = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        if (StoreSession.isLoggedIn()) {
            goMain()
            return
        }

        b = ActivityLoginBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(false)

        b.headerBox.padTopWithStatusBar()

        bindTabs()
        bindAgreement()
        bindInputs()

        b.getCode.setOnClickListener { vm.sendCode() }

        b.footerTip.setOnClickListener {
            toast("入驻申请请联系平台运营 400-000-0000")
        }

        b.submitBtn.setOnClickListener {
            if (!agreed) {
                toast("请先阅读并同意《商家入驻协议》")
                return@setOnClickListener
            }
            vm.submit()
        }

        vm.loggedIn.observe(this) { ok ->
            if (ok) {
                toast("欢迎回来，${StoreSession.name}")
                goMain()
            }
        }
        vm.message.observe(this) { msg ->
            if (!msg.isNullOrBlank()) {
                toast(msg)
                vm.clearMessage()
            }
        }
        vm.loading.observe(this) { loading ->
            b.submitBtn.isEnabled = !loading
            b.submitBtn.alpha = if (loading) 0.7f else 1f
            b.submitBtn.text = if (loading) "登录中…" else "登录 / 入驻"
        }

        applyMode(LoginMode.EMAIL)
    }

    private fun bindTabs() {
        b.tabPhone.setOnClickListener { applyMode(LoginMode.PHONE) }
        b.tabEmail.setOnClickListener { applyMode(LoginMode.EMAIL) }
    }

    private fun applyMode(mode: LoginMode) {
        vm.switchMode(mode)
        val phone = mode == LoginMode.PHONE
        b.phoneGroup.visibility = if (phone) View.VISIBLE else View.GONE
        b.emailGroup.visibility = if (phone) View.GONE else View.VISIBLE

        b.tabPhoneText.setTextColor(if (phone) Design.PRIMARY else Design.TEXT_SECONDARY)
        b.tabEmailText.setTextColor(if (phone) Design.TEXT_SECONDARY else Design.PRIMARY)
        b.tabPhoneText.setTypeface(b.tabPhoneText.typeface, if (phone) Typeface.BOLD else Typeface.NORMAL)
        b.tabEmailText.setTypeface(b.tabEmailText.typeface, if (phone) Typeface.NORMAL else Typeface.BOLD)
        b.tabPhoneLine.visibility = if (phone) View.VISIBLE else View.INVISIBLE
        b.tabEmailLine.visibility = if (phone) View.INVISIBLE else View.VISIBLE

        b.submitBtn.text = "登录 / 入驻"
    }

    private fun bindAgreement() {
        b.agreementCheck.setOnClickListener {
            agreed = !agreed
            b.agreementCheck.setBackgroundResource(
                if (agreed) R.drawable.bg_check_on else R.drawable.bg_switch_big_off
            )
            b.agreementCheckIcon.alpha = if (agreed) 1f else 0f
        }
    }

    private fun bindInputs() {
        b.phoneInput.setText(vm.phone.value)
        b.phoneInput.addTextChangedListener(TextWatcher { vm.onPhoneChange(it) })
        b.codeInput.addTextChangedListener(TextWatcher { vm.onCodeChange(it) })
        b.emailInput.setText(vm.email.value)
        b.emailInput.addTextChangedListener(TextWatcher { vm.onEmailChange(it) })
        b.passwordInput.setText(vm.password.value)
        b.passwordInput.addTextChangedListener(TextWatcher { vm.onPasswordChange(it) })

        b.passwordToggle.setOnClickListener {
            passwordVisible = !passwordVisible
            b.passwordInput.inputType = if (passwordVisible) {
                InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_VARIATION_VISIBLE_PASSWORD
            } else {
                InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_VARIATION_PASSWORD
            }
            val end = b.passwordInput.text.length
            b.passwordInput.setSelection(end)
            b.passwordToggle.text = if (passwordVisible) "隐藏" else "显示"
        }
    }

    private fun goMain() {
        startActivity(Intent(this, MainActivity::class.java))
        finish()
    }
}
