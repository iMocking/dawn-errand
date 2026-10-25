package com.dawnerrand.driver.ui.screens.login

import android.graphics.Color
import android.graphics.Typeface
import android.os.Bundle
import android.text.SpannableString
import android.text.Spanned
import android.text.style.ForegroundColorSpan
import android.view.View
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.dawnerrand.driver.R
import com.dawnerrand.driver.BuildConfig
import com.dawnerrand.driver.data.api.ApiClient
import com.dawnerrand.driver.databinding.ActivityLoginBinding
import com.dawnerrand.driver.ui.screens.main.MainActivity
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.toast
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

/**
 * 骑手登录：还原 design/driver/login.html（渐变头部 + 白卡片）。
 * 支持「手机号登录 / 邮箱登录」切换：手机号走验证码（后端无短信接口时演示模式），
 * 邮箱/密码直连真实接口 POST /api/auth/login（失败透传后端 message）。
 */
class LoginActivity : AppCompatActivity() {

    private lateinit var binding: ActivityLoginBinding
    private val vm: LoginViewModel by viewModels()

    /** true=邮箱登录（默认优先），false=手机号/短信验证码登录 */
    private var emailMode = true

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.dark(Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityLoginBinding.inflate(layoutInflater)
        setContentView(binding.root)

        buildAgreementText()
        bindEvents()
        observe()
        if (BuildConfig.DEBUG) {
            binding.etEmail.setText("rider.owner@dawnerrand.local")
            binding.etPassword.setText("Rider123456")
        }
        switchMode(true)
    }

    private fun buildAgreementText() {
        val prefix = getString(R.string.login_agreement_prefix)
        val name = getString(R.string.login_agreement_name)
        val suffix = getString(R.string.login_agreement_suffix)
        val spannable = SpannableString(prefix + name + suffix)
        spannable.setSpan(
            ForegroundColorSpan(Design.PRIMARY),
            prefix.length,
            prefix.length + name.length,
            Spanned.SPAN_EXCLUSIVE_EXCLUSIVE,
        )
        binding.tvAgreement.text = spannable

        val footer = getString(R.string.login_footer_prefix) + getString(R.string.login_footer_link)
        val footerSpan = SpannableString(footer)
        footerSpan.setSpan(
            ForegroundColorSpan(Design.PRIMARY),
            getString(R.string.login_footer_prefix).length,
            footer.length,
            Spanned.SPAN_EXCLUSIVE_EXCLUSIVE,
        )
        binding.tvFooter.text = footerSpan
        binding.tvFooter.setOnClickListener {
            toast("入驻审核：请上传身份证和健康证，等待平台审核通过")
        }
    }

    private fun bindEvents() {
        binding.checkBox.setOnClickListener {
            vm.toggleAgreement()
            refreshCheck()
        }
        binding.tvGetCode.setOnClickListener {
            vm.sendCode(binding.etPhone.text.toString())
        }
        binding.btnSubmit.setOnClickListener {
            if (emailMode) {
                vm.loginByEmail(binding.etEmail.text.toString(), binding.etPassword.text.toString())
            } else {
                vm.login(binding.etPhone.text.toString(), binding.etCode.text.toString())
            }
        }
        binding.tabPhone.setOnClickListener { switchMode(false) }
        binding.tabEmail.setOnClickListener { switchMode(true) }
        refreshCheck()
    }

    /** 切换登录方式：手机号验证码 / 邮箱密码 */
    private fun switchMode(email: Boolean) {
        emailMode = email
        val phoneActive = !email
        binding.tvTabPhone.setTextColor(if (phoneActive) Design.PRIMARY else Design.TEXT_FAINT)
        binding.tvTabPhone.setTypeface(binding.tvTabPhone.typeface, if (phoneActive) Typeface.BOLD else Typeface.NORMAL)
        binding.tvTabEmail.setTextColor(if (email) Design.PRIMARY else Design.TEXT_FAINT)
        binding.tvTabEmail.setTypeface(binding.tvTabEmail.typeface, if (email) Typeface.BOLD else Typeface.NORMAL)
        binding.indPhone.visibility = if (phoneActive) View.VISIBLE else View.INVISIBLE
        binding.indEmail.visibility = if (email) View.VISIBLE else View.INVISIBLE
        binding.layoutPhone.visibility = if (phoneActive) View.VISIBLE else View.GONE
        binding.layoutEmail.visibility = if (email) View.VISIBLE else View.GONE
    }

    private fun refreshCheck() {
        binding.checkBox.alpha = if (vm.agreementChecked) 1f else 0.4f
    }

    private fun observe() {
        vm.state.observe(this) { state ->
            when (state) {
                is LoginUiState.Countdown -> {
                    binding.tvGetCode.text = "${state.seconds}s后重试"
                    binding.tvGetCode.isClickable = false
                    binding.tvGetCode.setAlpha(0.5f)
                }
                LoginUiState.Idle -> {
                    binding.tvGetCode.text = getString(R.string.login_get_code)
                    binding.tvGetCode.isClickable = true
                    binding.tvGetCode.setAlpha(1f)
                }
                is LoginUiState.Error -> toast(state.message)
                LoginUiState.Sending -> binding.btnSubmit.alpha = 0.6f
                is LoginUiState.Success -> {
                    ApiClient.accessToken?.let { token ->
                        getSharedPreferences("driver_session", MODE_PRIVATE)
                            .edit().putString("access_token", token).apply()
                    }
                    lifecycleScope.launch {
                        withContext(Dispatchers.IO) {
                            com.dawnerrand.driver.data.repository.DriverRemoteRepository.refresh()
                        }
                        MainActivity.start(this@LoginActivity, state.demoMode)
                        finish()
                    }
                }
            }
        }
    }
}
