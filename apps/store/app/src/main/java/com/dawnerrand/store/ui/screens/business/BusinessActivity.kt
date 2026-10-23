package com.dawnerrand.store.ui.screens.business

import android.os.Bundle
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.R
import com.dawnerrand.store.databinding.ActivityBusinessBinding
import com.dawnerrand.store.ui.util.toast

/**
 * 营业设置：开启/歇业横幅 + 开关与配送参数（business 稿）。
 */
class BusinessActivity : AppCompatActivity() {

    private lateinit var b: ActivityBusinessBinding
    private val vm: BusinessViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityBusinessBinding.inflate(layoutInflater)
        setContentView(b.root)

        b.backBtn.setOnClickListener { finish() }

        // 固定值
        b.hoursValue.text = vm.setting.hours
        b.minPriceValue.text = vm.setting.minPrice
        b.rangeValue.text = vm.setting.deliveryRange
        b.feeValue.text = vm.setting.deliveryFee
        b.announcementValue.text = vm.setting.announcement
        b.promotionValue.text = vm.setting.promotion

        // 开关状态
        b.openSwitch.setOnCheckedChangeListener(null)
        b.openSwitch.isOn = vm.open.value ?: true
        b.autoAcceptSwitch.setOnCheckedChangeListener(null)
        b.autoAcceptSwitch.isOn = vm.autoAccept.value ?: true

        b.openSwitch.setOnCheckedChangeListener {
            vm.toggleOpen()
            refreshBanner()
        }
        b.autoAcceptSwitch.setOnCheckedChangeListener {
            vm.toggleAutoAccept()
            toast(if (it) "已开启自动接单" else "已关闭自动接单")
        }

        refreshBanner()

        // 可编辑行（演示提示）
        b.hoursValue.setOnClickListener { toast("营业时间编辑：演示环境暂未开放") }
        b.minPriceValue.setOnClickListener { toast("起送价编辑：演示环境暂未开放") }
        b.rangeValue.setOnClickListener { toast("配送范围编辑：演示环境暂未开放") }
        b.feeValue.setOnClickListener { toast("配送费编辑：演示环境暂未开放") }
        b.announcementValue.setOnClickListener { toast("店铺公告编辑：演示环境暂未开放") }
        b.promotionValue.setOnClickListener { toast("满减活动编辑：演示环境暂未开放") }
    }

    private fun refreshBanner() {
        val open = vm.open.value ?: true
        b.statusBanner.setBackgroundResource(
            if (open) R.drawable.bg_banner_green else R.drawable.bg_banner_gray
        )
        b.bannerBig.text = if (open) "营业中" else "已歇业"
        b.bannerSmall.text = if (open) "今日已接42单 · 营业至21:00" else "店铺已歇业，顾客暂时无法下单"
    }
}
