package com.dawnerrand.driver.ui.screens.settings

import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.driver.R
import com.dawnerrand.driver.data.api.ApiClient
import com.dawnerrand.driver.databinding.ActivitySettingsBinding
import com.dawnerrand.driver.ui.screens.login.LoginActivity
import com.dawnerrand.driver.ui.util.padTopWithStatusBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 设置：账号与安全、通知开关（持久化）、通用（清除缓存）、关于、退出登录。
 */
class SettingsActivity : AppCompatActivity() {

    private lateinit var binding: ActivitySettingsBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivitySettingsBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.navBar.padTopWithStatusBar()
        binding.btnBack.setOnClickListener { finish() }

        // 通知开关（持久化）
        val prefs = getSharedPreferences("driver_settings", MODE_PRIVATE)
        binding.switchOrder.setChecked(prefs.getBoolean("notify_order", true))
        binding.switchOrder.onCheckedChange = { on ->
            prefs.edit().putBoolean("notify_order", on).apply()
            toast(if (on) "已开启新订单提醒" else "已关闭新订单提醒")
        }
        binding.switchSystem.setChecked(prefs.getBoolean("notify_system", true))
        binding.switchSystem.onCheckedChange = { on ->
            prefs.edit().putBoolean("notify_system", on).apply()
            toast(if (on) "已开启系统通知" else "已关闭系统通知")
        }

        // 账号
        binding.cardAccount.setOnClickListener { }
        binding.rowPassword.setOnClickListener { toast("修改密码：请先通过客服热线核验身份") }

        // 清除缓存
        binding.rowCache.setOnClickListener {
            val before = cacheSizeMb()
            cacheDir.listFiles()?.forEach { it.deleteRecursively() }
            binding.tvCache.text = "0 KB"
            toast("已清除 $before 缓存")
        }

        // 语言
        binding.rowLanguage.setOnClickListener { toast("当前仅支持简体中文") }

        // 关于
        binding.rowPrivacy.setOnClickListener { toast("《隐私政策》请访问官网查看") }
        binding.rowAgreement.setOnClickListener { toast("《用户协议》请访问官网查看") }

        // 退出登录
        binding.btnLogout.setOnClickListener {
            AlertDialog.Builder(this)
                .setMessage("确定要退出登录吗？")
                .setPositiveButton("退出") { _, _ -> logout() }
                .setNegativeButton("取消", null)
                .show()
        }
    }

    private fun logout() {
        ApiClient.accessToken = null
        getSharedPreferences("driver_session", MODE_PRIVATE).edit().clear().apply()
        toast("已退出登录")
        val login = Intent(this, LoginActivity::class.java)
            .addFlags(Intent.FLAG_ACTIVITY_CLEAR_TASK or Intent.FLAG_ACTIVITY_NEW_TASK)
        startActivity(login)
        finish()
    }

    private fun cacheSizeMb(): String {
        val bytes = cacheDir.listFiles()?.sumOf { it.length() } ?: 0L
        val kb = bytes / 1024
        return if (kb < 1024) "${kb} KB" else String.format("%.1f MB", kb / 1024.0)
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, SettingsActivity::class.java))
        }
    }
}
