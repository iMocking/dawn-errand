package com.dawnerrand.client.ui.screens.login

import android.content.Intent
import android.os.Bundle
import android.os.CountDownTimer
import android.view.View
import android.widget.EditText
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.main.MainActivity
import com.dawnerrand.client.ui.util.toast

class LoginActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private lateinit var etPhone: EditText
    private lateinit var etCode: EditText
    private lateinit var btnCode: TextView
    private lateinit var chkAgree: TextView
    private var agreed = true
    private var countDown: CountDownTimer? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_login)

        etPhone = findViewById(R.id.etPhone)
        etCode = findViewById(R.id.etCode)
        btnCode = findViewById(R.id.btnCode)
        chkAgree = findViewById(R.id.chkAgree)
        val btnLogin = findViewById<View>(R.id.btnLogin)

        btnCode.setOnClickListener { sendCode() }
        chkAgree.setOnClickListener { agreed = !agreed; renderAgree() }
        btnLogin.setOnClickListener { login() }
        findViewById<View>(R.id.btnWechat).setOnClickListener {
            ctx.toast("微信登录建设中，请使用手机号验证码登录")
        }
        findViewById<View>(R.id.btnAlipay).setOnClickListener {
            ctx.toast("支付宝登录建设中，请使用手机号验证码登录")
        }
    }

    private fun renderAgree() {
        chkAgree.setBackgroundResource(if (agreed) R.drawable.bg_agreed else R.drawable.bg_agreed_off)
        chkAgree.text = if (agreed) "✓" else ""
    }

    private fun sendCode() {
        val phone = etPhone.text.toString().trim()
        if (phone.length != 11) {
            ctx.toast("请输入正确的11位手机号")
            return
        }
        ctx.toast("验证码已发送（演示环境任意6位数字均可登录）")
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
        val phone = etPhone.text.toString().trim()
        val code = etCode.text.toString().trim()
        if (phone.length != 11) {
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
        // 演示模式：后端短信接口就绪前任意验证码放行（与骑手端同规则）
        DemoRepository.saveLogin(this, "demo-token-" + System.currentTimeMillis(), phone)
        startActivity(Intent(this, MainActivity::class.java))
        finish()
    }

    override fun onDestroy() {
        countDown?.cancel()
        super.onDestroy()
    }
}
