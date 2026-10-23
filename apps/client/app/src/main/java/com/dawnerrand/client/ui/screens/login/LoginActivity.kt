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
    private lateinit var btnCode: TextView
    private lateinit var btnLogin: TextView
    private lateinit var chkAgree: TextView
    private var agreed = true
    private var countDown: CountDownTimer? = null
    private var demoMode = false
    private var sending = false
    private var loggingIn = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_login)

        etPhone = findViewById(R.id.etPhone)
        etCode = findViewById(R.id.etCode)
        btnCode = findViewById(R.id.btnCode)
        btnLogin = findViewById(R.id.btnLogin)
        chkAgree = findViewById(R.id.chkAgree)

        if (BuildConfig.DEBUG && etPhone.text.isBlank()) {
            etPhone.setText("13888888888")
        }

        btnCode.setOnClickListener { sendCode() }
        chkAgree.setOnClickListener { agreed = !agreed; renderAgree() }
        btnLogin.setOnClickListener { login() }
        findViewById<View>(R.id.btnWechat).setOnClickListener {
            ctx.toast("微信登录建设中，请使用手机号验证码登录")
        }
        findViewById<View>(R.id.btnAlipay).setOnClickListener {
            ctx.toast("支付宝登录建设中，请使用手机号验证码登录")
        }
        findViewById<View>(R.id.tvUserAgreement).setOnClickListener {
            startActivity(Intent(this, LegalActivity::class.java)
                .putExtra(LegalActivity.EXTRA_TYPE, LegalActivity.TYPE_AGREEMENT))
        }
        findViewById<View>(R.id.tvPrivacy).setOnClickListener {
            startActivity(Intent(this, LegalActivity::class.java)
                .putExtra(LegalActivity.EXTRA_TYPE, LegalActivity.TYPE_PRIVACY))
        }
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
                demoMode = false
                if (!devCode.isNullOrBlank()) {
                    etCode.setText(devCode)
                    ctx.toast("验证码已发送（开发环境已自动填充）")
                } else {
                    ctx.toast("验证码已发送，请查收短信")
                }
                startCountdown()
            } catch (e: ApiException) {
                if (e.code == 1401 || e.code == -1) {
                    enterDemo("短信服务尚未部署，已切换演示模式：任意6位验证码可登录")
                } else {
                    btnCode.isEnabled = true
                    btnCode.text = "重新获取"
                    ctx.toast(e.message)
                }
            } catch (e: IOException) {
                enterDemo("网络不可用，已切换演示模式：任意6位验证码可登录")
            } finally {
                sending = false
            }
        }
    }

    private fun enterDemo(message: String) {
        demoMode = true
        ctx.toast(message)
        startCountdown()
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
                if (demoMode) {
                    demoLogin(phone)
                    entered = true
                    return@launch
                }
                val result = ApiClient.loginBySms(phone, code)
                DemoRepository.login(
                    token = result.token,
                    refreshToken = result.refreshToken,
                    phone = phone,
                    name = result.name,
                    userId = result.userId,
                    demo = false,
                )
                ctx.toast("登录成功，欢迎 ${DemoRepository.currentUser.name}")
                enterMain()
                entered = true
            } catch (e: ApiException) {
                if (e.code == 1401 || e.code == -1) {
                    demoLogin(phone)
                    entered = true
                } else {
                    ctx.toast(e.message)
                }
            } catch (e: IOException) {
                demoLogin(phone)
                entered = true
            } finally {
                loggingIn = false
                if (!entered) {
                    btnLogin.isEnabled = true
                    btnLogin.text = "登录 / 注册"
                }
            }
        }
    }

    private fun demoLogin(phone: String) {
        DemoRepository.login(
            token = "demo-token-" + System.currentTimeMillis(),
            refreshToken = null,
            phone = phone,
            name = DemoRepository.defaultNameFor(phone),
            userId = "",
            demo = true,
        )
        ctx.toast("演示模式登录成功，欢迎 ${DemoRepository.currentUser.name}")
        enterMain()
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
