package com.dawnerrand.store.ui.screens.finance

import android.os.Bundle
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.databinding.ActivityFinanceBinding
import com.dawnerrand.store.ui.adapter.FinanceRecordAdapter
import com.dawnerrand.store.ui.util.toast

/**
 * 收入提现：余额卡 + 提现 + 近期收支记录（finance 稿）。
 */
class FinanceActivity : AppCompatActivity() {

    private lateinit var b: ActivityFinanceBinding
    private val vm: FinanceViewModel by viewModels()
    private val adapter = FinanceRecordAdapter()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityFinanceBinding.inflate(layoutInflater)
        setContentView(b.root)

        b.backBtn.setOnClickListener { finish() }

        // 余额卡
        b.balanceValue.text = "¥${vm.balance}"
        b.todayRevenue.text = vm.todayRevenue
        b.monthRevenue.text = vm.monthRevenue
        b.monthOrders.text = vm.monthOrders

        // 提现
        b.withdrawBtn.setOnClickListener {
            toast("提现申请已提交，演示环境将在一个工作日内到账")
        }

        // 记录
        b.recordList.layoutManager = LinearLayoutManager(this)
        b.recordList.adapter = adapter
        adapter.submit(vm.records)
    }
}
