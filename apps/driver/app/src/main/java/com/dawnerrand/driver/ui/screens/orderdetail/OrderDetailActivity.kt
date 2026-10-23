package com.dawnerrand.driver.ui.screens.orderdetail

import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.driver.data.repository.DriverRepository
import com.dawnerrand.driver.databinding.ActivityOrderDetailBinding
import com.dawnerrand.driver.ui.util.padBottomWithNavBar
import com.dawnerrand.driver.ui.util.padTopWithStatusBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 配送详情：还原 design/driver/order-detail.html（取送步骤卡 + 订单信息 + 本单收入 + 底部操作栏）。
 */
class OrderDetailActivity : AppCompatActivity() {

    private lateinit var binding: ActivityOrderDetailBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityOrderDetailBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.navBar.padTopWithStatusBar()
        binding.btnBack.setOnClickListener { finish() }
        binding.bottomBar.padBottomWithNavBar(extra = 0)

        val order = DriverRepository.activeOrder
        binding.tvPickName.text = order.pickName
        binding.tvPickAddr.text = order.pickAddr
        binding.tvPickContact.text = order.pickContact
        binding.tvDropName.text = order.dropName
        binding.tvDropAddr.text = order.dropAddr
        binding.tvDropContact.text = order.dropContact

        binding.btnCallPick.setOnClickListener { toast("正在呼叫 ${order.pickContact}") }
        binding.btnCallDrop.setOnClickListener { toast("正在呼叫 ${order.dropContact}") }
        binding.btnContact.setOnClickListener { toast("正在呼叫 ${order.dropContact}") }
        binding.btnConfirm.setOnClickListener {
            toast("已确认送达，本单收入 ¥7.0 已到账")
            finish()
        }
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, OrderDetailActivity::class.java))
        }
    }
}
