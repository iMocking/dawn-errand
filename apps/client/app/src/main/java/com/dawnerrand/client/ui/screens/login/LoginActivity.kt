package com.dawnerrand.client.ui.screens.login

import android.content.Intent
import android.os.Bundle
import android.os.CountDownTimer
import android.view.View
import android.widget.EditText
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.dawnerrand.client.BuildConfig
import com.dawnerrand.client.R
import com.dawnerrand.client.data.ApiClient
import com.dawnerrand.client.data.ApiException
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.legal.LegalActivity
import com.dawnerrand.client.ui.screens.main.MainActivity
import com.dawnerrand.client.ui.util.toast
import kotlinx.coroutines.launch
import java.io.IOException

/**
 * 登录 / 注册：
 * - 优先调用后端 `POST /api/auth/sms/send`、`POST /api/auth/sms/login`（dev 环境自动回填验证码）；
 * - 后端未部署短信接口（1401/404）或断网时降级为演示模式：任意 6 位验证码放行；
 * - 登录成功持久化 token / refresh_token / 手机号 / 昵称。
 */
class LoginActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private lateinit var etPhone: EditText
    private lateinit var etCode: EditText
    private lateinit var etEmail: EditText
    private lateinit var etPassword: EditText
    private lateinit var emailGroup: View
    private lateinit var phoneGroup: View
    private lateinit var tabEmail: TextView
    private lateinit var tabPhone: TextView
    private lateinit var btnCode: TextView
    private lateinit var btnLogin: TextView
    private lateinit var chkAgree: TextView
    private var emailMode = true
    private var passwordVisible = false
    private var agreed = true
    private var countDown: CountDownTimer? = null
    private var sending = false
    private var loggingIn = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_login)

        etPhone = findViewById(R.id.etPhone)
        etCode = findViewById(R.id.etCode)
        etEmail = findViewById(R.id.etEmail)
        etPassword = findViewById(R.id.etPassword)
        emailGroup = findViewById(R.id.emailGroup)
        phoneGroup = findViewById(R.id.phoneGroup)
        tabEmail = findViewById(R.id.tabEmail)
        tabPhone = findViewById(R.id.tabPhone)
        btnCode = findViewById(R.id.btnCode)
        btnLogin = findViewById(R.id.btnLogin)
        chkAgree = findViewById(R.id.chkAgree)

        if (BuildConfig.DEBUG) {
            etEmail.setText("client.test@dawnerrand.local")
            etPassword.setText("Client123456")
            etPhone.setText("13888888888")
        }

        tabEmail.setOnClickListener { switchMode(true) }
        tabPhone.setOnClickListener { switchMode(false) }
        findViewById<View>(R.id.btnTogglePassword).setOnClickListener {
            passwordVisible = !passwordVisible
            etPassword.inputType = if (passwordVisible) {
                android.text.InputType.TYPE_CLASS_TEXT or android.text.InputType.TYPE_TEXT_VARIATION_VISIBLE_PASSWORD
            } else {
                android.text.InputType.TYPE_CLASS_TEXT or android.text.InputType.TYPE_TEXT_VARIATION_PASSWORD
            }
            val end = etPassword.text.length
            etPassword.setSelection(end)
            findViewById<TextView>(R.id.btnTogglePassword).text = if (passwordVisible) "隐藏" else "显示"
        }
        btnCode.setOnClickListener { sendCode() }
        chkAgree.setOnClickListener { agreed = !agreed; renderAgree() }
        btnLogin.setOnClickListener { login() }
        findViewById<View>(R.id.btnWechat).setOnClickListener {
            ctx.toast("微信登录建设中，请使用邮箱登录")
        }
        findViewById<View>(R.id.btnAlipay).setOnClickListener {
            ctx.toast("支付宝登录建设中，请使用邮箱登录")
        }
        findViewById<View>(R.id.tvUserAgreement).setOnClickListener {
            startActivity(Intent(this, LegalActivity::class.java)
                .putExtra(LegalActivity.EXTRA_TYPE, LegalActivity.TYPE_AGREEMENT))
        }
        findViewById<View>(R.id.tvPrivacy).setOnClickListener {
            startActivity(Intent(this, LegalActivity::class.java)
                .putExtra(LegalActivity.EXTRA_TYPE, LegalActivity.TYPE_PRIVACY))
        }
        switchMode(true)
    }

    /** 邮箱登录优先；手机号验证码作为备用方式。 */
    private fun switchMode(email: Boolean) {
        emailMode = email
        emailGroup.visibility = if (email) View.VISIBLE else View.GONE
        phoneGroup.visibility = if (email) View.GONE else View.VISIBLE
        tabEmail.setTextColor(if (email) getColor(R.color.primary) else getColor(R.color.text_secondary))
        tabPhone.setTextColor(if (email) getColor(R.color.text_secondary) else getColor(R.color.primary))
        tabEmail.setTypeface(null, if (email) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL)
        tabPhone.setTypeface(null, if (!email) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL)
        btnLogin.text = "登录 / 注册"
    }

    private fun renderAgree() {
        chkAgree.setBackgroundResource(if (agreed) R.drawable.bg_agreed else R.drawable.bg_agreed_off)
        chkAgree.text = if (agreed) "✓" else ""
    }

    private fun sendCode() {
        if (sending) return
        val phone = etPhone.text.toString().trim()
        if (phone.length != 11 || !phone.startsWith("1")) {
            ctx.toast("请输入正确的11位手机号")
            return
        }
        sending = true
        btnCode.isEnabled = false
        btnCode.text = "发送中…"
        lifecycleScope.launch {
            try {
                val devCode = ApiClient.sendSmsCode(phone)
                if (!devCode.isNullOrBlank()) {
                    etCode.setText(devCode)
                    ctx.toast("验证码已发送（开发环境已自动填充）")
                } else {
                    ctx.toast("验证码已发送，请查收短信")
                }
                startCountdown()
            } catch (e: ApiException) {
                btnCode.isEnabled = true
                btnCode.text = "重新获取"
                ctx.toast(e.message)
            } catch (e: IOException) {
                btnCode.isEnabled = true
                btnCode.text = "重新获取"
                ctx.toast("网络不可用，请检查后端服务后重试")
            } finally {
                sending = false
            }
        }
    }

    private fun startCountdown() {
        countDown?.cancel()
        btnCode.isEnabled = false
        btnCode.text = "60s"
        countDown = object : CountDownTimer(60_000, 1000) {
            override fun onTick(ms: Long) {
                btnCode.text = "${ms / 1000}s"
            }

            override fun onFinish() {
                btnCode.isEnabled = true
                btnCode.text = "重新获取"
            }
        }.start()
    }

    private fun login() {
        if (loggingIn) return
        if (!emailMode) {
            loginByPhone()
            return
        }
        val account = etEmail.text.toString().trim()
        val password = etPassword.text.toString()
        if (!account.contains("@")) {
            ctx.toast("请输入正确的邮箱地址")
            return
        }
        if (password.length < 6) {
            ctx.toast("请输入至少 6 位密码")
            return
        }
        if (!agreed) {
            ctx.toast("请先阅读并同意《用户协议》和《隐私政策》")
            return
        }
        loggingIn = true
        btnLogin.isEnabled = false
        btnLogin.text = "登录中…"
        lifecycleScope.launch {
            var entered = false
            try {
                val result = ApiClient.loginByPassword(account, password)
                DemoRepository.login(
                    token = result.token,
                    refreshToken = result.refreshToken,
                    phone = account,
                    name = result.name,
                    userId = result.userId,
                    demo = false,
                )
                com.dawnerrand.client.data.ClientRemoteRepository.refresh()
                ctx.toast("登录成功，欢迎 ${DemoRepository.currentUser.name}")
                enterMain()
                entered = true
            } catch (e: ApiException) {
                ctx.toast(e.message)
            } catch (e: IOException) {
                ctx.toast("网络不可用，请检查后端服务后重试")
            } finally {
                loggingIn = false
                if (!entered) {
                    btnLogin.isEnabled = true
                    btnLogin.text = "登录 / 注册"
                }
            }
        }
    }

    private fun loginByPhone() {
        val phone = etPhone.text.toString().trim()
        val code = etCode.text.toString().trim()
        if (phone.length != 11 || !phone.startsWith("1")) {
            ctx.toast("请输入正确的11位手机号")
            return
        }
        if (code.length != 6) {
            ctx.toast("请输入6位验证码")
            return
        }
        if (!agreed) {
            ctx.toast("请先阅读并同意《用户协议》和《隐私政策》")
            return
        }
        loggingIn = true
        btnLogin.isEnabled = false
        btnLogin.text = "登录中…"
        lifecycleScope.launch {
            var entered = false
            try {
                val result = ApiClient.loginBySms(phone, code)
                DemoRepository.login(
                    token = result.token,
                    refreshToken = result.refreshToken,
                    phone = phone,
                    name = result.name,
                    userId = result.userId,
                    demo = false,
                )
                com.dawnerrand.client.data.ClientRemoteRepository.refresh()
                ctx.toast("登录成功，欢迎 ${DemoRepository.currentUser.name}")
                enterMain()
                entered = true
            } catch (e: ApiException) {
                ctx.toast(e.message)
            } catch (e: IOException) {
                ctx.toast("网络不可用，请检查后端服务后重试")
            } finally {
                loggingIn = false
                if (!entered) {
                    btnLogin.isEnabled = true
                    btnLogin.text = "登录 / 注册"
                }
            }
        }
    }

    private fun enterMain() {
        startActivity(Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        })
        finish()
    }

    override fun onDestroy() {
        countDown?.cancel()
        super.onDestroy()
    }
}
