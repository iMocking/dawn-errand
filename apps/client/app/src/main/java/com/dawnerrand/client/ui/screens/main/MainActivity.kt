package com.dawnerrand.client.ui.screens.main

import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.fragment.app.Fragment
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.components.TabBarView
import com.dawnerrand.client.ui.screens.errand.ErrandFragment
import com.dawnerrand.client.ui.screens.home.HomeFragment
import com.dawnerrand.client.ui.screens.orders.OrdersFragment
import com.dawnerrand.client.ui.screens.profile.ProfileFragment
import com.dawnerrand.client.ui.util.padBottomWithNavBar
import com.dawnerrand.client.ui.util.setLightStatusBars

/**
 * 主页容器：底部四 Tab（首页 / 跑腿 / 订单 / 我的）全部为 Tab 内容，
 * 不再把跑腿、订单跳转到独立 Activity。
 */
class MainActivity : AppCompatActivity() {

    private lateinit var tabBar: TabBarView
    private lateinit var home: HomeFragment
    private lateinit var errand: ErrandFragment
    private lateinit var orders: OrdersFragment
    private lateinit var profile: ProfileFragment
    private var current: Fragment? = null
    private var currentIndex = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        tabBar = findViewById(R.id.tabBar)
        tabBar.padBottomWithNavBar()

        if (savedInstanceState == null) {
            home = HomeFragment()
            errand = ErrandFragment.newInstance(embedded = true)
            orders = OrdersFragment.newInstance(embedded = true)
            profile = ProfileFragment()
            supportFragmentManager.beginTransaction().apply {
                add(R.id.container, home, TAG_HOME)
                add(R.id.container, errand, TAG_ERRAND).hide(errand)
                add(R.id.container, orders, TAG_ORDERS).hide(orders)
                add(R.id.container, profile, TAG_PROFILE).hide(profile)
            }.commit()
            current = home
            currentIndex = 0
        } else {
            home = supportFragmentManager.findFragmentByTag(TAG_HOME) as HomeFragment
            errand = supportFragmentManager.findFragmentByTag(TAG_ERRAND) as ErrandFragment
            orders = supportFragmentManager.findFragmentByTag(TAG_ORDERS) as OrdersFragment
            profile = supportFragmentManager.findFragmentByTag(TAG_PROFILE) as ProfileFragment
            current = listOf(home, errand, orders, profile).firstOrNull { !it.isHidden } ?: home
            currentIndex = indexOf(current ?: home)
        }
        tabBar.selectedIndex = currentIndex
        tabBar.onTabSelected = { index -> selectTab(index) }
        setLightStatusBars(currentIndex == 1 || currentIndex == 2)
    }

    /** 供首页 / 我的等 Tab 内容切换底部 Tab；status 用于订单按状态过滤 */
    fun selectTab(index: Int, status: String? = null) {
        val target: Fragment = when (index) {
            0 -> home
            1 -> errand
            2 -> orders
            3 -> profile
            else -> home
        }
        if (index == 2) {
            if (status != null) orders.showStatus(status) else orders.resetFilter()
        }
        if (target !== current) {
            supportFragmentManager.beginTransaction()
                .setReorderingAllowed(true)
                .apply {
                    current?.let { hide(it) }
                    show(target)
                }
                .commit()
            current = target
            currentIndex = index
        }
        tabBar.selectedIndex = index
        setLightStatusBars(index == 1 || index == 2)
    }

    private fun indexOf(fragment: Fragment): Int = when (fragment) {
        home -> 0
        errand -> 1
        orders -> 2
        else -> 3
    }

    companion object {
        private const val TAG_HOME = "tab_home"
        private const val TAG_ERRAND = "tab_errand"
        private const val TAG_ORDERS = "tab_orders"
        private const val TAG_PROFILE = "tab_profile"
    }
}
