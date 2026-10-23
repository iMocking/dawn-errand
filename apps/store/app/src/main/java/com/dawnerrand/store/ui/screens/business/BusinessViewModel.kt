package com.dawnerrand.store.ui.screens.business

import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.BusinessSetting
import com.dawnerrand.store.data.StoreRepository

/**
 * 营业设置：开启/歇业、营业时间、自动接单、起送价、配送范围、配送费、公告、满减（business 稿）。
 */
class BusinessViewModel : ViewModel() {

    val setting: BusinessSetting = StoreRepository.business

    private val _open = MutableLiveData(setting.open)
    val open = _open

    private val _autoAccept = MutableLiveData(setting.autoAccept)
    val autoAccept = _autoAccept

    fun toggleOpen() {
        _open.value = !(_open.value ?: true)
    }

    fun toggleAutoAccept() {
        _autoAccept.value = !(_autoAccept.value ?: true)
    }
}
