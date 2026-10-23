package com.dawnerrand.store.ui.screens.business

import android.os.Bundle
import android.text.InputType
import android.widget.EditText
import android.widget.LinearLayout
import androidx.activity.viewModels
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.R
import com.dawnerrand.store.data.StoreSettings
import com.dawnerrand.store.databinding.ActivityBusinessBinding
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast

/**
 * 营业设置：开启/歇业横幅 + 开关与配送参数（business 稿）。
 * 营业时间/起送价/配送范围/配送费/公告/满减均支持点击编辑并本地持久化。
 */
class BusinessActivity : AppCompatActivity() {

    private lateinit var b: ActivityBusinessBinding
    private val vm: BusinessViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityBusinessBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.backBtn.setOnClickListener { finish() }

        bindSwitches()
        bindRows()
        refresh()

        vm.revision.observe(this) { refresh() }
    }

    private fun bindSwitches() {
        b.openSwitch.isOn = vm.setting.open
        b.openSwitch.setOnCheckedChangeListener { checked ->
            StoreSettings.setOpen(this, checked)
            toast(if (checked) "店铺已开始营业" else "店铺已歇业")
        }
        b.autoAcceptSwitch.isOn = vm.setting.autoAccept
        b.autoAcceptSwitch.setOnCheckedChangeListener { checked ->
            StoreSettings.setAutoAccept(this, checked)
            toast(if (checked) "已开启自动接单" else "已关闭自动接单")
        }
    }

    private fun bindRows() {
        b.hoursValue.parentClick {
            askText("营业时间", vm.setting.hours, "如 09:00 - 21:00") { value ->
                val normalized = value.replace(Regex("\\s+"), " ")
                if (!normalized.matches(Regex("^\\d{1,2}:\\d{2} - \\d{1,2}:\\d{2}$"))) {
                    "格式应为 HH:mm - HH:mm"
                } else {
                    StoreSettings.setHours(this, normalized)
                    toast("营业时间已更新")
                    null
                }
            }
        }
        b.minPriceValue.parentClick {
            askNumber("起送价", stripNumber(vm.setting.minPrice), "元") { v ->
                StoreSettings.setMinPrice(this, "¥${format(v)}")
                toast("起送价已更新")
            }
        }
        b.rangeValue.parentClick {
            askNumber("配送范围", stripNumber(vm.setting.deliveryRange), "公里") { v ->
                StoreSettings.setDeliveryRange(this, "${format(v)}公里")
                toast("配送范围已更新")
            }
        }
        b.feeValue.parentClick {
            askNumber("配送费", stripNumber(vm.setting.deliveryFee), "元/单") { v ->
                StoreSettings.setDeliveryFee(this, "¥${format(v)}/单")
                toast("配送费已更新")
            }
        }
        b.announcementValue.parentClick {
            askText("店铺公告", vm.setting.announcement, "如：今日辣椒用完，少辣请备注", multiline = true) { value ->
                if (value.isBlank()) {
                    "店铺公告不能为空"
                } else {
                    StoreSettings.setAnnouncement(this, value)
                    toast("店铺公告已更新")
                    null
                }
            }
        }
        b.promotionValue.parentClick {
            askText("满减活动", vm.setting.promotion, "如：满25减3") { value ->
                if (value.isBlank()) {
                    "请填写满减活动"
                } else {
                    StoreSettings.setPromotion(this, value)
                    toast("满减活动已更新")
                    null
                }
            }
        }
    }

    private fun refresh() {
        val s = vm.setting
        b.hoursValue.text = s.hours
        b.minPriceValue.text = s.minPrice
        b.rangeValue.text = s.deliveryRange
        b.feeValue.text = s.deliveryFee
        b.announcementValue.text = s.announcement
        b.promotionValue.text = s.promotion

        b.openSwitch.setOnCheckedChangeListener(null)
        b.openSwitch.isOn = s.open
        b.openSwitch.setOnCheckedChangeListener { checked ->
            StoreSettings.setOpen(this, checked)
            toast(if (checked) "店铺已开始营业" else "店铺已歇业")
        }
        b.autoAcceptSwitch.setOnCheckedChangeListener(null)
        b.autoAcceptSwitch.isOn = s.autoAccept
        b.autoAcceptSwitch.setOnCheckedChangeListener { checked ->
            StoreSettings.setAutoAccept(this, checked)
            toast(if (checked) "已开启自动接单" else "已关闭自动接单")
        }
        refreshBanner()
    }

    private fun refreshBanner() {
        val open = vm.setting.open
        b.statusBanner.setBackgroundResource(
            if (open) R.drawable.bg_banner_green else R.drawable.bg_banner_gray
        )
        b.bannerBig.text = if (open) "营业中" else "已歇业"
        b.bannerSmall.text = if (open) {
            "今日已接42单 · 营业至${StoreSettings.closeTime()}"
        } else {
            "店铺已歇业，顾客暂时无法下单"
        }
    }

    private fun stripNumber(value: String): String =
        value.replace(Regex("[^0-9.]"), "")

    private fun format(v: Double): String =
        if (v % 1.0 == 0.0) v.toInt().toString() else String.format("%.1f", v)

    private fun askNumber(title: String, current: String, unit: String, onSave: (Double) -> Unit) {
        val input = EditText(this).apply {
            inputType = InputType.TYPE_CLASS_NUMBER or InputType.TYPE_NUMBER_FLAG_DECIMAL
            setText(current)
            hint = "请输入$title（$unit）"
            setPadding(dp(20), dp(8), dp(20), dp(8))
            setSelection(text.length)
        }
        val dialog = AlertDialog.Builder(this)
            .setTitle(title)
            .setView(input)
            .setPositiveButton("保存", null)
            .setNegativeButton("取消", null)
            .create()
        dialog.setOnShowListener {
            dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener {
                val v = input.text.toString().toDoubleOrNull()
                if (v == null || v <= 0) {
                    input.error = "请输入大于 0 的数值"
                } else {
                    onSave(v)
                    dialog.dismiss()
                }
            }
        }
        dialog.show()
    }

    private fun askText(
        title: String,
        current: String,
        hintText: String,
        multiline: Boolean = false,
        onSave: (String) -> String?,
    ) {
        val input = EditText(this).apply {
            setText(current)
            hint = hintText
            setPadding(dp(20), dp(8), dp(20), dp(8))
            if (multiline) {
                inputType = InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_FLAG_MULTI_LINE
                minLines = 3
                gravity = android.view.Gravity.TOP
            } else {
                inputType = InputType.TYPE_CLASS_TEXT
                setSingleLine()
            }
            setSelection(text.length)
        }
        val dialog = AlertDialog.Builder(this)
            .setTitle(title)
            .setView(input)
            .setPositiveButton("保存", null)
            .setNegativeButton("取消", null)
            .create()
        dialog.setOnShowListener {
            dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener {
                val error = onSave(input.text.toString().trim())
                if (error != null) {
                    input.error = error
                } else {
                    dialog.dismiss()
                }
            }
        }
        dialog.show()
    }

    /** 让整行（label + value）可点击 */
    private fun android.view.View.parentClick(action: () -> Unit) {
        val parent = parent
        if (parent is LinearLayout) {
            parent.isClickable = true
            parent.setOnClickListener { action() }
        } else {
            setOnClickListener { action() }
        }
    }
}
