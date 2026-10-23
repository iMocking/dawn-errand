package com.dawnerrand.driver.ui.screens.cardwallet

import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.widget.EditText
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.driver.R
import com.dawnerrand.driver.data.repository.BankCard
import com.dawnerrand.driver.data.repository.CardWalletRepository
import com.dawnerrand.driver.databinding.ActivityCardWalletBinding
import com.dawnerrand.driver.ui.theme.Design
import com.dawnerrand.driver.ui.util.dp
import com.dawnerrand.driver.ui.util.padTopWithStatusBar
import com.dawnerrand.driver.ui.util.toast

/**
 * 卡包管理：健康证 / 身份证（唯一，模拟卡片样式 + 有效期起止）+ 银行卡（多卡增删，含有效期）。
 * 收入提现闭环：提现时无银行卡则引导至此绑定。
 */
class CardWalletActivity : AppCompatActivity() {

    private lateinit var binding: ActivityCardWalletBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityCardWalletBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.navBar.padTopWithStatusBar()
        binding.btnBack.setOnClickListener { finish() }

        CardWalletRepository.load(this)

        // 证件模拟卡片：持证人 / 编号 / 有效期起止
        val health = CardWalletRepository.healthCert
        binding.tvHealthHolder.text = health.name
        binding.tvHealthNo.text = health.certNo
        binding.tvHealthValid.text = "${health.validFrom} ~ ${health.validUntil}"
        val idCert = CardWalletRepository.idCert
        binding.tvIdHolder.text = idCert.name
        binding.tvIdNo.text = idCert.certNo
        binding.tvIdValid.text = "${idCert.validFrom} ~ ${idCert.validUntil}"

        buildBankCards()
        binding.btnAdd.setOnClickListener { showAddDialog() }
    }

    private fun buildBankCards() {
        binding.llBankCards.removeAllViews()
        CardWalletRepository.bankCards.forEachIndexed { index, card ->
            val row = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                gravity = Gravity.CENTER_VERTICAL
                background = getDrawable(R.drawable.bg_card_bank)
                setPadding(dp(16), dp(14), dp(12), dp(14))
                val lp = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.MATCH_PARENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                )
                lp.bottomMargin = dp(12)
                layoutParams = lp
            }

            val info = LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            }

            val bank = TextView(this).apply {
                text = card.bank
                setTextColor(Design.WHITE)
                textSize = 15f
                setTypeface(typeface, android.graphics.Typeface.BOLD)
            }
            info.addView(bank)

            val cardNo = TextView(this).apply {
                text = maskCardNo(card.cardNo)
                setTextColor(Color.argb(230, 255, 255, 255))
                textSize = 13f
                val lp = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                )
                lp.topMargin = dp(4)
                layoutParams = lp
            }
            info.addView(cardNo)

            // 底部两列：持卡人 / 有效期
            val bottom = LinearLayout(this).apply {
                orientation = LinearLayout.HORIZONTAL
                val lp = LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.MATCH_PARENT,
                    LinearLayout.LayoutParams.WRAP_CONTENT,
                )
                lp.topMargin = dp(10)
                layoutParams = lp
            }

            fun labelCol(title: String, value: String): LinearLayout {
                val col = LinearLayout(this@CardWalletActivity).apply {
                    orientation = LinearLayout.VERTICAL
                    layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
                }
                col.addView(TextView(this@CardWalletActivity).apply {
                    text = title
                    setTextColor(Color.argb(128, 255, 255, 255))
                    textSize = 11f
                })
                col.addView(TextView(this@CardWalletActivity).apply {
                    text = value
                    setTextColor(Design.WHITE)
                    textSize = 12f
                    val lp = LinearLayout.LayoutParams(
                        LinearLayout.LayoutParams.WRAP_CONTENT,
                        LinearLayout.LayoutParams.WRAP_CONTENT,
                    )
                    lp.topMargin = dp(2)
                    layoutParams = lp
                })
                return col
            }

            bottom.addView(labelCol("持卡人", card.holder))
            bottom.addView(labelCol("有效期", card.expiry))
            info.addView(bottom)

            row.addView(info)

            val close = ImageView(this).apply {
                setImageResource(R.drawable.ic_close)
                setColorFilter(Design.WHITE)
                setPadding(dp(6), dp(6), dp(6), dp(6))
                val lp = LinearLayout.LayoutParams(dp(32), dp(32))
                layoutParams = lp
            }
            close.setOnClickListener { showUnbindDialog(index, card) }
            row.addView(close)

            binding.llBankCards.addView(row)
        }
        binding.tvEmpty.visibility = if (CardWalletRepository.bankCards.isEmpty()) View.VISIBLE else View.GONE
    }

    /** 添加银行卡弹窗：输入框样式对齐登录页（label + 透明底下划线输入行） */
    private fun showAddDialog() {
        val body = layoutInflater.inflate(R.layout.dialog_add_bank, null)
        val holderEt = body.findViewById<EditText>(R.id.etHolder)
        val cardNoEt = body.findViewById<EditText>(R.id.etCardNo)
        val bankEt = body.findViewById<EditText>(R.id.etBank)
        val expiryEt = body.findViewById<EditText>(R.id.etExpiry)

        val dialog = AlertDialog.Builder(this)
            .setTitle(R.string.wallet_add_bank)
            .setView(body)
            .setNegativeButton(R.string.wallet_cancel, null)
            .setPositiveButton(R.string.wallet_confirm, null)
            .create()
        dialog.setOnShowListener {
            // 校验失败时保持对话框打开，避免输入内容丢失
            dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener {
                val holder = holderEt.text.toString().trim()
                val cardNo = cardNoEt.text.toString().trim()
                val bank = bankEt.text.toString().trim()
                val expiry = expiryEt.text.toString().trim()
                when {
                    holder.isEmpty() -> toast("请输入持卡人姓名")
                    cardNo.length < 12 || !cardNo.all { it.isDigit() } -> toast("请输入正确的银行卡号")
                    bank.isEmpty() -> toast("请输入开户银行")
                    !isValidExpiry(expiry) -> toast("请输入正确的有效期，如 08/28")
                    else -> {
                        CardWalletRepository.addBank(
                            this,
                            BankCard(holder, bank, maskCardNo(cardNo), expiry),
                        )
                        buildBankCards()
                        dialog.dismiss()
                        toast(getString(R.string.wallet_bind_success))
                    }
                }
            }
        }
        dialog.show()
    }

    private fun showUnbindDialog(index: Int, card: BankCard) {
        AlertDialog.Builder(this)
            .setTitle(R.string.wallet_unbind_title)
            .setMessage("${getString(R.string.wallet_unbind_msg)}\n${card.bank} · ${card.cardNo}")
            .setNegativeButton(R.string.wallet_cancel, null)
            .setPositiveButton(R.string.wallet_confirm) { _, _ ->
                CardWalletRepository.removeBank(this, index)
                buildBankCards()
                toast(getString(R.string.wallet_unbind_success))
            }
            .create()
            .show()
    }

    private fun isValidExpiry(s: String): Boolean {
        if (s.length != 5 || s[2] != '/') return false
        val month = s.substring(0, 2)
        return (month.toIntOrNull() ?: 0) in 1..12 && s.substring(3).all { it.isDigit() }
    }

    /** 入参卡号仅存掩码尾号；新卡输入完整卡号，仅保留首尾 + 星号 */
    private fun maskCardNo(raw: String): String {
        if (raw.length <= 4) return raw
        return when {
            raw.contains("*") -> raw
            raw.length >= 16 -> "${raw.take(4)} **** **** ${raw.takeLast(4)}"
            else -> "${raw.take(4)} **** ${raw.takeLast(4)}"
        }
    }

    companion object {
        fun start(context: Context) {
            context.startActivity(Intent(context, CardWalletActivity::class.java))
        }
    }
}
