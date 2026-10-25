package com.dawnerrand.driver.ui.screens.home

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

    fun setOnline(value: Boolean) {
        isOnline = value
        viewModelScope.launch(Dispatchers.IO) {
            DriverRemoteRepository.setOnline(value)
            DriverRemoteRepository.refreshOrders()
        }
    }
}
