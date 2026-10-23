package com.dawnerrand.client.ui.screens.editor

import android.os.Bundle
import android.view.View
import android.widget.EditText
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 个人资料：修改昵称（手机号只读）。 */
class ProfileEditActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_profile_edit)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        val etName = findViewById<EditText>(R.id.etName)
        etName.setText(DemoRepository.currentUser.name)
        findViewById<TextView>(R.id.tvPhone).text = DemoRepository.displayPhone()

        findViewById<View>(R.id.btnSave).setOnClickListener {
            val name = etName.text.toString().trim()
            when {
                name.length < 2 -> ctx.toast("昵称至少需要2个字符")
                name.length > 12 -> ctx.toast("昵称最多12个字符")
                else -> {
                    DemoRepository.updateUserName(name)
                    ctx.toast("资料已保存")
                    finish()
                }
            }
        }
    }
}
