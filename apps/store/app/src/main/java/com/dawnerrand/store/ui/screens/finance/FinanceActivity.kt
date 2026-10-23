package com.dawnerrand.store.ui.screens.finance

import android.os.Bundle
import android.text.InputType
import android.view.Gravity
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.viewModels
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.data.FinanceStore
import com.dawnerrand.store.databinding.ActivityFinanceBinding
import com.dawnerrand.store.ui.adapter.FinanceRecordAdapter
import com.dawnerrand.store.ui.theme.Design
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast

/**
 * 收入提现：余额卡 + 提现 + 近期收支记录（finance 稿）。
 * 提现闭环：输入金额 → 校验 → 扣减余额 → 追加流水（本地持久化）。
 */
class FinanceActivity : AppCompatActivity() {

    private lateinit var b: ActivityFinanceBinding
    private val vm: FinanceViewModel by viewModels()
    private val adapter = FinanceRecordAdapter()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityFinanceBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.backBtn.setOnClickListener { finish() }

        b.todayRevenue.text = vm.todayRevenue
        b.monthRevenue.text = vm.monthRevenue
        b.monthOrders.text = vm.monthOrders
        b.feeTip.text = vm.feeTip()

        b.recordList.layoutManager = LinearLayoutManager(this)
        b.recordList.adapter = adapter

        vm.balance.observe(this) { value ->
            b.balanceValue.text = "¥${FinanceStore.formatMoney(value)}"
        }
        vm.records.observe(this) { adapter.submit(it) }

        b.withdrawBtn.setOnClickListener { showWithdrawDialog() }

        refreshTarget()
    }

    override fun onResume() {
        super.onResume()
        refreshTarget()
    }

    private fun refreshTarget() {
        b.withdrawBtn.text = vm.targetText(this)
    }

    private fun showWithdrawDialog() {
        val balance = vm.balance.value ?: 0.0

        val content = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(20), dp(8), dp(20), dp(4))
        }
        val balanceTip = TextView(this).apply {
            text = "可提现余额 ¥${FinanceStore.formatMoney(balance)}"
            textSize = 13f
            setTextColor(Design.TEXT_TERTIARY)
        }
        val inputRow = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, dp(12), 0, 0)
        }
        val input = EditText(this).apply {
            inputType = InputType.TYPE_CLASS_NUMBER or InputType.TYPE_NUMBER_FLAG_DECIMAL
            hint = "请输入提现金额"
            textSize = 18f
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        val allBtn = TextView(this).apply {
            text = "全部提现"
            textSize = 13f
            setTextColor(Design.PRIMARY)
            setPadding(dp(12), dp(6), 0, dp(6))
            setOnClickListener {
                input.setText(FinanceStore.formatMoney(balance).replace(",", ""))
                input.setSelection(input.text.length)
            }
        }
        inputRow.addView(input)
        inputRow.addView(allBtn)

        val feeTip = TextView(this).apply {
            text = "单笔最低 ¥1.00，手续费 ${FinanceStore.feeRateText()}，预计 1 个工作日内到账"
            textSize = 11f
            setTextColor(Design.TEXT_TERTIARY)
            setPadding(0, dp(8), 0, 0)
        }
        content.addView(balanceTip)
        content.addView(inputRow)
        content.addView(feeTip)

        val dialog = AlertDialog.Builder(this)
            .setTitle(vm.targetText(this))
            .setView(content)
            .setPositiveButton("确认提现", null)
            .setNegativeButton("取消", null)
            .create()
        dialog.setOnShowListener {
            dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener {
                val result = vm.withdraw(this, input.text.toString())
                if (!result.ok) {
                    input.error = result.message
                } else {
                    dialog.dismiss()
                    toast(result.message)
                }
            }
        }
        dialog.show()
    }
}
