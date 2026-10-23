package com.dawnerrand.store.ui.screens.main

import android.content.Intent
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.fragment.app.Fragment
import com.dawnerrand.store.R
import com.dawnerrand.store.databinding.ActivityMainBinding
import com.dawnerrand.store.ui.screens.dashboard.DashboardFragment
import com.dawnerrand.store.ui.screens.menu.MenuActivity
import com.dawnerrand.store.ui.screens.orders.OrdersActivity
import com.dawnerrand.store.ui.screens.profile.ProfileFragment
import com.dawnerrand.store.ui.util.padBottomWithNavBar
import com.dawnerrand.store.ui.util.setLightStatusBars

/**
 * 主容器：底部四 Tab（工作台/订单/菜品/我的）。
 * 订单与菜品按设计稿为独立全屏页面（orders.html / menu.html 无底部 Tab）。
 */
class MainActivity : AppCompatActivity() {

    private val dashboard = DashboardFragment()
    private val profile = ProfileFragment()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val b = ActivityMainBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(false)

        b.tabBar.padBottomWithNavBar()

        if (savedInstanceState == null) {
            supportFragmentManager.beginTransaction()
                .add(R.id.container, dashboard)
                .commit()
        }

        b.tabBar.onTabSelected = { index ->
            when (index) {
                0 -> show(dashboard)
                1 -> {
                    b.tabBar.selectedIndex = 0
                    startActivity(Intent(this, OrdersActivity::class.java))
                }
                2 -> {
                    b.tabBar.selectedIndex = 0
                    startActivity(Intent(this, MenuActivity::class.java))
                }
                3 -> show(profile)
            }
        }
    }

    private fun show(fragment: Fragment) {
        supportFragmentManager.beginTransaction()
            .replace(R.id.container, fragment)
            .commit()
    }
}
