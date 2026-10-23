package com.dawnerrand.driver.ui.screens.home

import androidx.lifecycle.ViewModel
import com.dawnerrand.driver.data.repository.DriverRepository

class HomeViewModel : ViewModel() {
    val rider = DriverRepository.rider
    val activeOrder = DriverRepository.activeOrder
    var isOnline: Boolean = true
        private set

    fun setOnline(value: Boolean) {
        isOnline = value
    }
}
