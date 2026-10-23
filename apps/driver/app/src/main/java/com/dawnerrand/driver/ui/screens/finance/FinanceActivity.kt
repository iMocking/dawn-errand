package com.dawnerrand.driver.ui.screens.finance

import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.view.LayoutInflater
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.driver.R
import com.dawnerrand.driver.data.repository.CardWalletRepository
import com.dawnerrand.driver.data.repository.DriverRepository
import com.dawnerrand.driver.databinding.ActivityFinanceBinding
import com.dawnerrand.driver.databinding.ItemFinanceRecordBinding
import com.dawnerrand.driver.ui.screens.cardwallet.CardWalletActivity
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.padTopWithStatusBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 收入提现：还原 design/driver/finance.html（余额渐变卡 + 提现按钮 + 今日流水）。
 */
class FinanceActivity : AppCompatActivity() {

    private lateinit var binding: ActivityFinanceBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityFinanceBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.navBar.padTopWithStatusBar()
        binding.btnBack.setOnClickListener { finish() }

        val rider = DriverRepository.rider
        binding.tvBalance.text = rider.balance
        binding.tvTodayIncome.text = rider.todayIncome
        binding.tvMonthIncome.text = rider.monthIncome
        binding.tvMonthOrders.text = rider.monthOrders

        binding.btnWithdraw.setOnClickListener {
            // 收入提现闭环：无绑定银行卡时引导前往卡包管理
            CardWalletRepository.load(this)
            if (CardWalletRepository.bankCards.isEmpty()) {
                toast("请先绑定银行卡，再发起提现")
                CardWalletActivity.start(this)
            } else {
                toast("提现申请已提交，预计1个工作日内到账")
            }
        }

        buildRecords()
    }

    private fun buildRecords() {
        val inflater = LayoutInflater.from(this)
        binding.recordList.removeAllViews()
        DriverRepository.financeRecords.forEach { record ->
            val item = ItemFinanceRecordBinding.inflate(inflater, binding.recordList, false)
            item.tvRecordTitle.text = record.title
            item.tvRecordTime.text = record.time
            item.tvRecordAmount.text = record.amount
            item.tvRecordAmount.setTextColor(
                if (record.incoming) Design.SUCCESS else Design.TEXT_PRIMARY
            )
            val lp = binding.recordList.layoutParams?.let {
                android.widget.LinearLayout.LayoutParams(
                    android.widget.LinearLayout.LayoutParams.MATCH_PARENT,
                    android.widget.LinearLayout.LayoutParams.WRAP_CONTENT,
                )
            } ?: return@forEach
            lp.bottomMargin = 8
            item.root.layoutParams = lp
            binding.recordList.addView(item.root)
        }
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, FinanceActivity::class.java))
        }
    }
}
