package com.dawnerrand.driver.ui.screens.main

import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.view.ViewGroup
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.fragment.app.Fragment
import com.dawnerrand.driver.R
import com.dawnerrand.driver.databinding.ActivityMainBinding
import com.dawnerrand.driver.ui.screens.home.HomeFragment
import com.dawnerrand.driver.ui.screens.orderpool.OrderPoolFragment
import com.dawnerrand.driver.ui.screens.orders.OrdersFragment
import com.dawnerrand.driver.ui.screens.profile.ProfileFragment
import com.dawnerrand.driver.ui.util.dp

/**
 * 主框架：底部四 Tab（工作台 / 接单 / 订单 / 我的），还原 home.html 的 .tabbar。
 */
class MainActivity : AppCompatActivity() {

    private lateinit var binding: ActivityMainBinding

    private val tabTags = listOf("tab_home", "tab_pool", "tab_orders", "tab_profile")

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.tabBar.onTabSelected = ::switchTo

        if (savedInstanceState == null) {
            switchTo(0)
            binding.tabBar.selectedIndex = 0
        } else {
            val current = savedInstanceState.getInt("current_tab", 0)
            binding.tabBar.selectedIndex = current
            updateStatusBarStyle(current)
        }

        applySystemBarInsets()
    }

    override fun onSaveInstanceState(outState: Bundle) {
        super.onSaveInstanceState(outState)
        outState.putInt("current_tab", binding.tabBar.selectedIndex)
    }

    private fun fragmentFor(tag: String): Fragment {
        supportFragmentManager.findFragmentByTag(tag)?.let { return it }
        return when (tag) {
            "tab_home" -> HomeFragment()
            "tab_pool" -> OrderPoolFragment()
            "tab_orders" -> OrdersFragment()
            else -> ProfileFragment()
        }
    }

    private fun switchTo(index: Int) {
        val tx = supportFragmentManager.beginTransaction()
        tabTags.forEachIndexed { i, tag ->
            val f = fragmentFor(tag)
            if (!f.isAdded) tx.add(R.id.fragmentContainer, f, tag)
            if (i != index) tx.hide(f) else tx.show(f)
        }
        tx.commit()
        updateStatusBarStyle(index)
    }

    private fun updateStatusBarStyle(index: Int) {
        // 工作台 / 我的 顶部为橙色渐变 → 白色状态栏图标；接单 / 订单 白底 → 深色图标
        val light = index == 0 || index == 3
        WindowInsetsControllerCompat(window, window.decorView).isAppearanceLightStatusBars = !light
    }

    private fun applySystemBarInsets() {
        val tabBar = binding.tabBar
        ViewCompat.setOnApplyWindowInsetsListener(tabBar) { v, insets ->
            val nav = insets.getInsets(WindowInsetsCompat.Type.navigationBars()).bottom
            val baseHeight = dp(64)
            val params = v.layoutParams as ViewGroup.LayoutParams
            params.height = baseHeight + nav
            v.layoutParams = params
            v.setPadding(0, 0, 0, nav + dp(8))
            val containerParams = binding.fragmentContainer.layoutParams as ViewGroup.MarginLayoutParams
            containerParams.bottomMargin = baseHeight + nav
            binding.fragmentContainer.layoutParams = containerParams
            insets
        }
    }

    companion object {
        fun start(context: Context, demoMode: Boolean = false) {
            val intent = Intent(context, MainActivity::class.java)
            intent.putExtra("demo_mode", demoMode)
            context.startActivity(intent)
        }
    }
}
