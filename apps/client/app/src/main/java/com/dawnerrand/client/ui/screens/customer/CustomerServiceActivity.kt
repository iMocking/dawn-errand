package com.dawnerrand.client.ui.screens.customer

import android.content.ActivityNotFoundException
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.view.View
import android.widget.EditText
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.help.HelpCenterActivity
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 联系客服：在线客服入口、400 热线、常见问题、意见反馈。 */
class CustomerServiceActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_customer_service)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        findViewById<View>(R.id.btnOnline).setOnClickListener { showOnlineService() }
        findViewById<View>(R.id.btnCall).setOnClickListener { callHotline() }
        findViewById<View>(R.id.rowCall).setOnClickListener { callHotline() }
        listOf(R.id.rowFaq1, R.id.rowFaq2, R.id.rowFaq3).forEach { id ->
            findViewById<View>(id).setOnClickListener {
                startActivity(Intent(this, HelpCenterActivity::class.java))
            }
        }
        findViewById<View>(R.id.btnFeedback).setOnClickListener { submitFeedback() }
    }

    private fun callHotline() {
        try {
            startActivity(Intent(Intent.ACTION_DIAL, Uri.parse("tel:4008880000")))
        } catch (e: ActivityNotFoundException) {
            ctx.toast("未找到拨号应用，请手动拨打 400-888-0000")
        }
    }

    private fun showOnlineService() {
        AlertDialog.Builder(this)
            .setTitle("在线客服 · 小曙")
            .setMessage("小曙正在接入…\n\n演示环境暂未接入实时会话，可拨打客服热线 400-888-0000（09:00-21:00），或先到帮助中心查看常见问题。")
            .setNegativeButton("取消", null)
            .setPositiveButton("去帮助中心") { _, _ ->
                startActivity(Intent(this, HelpCenterActivity::class.java))
            }
            .show()
    }

    private fun submitFeedback() {
        val input = findViewById<EditText>(R.id.etFeedback)
        val content = input.text.toString().trim()
        if (content.length < 5) {
            ctx.toast("请至少输入5个字符，方便我们了解问题")
            return
        }
        DemoRepository.submitFeedback(content)
        input.setText("")
        ctx.toast("反馈已提交，客服会尽快联系你")
    }
}
