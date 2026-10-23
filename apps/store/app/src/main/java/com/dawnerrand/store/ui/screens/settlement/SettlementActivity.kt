package com.dawnerrand.store.ui.screens.settlement

import android.os.Bundle
import android.view.View
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.data.SettlementStore
import com.dawnerrand.store.databinding.ActivitySettlementBinding
import com.dawnerrand.store.databinding.DialogBindCardBinding
import com.dawnerrand.store.ui.util.setLightStatusBars
import com.dawnerrand.store.ui.util.toast

/**
 * 结算账户：微信钱包（默认）/ 银行卡（绑定、更换、解绑、切换默认账户）。
 * 绑定信息本地持久化，提现页按钮文案与流水跟随当前账户。
 */
class SettlementActivity : AppCompatActivity() {

    private lateinit var b: ActivitySettlementBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivitySettlementBinding.inflate(layoutInflater)
        setContentView(b.root)
        setLightStatusBars(true)

        b.navBar.onBack = { finish() }

        b.wechatRow.setOnClickListener {
            SettlementStore.useWechat(this)
            refresh()
            toast("已切换为微信钱包结算")
        }
        b.bankRow.setOnClickListener {
            if (SettlementStore.hasBank()) {
                SettlementStore.useBank(this)
                refresh()
                toast("已切换为银行卡结算")
            } else {
                showBindDialog()
            }
        }
        b.bindBtn.setOnClickListener { showBindDialog() }
        b.unbindBtn.setOnClickListener { confirmUnbind() }

        refresh()
    }

    private fun refresh() {
        b.accountValue.text = SettlementStore.summary(this)
        b.accountSub.text = if (SettlementStore.type == "bank") {
            "银行卡 ${SettlementStore.maskedCard()} · 变更后次日生效"
        } else {
            "提现默认打款至微信钱包 · 变更后次日生效"
        }
        b.bankSub.text = if (SettlementStore.hasBank()) {
            "${SettlementStore.bankName} · ${SettlementStore.maskedCard()}"
        } else {
            "未绑定，绑定后可用于提现"
        }
        b.wechatCheck.visibility =
            if (SettlementStore.type != "bank") View.VISIBLE else View.INVISIBLE
        b.bankCheck.visibility =
            if (SettlementStore.type == "bank") View.VISIBLE else View.INVISIBLE
        b.bindBtn.text = if (SettlementStore.hasBank()) "更换银行卡" else "绑定银行卡"
        b.unbindBtn.visibility = if (SettlementStore.hasBank()) View.VISIBLE else View.GONE
    }

    private fun showBindDialog() {
        val binding = DialogBindCardBinding.inflate(layoutInflater)
        binding.holderInput.setText(SettlementStore.holder)
        binding.bankInput.setText(SettlementStore.bankName)
        binding.cardInput.setText(SettlementStore.cardNo)

        val dialog = AlertDialog.Builder(this)
            .setTitle(if (SettlementStore.hasBank()) "更换银行卡" else "绑定银行卡")
            .setView(binding.root)
            .setPositiveButton("保存", null)
            .setNegativeButton("取消", null)
            .create()
        dialog.setOnShowListener {
            dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener {
                val holder = binding.holderInput.text.toString().trim()
                val bank = binding.bankInput.text.toString().trim()
                val card = binding.cardInput.text.toString().trim()
                var valid = true
                if (holder.isEmpty()) {
                    binding.holderInput.error = "请填写持卡人姓名"; valid = false
                }
                if (bank.isEmpty()) {
                    binding.bankInput.error = "请填写开户银行"; valid = false
                }
                if (card.filter { it.isDigit() }.length < 12) {
                    binding.cardInput.error = "银行卡号至少 12 位"; valid = false
                }
                if (!valid) return@setOnClickListener
                val error = SettlementStore.saveBank(this, bank, card, holder)
                if (error != null) {
                    toast(error)
                } else {
                    dialog.dismiss()
                    refresh()
                    toast("银行卡已绑定，可在提现时使用")
                }
            }
        }
        dialog.show()
    }

    private fun confirmUnbind() {
        AlertDialog.Builder(this)
            .setTitle("解绑银行卡")
            .setMessage("解绑后将切换回微信钱包结算，确定解绑 ${SettlementStore.bankName} 尾号 ${SettlementStore.last4()} 吗？")
            .setPositiveButton("解绑") { _, _ ->
                SettlementStore.removeBank(this)
                refresh()
                toast("银行卡已解绑")
            }
            .setNegativeButton("取消", null)
            .show()
    }
}
