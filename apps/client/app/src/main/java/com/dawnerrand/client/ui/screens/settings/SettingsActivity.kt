package com.dawnerrand.client.ui.screens.settings

import android.content.Intent
import android.os.Bundle
import android.view.View
import android.widget.TextView
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.BuildConfig
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.components.SwitchView
import com.dawnerrand.client.ui.screens.customer.CustomerServiceActivity
import com.dawnerrand.client.ui.screens.editor.ProfileEditActivity
import com.dawnerrand.client.ui.screens.help.HelpCenterActivity
import com.dawnerrand.client.ui.screens.legal.LegalActivity
import com.dawnerrand.client.ui.screens.login.LoginActivity
import com.dawnerrand.client.ui.util.confirmDialog
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 设置：账号资料、消息通知、缓存、协议、帮助、客服、退出登录。 */
class SettingsActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_settings)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        findViewById<View>(R.id.rowAccount).setOnClickListener {
            startActivity(Intent(this, ProfileEditActivity::class.java))
        }
        findViewById<View>(R.id.rowHelp).setOnClickListener {
            startActivity(Intent(this, HelpCenterActivity::class.java))
        }
        findViewById<View>(R.id.rowCustomer).setOnClickListener {
            startActivity(Intent(this, CustomerServiceActivity::class.java))
        }
        findViewById<View>(R.id.rowAgreement).setOnClickListener {
            startActivity(Intent(this, LegalActivity::class.java)
                .putExtra(LegalActivity.EXTRA_TYPE, LegalActivity.TYPE_AGREEMENT))
        }
        findViewById<View>(R.id.rowPrivacy).setOnClickListener {
            startActivity(Intent(this, LegalActivity::class.java)
                .putExtra(LegalActivity.EXTRA_TYPE, LegalActivity.TYPE_PRIVACY))
        }

        bindSwitch(R.id.switchOrder, DemoRepository.settings.orderNotify) {
            DemoRepository.updateSettings(orderNotify = it)
        }
        bindSwitch(R.id.switchPromo, DemoRepository.settings.promoNotify) {
            DemoRepository.updateSettings(promoNotify = it)
        }
        bindSwitch(R.id.switchSound, DemoRepository.settings.sound) {
            DemoRepository.updateSettings(sound = it)
        }

        findViewById<TextView>(R.id.tvCache).text = cacheText()
        findViewById<View>(R.id.rowClearCache).setOnClickListener {
            DemoRepository.clearSearchHistory()
            findViewById<TextView>(R.id.tvCache).text = "0.0MB"
            ctx.toast("缓存已清除")
        }
        findViewById<TextView>(R.id.tvVersion).text = "v${BuildConfig.VERSION_NAME}"
        findViewById<View>(R.id.rowAbout).setOnClickListener { showAbout() }
        findViewById<View>(R.id.btnLogout).setOnClickListener { confirmLogout() }
        refreshAccount()
    }

    override fun onResume() {
        super.onResume()
        refreshAccount()
    }

    private fun refreshAccount() {
        findViewById<TextView>(R.id.tvAccountName).text = DemoRepository.currentUser.name
        findViewById<TextView>(R.id.tvAccountPhone).text = DemoRepository.maskPhone(
            DemoRepository.currentUser.phone)
    }

    private fun bindSwitch(id: Int, value: Boolean, onChange: (Boolean) -> Unit) {
        val sw = findViewById<SwitchView>(id)
        sw.onCheckedChange = onChange
        sw.isOn = value
    }

    private fun cacheText(): String =
        "%.1fMB".format(12.0 + DemoRepository.searchHistory.size * 0.3)

    private fun showAbout() {
        AlertDialog.Builder(this)
            .setTitle("关于曙光跑腿")
            .setMessage(
                "曙光跑腿 v${BuildConfig.VERSION_NAME}\n\n" +
                    "乡镇外卖 · 跑腿代办 · 随叫随到\n\n" +
                    "客服热线：400-888-0000\n" +
                    "服务时间：09:00-21:00\n\n" +
                    "© 2026 曙光跑腿"
            )
            .setPositiveButton("知道了", null)
            .show()
    }

    private fun confirmLogout() {
        ctx.confirmDialog("退出登录", "退出后仍会保留本机演示数据，确定退出吗？") {
            DemoRepository.logout(this)
            ctx.toast("已退出登录")
            startActivity(Intent(this, LoginActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
            })
            finish()
        }
    }
}
