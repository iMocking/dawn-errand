package com.dawnerrand.store.ui.screens.profile

import androidx.lifecycle.ViewModel
import com.dawnerrand.store.data.StoreRepository

/**
 * 商家中心：店铺信息 + 资质/结算/经营/评价/客服/设置菜单（profile 稿）。
 */
class ProfileViewModel : ViewModel() {

    val shop: com.dawnerrand.store.data.ShopInfo
        get() = StoreRepository.shop
    val menus = StoreRepository.profileMenus
}
