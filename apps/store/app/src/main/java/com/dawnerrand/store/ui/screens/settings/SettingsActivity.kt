package com.dawnerrand.store.ui.screens.settings

import android.content.Intent
import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.data.StoreRepository
import com.dawnerrand.store.data.StoreSession
import com.dawnerrand.store.data.StoreSettings
import com.dawnerrand.store.databinding.ActivitySettingsBinding
import com.dawnerrand.store.ui.screens.login.LoginActivity
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast
import java.util.Locale

/**
 * 设置：账号信息、通知开关、清除缓存、语言、版本、隐私政策/用户协议、退出登录。
 */
class SettingsActivity : AppCompatActivity() {

    private lateinit var b: ActivitySettingsBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivitySettingsBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.navBar.onBack = { finish() }

        b.shopName.text = StoreRepository.shop.name
        b.roleTag.text = StoreSession.roleLabel()
        b.accountValue.text = StoreSession.accountLabel()

        b.orderSwitch.isOn = StoreSettings.orderNotify
        b.orderSwitch.setOnCheckedChangeListener { checked ->
            StoreSettings.setOrderNotify(this, checked)
            toast(if (checked) "已开启新订单提醒" else "已关闭新订单提醒")
        }
        b.systemSwitch.isOn = StoreSettings.systemNotify
        b.systemSwitch.setOnCheckedChangeListener { checked ->
            StoreSettings.setSystemNotify(this, checked)
            toast(if (checked) "已开启系统通知" else "已关闭系统通知")
        }

        b.cacheValue.text = formatSize(cacheSize())
        b.cacheRow.setOnClickListener { clearCache() }
        b.languageRow.setOnClickListener { toast("当前仅支持简体中文") }
        b.privacyRow.setOnClickListener { toast("隐私政策：商家数据仅用于平台经营与结算") }
        b.agreementRow.setOnClickListener { toast("用户协议：以平台最新发布版本为准") }
        b.logoutBtn.setOnClickListener { confirmLogout() }
    }

    private fun cacheSize(): Long {
        val files = cacheDir.listFiles() ?: return 0L
        return files.sumOf { file -> file.walkTopDown().filter { it.isFile }.sumOf { it.length() } }
    }

    private fun clearCache() {
        try {
            cacheDir.listFiles()?.forEach { it.deleteRecursively() }
            b.cacheValue.text = formatSize(0L)
            Toast.makeText(this, "缓存已清除", Toast.LENGTH_SHORT).show()
        } catch (e: Exception) {
            toast("清除缓存失败，请稍后重试")
        }
    }

    private fun formatSize(bytes: Long): String = when {
        bytes >= 1024 * 1024 -> String.format(Locale.CHINA, "%.1f MB", bytes / 1024.0 / 1024.0)
        bytes >= 1024 -> String.format(Locale.CHINA, "%.1f KB", bytes / 1024.0)
        else -> "0 KB"
    }

    private fun confirmLogout() {
        AlertDialog.Builder(this)
            .setTitle("退出登录")
            .setMessage("退出后需要重新登录，确定退出当前商家账号吗？")
            .setPositiveButton("退出") { _, _ ->
                StoreSession.logout(this)
                startActivity(
                    Intent(this, LoginActivity::class.java)
                        .addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK)
                )
                finish()
            }
            .setNegativeButton("取消", null)
            .show()
    }
}
