package com.dawnerrand.driver.ui.screens.home

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dawnerrand.driver.data.repository.DriverRemoteRepository
import com.dawnerrand.driver.data.repository.DriverRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class HomeViewModel : ViewModel() {
    val rider: com.dawnerrand.driver.data.model.RiderProfile
        get() = DriverRepository.rider
    val activeOrder: com.dawnerrand.driver.data.model.ActiveOrder
        get() = DriverRepository.activeOrder
    var isOnline: Boolean = true
        private set

    private val _statusError = MutableLiveData<String?>()

    /** 上下线失败的一次性提示；不再让「请求被服务端拒绝」显示成切换成功。 */
    val statusError: LiveData<String?> = _statusError

    fun setOnline(value: Boolean) {
        isOnline = value
        viewModelScope.launch(Dispatchers.IO) {
            val resp = DriverRemoteRepository.setOnline(value)
            if (!resp.ok) _statusError.postValue(resp.message.ifBlank { "状态更新失败，请重试" })
            DriverRemoteRepository.refreshOrders()
        }
    }
}
