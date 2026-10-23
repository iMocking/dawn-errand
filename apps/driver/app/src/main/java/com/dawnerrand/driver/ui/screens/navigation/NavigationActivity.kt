package com.dawnerrand.driver.ui.screens.navigation

import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.driver.databinding.ActivityNavigationBinding
import com.dawnerrand.driver.ui.util.padBottomWithNavBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 导航页：还原 design/driver/navigation.html（模拟地图 + 路线 + ETA + 开始导航）。
 */
class NavigationActivity : AppCompatActivity() {

    private lateinit var binding: ActivityNavigationBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.dark(Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityNavigationBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.btnNavBack.setOnClickListener { finish() }
        binding.btnStartNav.setOnClickListener { toast("开始导航，请沿当前路线前往取餐点") }

        binding.root.padBottomWithNavBar(extra = 0)
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, NavigationActivity::class.java))
        }
    }
}
