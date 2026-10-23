package com.dawnerrand.client.ui.screens.addressedit

import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.Address
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.components.SwitchView
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.confirmDialog
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 新增 / 编辑地址：联系人、手机号、详细地址、标签、默认地址、删除。 */
class AddressEditActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private val tagOptions = listOf("家", "公司", "学校", "其他")
    private var addressId: String? = null
    private var selectedTag = "家"
    private val tagViews = mutableListOf<TextView>()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_address_edit)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        addressId = intent.getStringExtra(EXTRA_ADDRESS_ID)
        val existing: Address? = addressId?.let { DemoRepository.addressById(it) }
        val isEdit = existing != null
        findViewById<TextView>(R.id.tvTitle).text = if (isEdit) "编辑地址" else "新增地址"

        val etName = findViewById<EditText>(R.id.etName)
        val etPhone = findViewById<EditText>(R.id.etPhone)
        val etDetail = findViewById<EditText>(R.id.etDetail)
        val switchDefault = findViewById<SwitchView>(R.id.switchDefault)
        val btnDelete = findViewById<View>(R.id.btnDelete)

        if (existing != null) {
            etName.setText(existing.name)
            etPhone.setText(existing.phone)
            etDetail.setText(existing.detail)
            selectedTag = existing.tag
            switchDefault.isOn = existing.isDefault
            btnDelete.visibility = View.VISIBLE
            btnDelete.setOnClickListener { deleteAddress(existing) }
        } else {
            etName.setText(DemoRepository.currentUser.name)
            etPhone.setText(DemoRepository.currentUser.phone)
            switchDefault.isOn = DemoRepository.addresses.isEmpty()
        }

        val tagRow = findViewById<LinearLayout>(R.id.tagRow)
        tagOptions.forEach { tagRow.addView(buildTagChip(it)) }
        renderTags()

        findViewById<View>(R.id.btnSave).setOnClickListener {
            val name = etName.text.toString().trim()
            val phone = etPhone.text.toString().trim()
            val detail = etDetail.text.toString().trim()
            when {
                name.isEmpty() -> ctx.toast("请填写联系人")
                phone.length != 11 || !phone.startsWith("1") -> ctx.toast("请填写正确的11位手机号")
                detail.isEmpty() -> ctx.toast("请填写详细地址")
                else -> {
                    DemoRepository.saveAddress(
                        id = addressId,
                        name = name,
                        phone = phone,
                        detail = detail,
                        tag = selectedTag,
                        isDefault = switchDefault.isOn,
                    )
                    ctx.toast("地址已保存")
                    setResult(RESULT_OK)
                    finish()
                }
            }
        }
    }

    private fun buildTagChip(text: String): TextView = TextView(this).apply {
        this.text = text
        textSize = 13f
        gravity = Gravity.CENTER
        setPadding(ctx.dp(16), ctx.dp(8), ctx.dp(16), ctx.dp(8))
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT)
        (layoutParams as LinearLayout.LayoutParams).rightMargin = ctx.dp(10)
        isClickable = true
        setOnClickListener {
            selectedTag = text
            renderTags()
        }
        tagViews.add(this)
    }

    private fun renderTags() {
        tagViews.forEach { view ->
            val active = view.text.toString() == selectedTag
            view.setBackgroundResource(if (active) R.drawable.bg_chip_active else R.drawable.bg_chip)
            view.setTextColor(if (active) Design.WHITE else Design.TEXT_SECONDARY)
        }
    }

    private fun deleteAddress(address: Address) {
        ctx.confirmDialog("删除地址", "确定删除「${address.name} · ${address.shortDetail}」吗？") {
            DemoRepository.deleteAddress(address.id)
            ctx.toast("地址已删除")
            setResult(RESULT_OK)
            finish()
        }
    }

    companion object {
        const val EXTRA_ADDRESS_ID = "address_id"
    }
}
