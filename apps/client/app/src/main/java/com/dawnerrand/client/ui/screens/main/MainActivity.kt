package com.dawnerrand.client.ui.screens.main

import android.content.Intent
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.fragment.app.Fragment
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.components.TabBarView
import com.dawnerrand.client.ui.screens.errand.ErrandActivity
import com.dawnerrand.client.ui.screens.home.HomeFragment
import com.dawnerrand.client.ui.screens.orders.OrdersActivity
import com.dawnerrand.client.ui.screens.profile.ProfileFragment
import com.dawnerrand.client.ui.util.padBottomWithNavBar

/** 主页容器：底部四 Tab（首页/跑腿/订单/我的），跑腿与订单按设计稿为独立页跳转。 */
class MainActivity : AppCompatActivity() {

    private val home = HomeFragment()
    private val profile = ProfileFragment()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val tabBar = findViewById<TabBarView>(R.id.tabBar)
        tabBar.padBottomWithNavBar()

        if (savedInstanceState == null) {
            supportFragmentManager.beginTransaction()
                .add(R.id.container, home)
                .commit()
        }

        tabBar.onTabSelected = { index ->
            when (index) {
                0 -> show(home)
                1 -> {
                    tabBar.selectedIndex = 0
                    startActivity(Intent(this, ErrandActivity::class.java))
                }
                2 -> {
                    tabBar.selectedIndex = 0
                    startActivity(Intent(this, OrdersActivity::class.java))
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
