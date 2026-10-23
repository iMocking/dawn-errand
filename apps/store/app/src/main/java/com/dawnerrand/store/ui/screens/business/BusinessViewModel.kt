package com.dawnerrand.store.ui.screens.business

import androidx.lifecycle.LiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.BusinessSetting
import com.dawnerrand.store.data.StoreSettings

/**
 * 营业设置：开启/歇业、营业时间、自动接单、起送价、配送范围、配送费、公告、满减（business 稿）。
 * 数据读写 StoreSettings（本地持久化），保存后 revision 触发界面刷新。
 */
class BusinessViewModel : ViewModel() {

    val revision: LiveData<Int> = StoreSettings.revision

    val setting: BusinessSetting
        get() = StoreSettings.toBusinessSetting()
}
