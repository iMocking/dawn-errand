package com.dawnerrand.driver.ui.screens.profile

import androidx.lifecycle.ViewModel
import com.dawnerrand.driver.data.repository.DriverRepository

class ProfileViewModel : ViewModel() {
    val rider: com.dawnerrand.driver.data.model.RiderProfile
        get() = DriverRepository.rider
}
