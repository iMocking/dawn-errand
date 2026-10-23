package com.dawnerrand.client.ui.screens.orderconfirm

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.dawnerrand.client.R
import com.dawnerrand.client.data.Address
import com.dawnerrand.client.data.Coupon
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.data.OrderLine
import com.dawnerrand.client.ui.screens.address.AddressActivity
import com.dawnerrand.client.ui.screens.coupons.CouponsActivity
import com.dawnerrand.client.ui.screens.paysuccess.PaySuccessActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padBottomWithNavBar
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/** 确认订单：地址选择、商品明细、优惠券、备注、支付方式、提交支付。 */
class OrderConfirmActivity : AppCompatActivity() {
    private val ctx: android.content.Context get() = this

    private var merchantId = "m1"
    private var selectedAddress: Address? = null
    private var selectedCoupon: Coupon? = null
    private var payMethod = "微信支付"
    private var remark = ""
    private var submitting = false

    private val addressPicker = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == Activity.RESULT_OK) {
            val id = result.data?.getStringExtra(AddressActivity.EXTRA_SELECTED_ID)
            val address = id?.let { DemoRepository.addressById(it) }
            if (address != null) {
                selectedAddress = address
                renderAddress()
            }
        }
    }

    private val couponPicker = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == Activity.RESULT_OK) {
            val id = result.data?.getStringExtra(CouponsActivity.EXTRA_SELECTED_COUPON_ID)
            selectedCoupon = id?.let { DemoRepository.couponById(it) }
            renderTotals()
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_order_confirm)
        window.decorView.padTopWithStatusBar()
        window.decorView.padBottomWithNavBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        merchantId = intent.getStringExtra("merchant_id") ?: "m1"
        selectedAddress = DemoRepository.defaultAddress()
        selectedCoupon = null

        findViewById<View>(R.id.cardAddress).setOnClickListener {
            addressPicker.launch(Intent(this, AddressActivity::class.java)
                .putExtra(AddressActivity.EXTRA_PICK, true))
        }
        findViewById<View>(R.id.rowPay).setOnClickListener { choosePayMethod() }
        findViewById<View>(R.id.rowCoupon).setOnClickListener {
            if (DemoRepository.availableCoupons().isEmpty()) {
                ctx.toast("暂无可用优惠券")
                return@setOnClickListener
            }
            couponPicker.launch(Intent(this, CouponsActivity::class.java)
                .putExtra(CouponsActivity.EXTRA_PICK, true))
        }
        findViewById<View>(R.id.rowRemark).setOnClickListener { editRemark() }
        findViewById<View>(R.id.btnSubmit).setOnClickListener { submit() }

        renderAddress()
        renderGoods()
        renderTotals()
    }

    private fun renderAddress() {
        val a = selectedAddress
        val name = findViewById<TextView>(R.id.tvAddressName)
        val detail = findViewById<TextView>(R.id.tvAddressDetail)
        if (a == null) {
            name.text = "请选择收货地址"
            detail.text = "点击新增或选择收货地址"
        } else {
            name.text = "${a.name}  ${DemoRepository.maskPhone(a.phone)}"
            detail.text = a.shortDetail
        }
    }

    private fun renderGoods() {
        val goods = findViewById<LinearLayout>(R.id.cardGoods)
        goods.removeAllViews()
        val merchant = DemoRepository.merchantById(merchantId)
        goods.addView(TextView(this).apply {
            text = merchant.name
            setTextColor(Design.TEXT_PRIMARY)
            textSize = 15f
            typeface = android.graphics.Typeface.DEFAULT_BOLD
            setPadding(0, 0, 0, ctx.dp(10))
        })
        if (DemoRepository.cart.isEmpty() || DemoRepository.cartMerchantId != merchantId) {
            goods.addView(TextView(this).apply {
                text = "购物车为空，请返回商家选择商品"
                setTextColor(Design.TEXT_FAINT)
                textSize = 13f
                setPadding(0, ctx.dp(10), 0, ctx.dp(10))
            })
            return
        }
        DemoRepository.cart.forEach { item ->
            goods.addView(row(
                "${item.dish.name} ×${item.count}",
                "¥%.1f".format(item.dish.price * item.count),
            ))
        }
        goods.addView(divider())
        val subtotal = DemoRepository.cartTotal()
        goods.addView(row("商品小计", "¥%.1f".format(subtotal)))
        goods.addView(row("配送费", "¥3.0"))
        goods.addView(row("打包费", "¥1.5"))
        goods.addView(row("新用户立减", "-¥3.0", Design.PRIMARY))
        selectedCoupon?.let { coupon ->
            if (DemoRepository.couponUsableFor(coupon, subtotal)) {
                goods.addView(row("优惠券（${coupon.title}）", "-¥%.1f".format(coupon.value), Design.PRIMARY))
            }
        }
    }

    private fun renderTotals() {
        val subtotal = DemoRepository.cartTotal()
        val couponValue = selectedCoupon?.takeIf { DemoRepository.couponUsableFor(it, subtotal) }?.value ?: 0.0
        val total = (subtotal + 3.0 + 1.5 - 3.0 - couponValue).coerceAtLeast(0.0)
        findViewById<TextView>(R.id.tvTotal).text = "合计 ¥%.1f".format(total)
        findViewById<TextView>(R.id.tvPay).text = payMethod
        findViewById<TextView>(R.id.tvRemark).text = remark.ifBlank { "选填" }
        val couponText = findViewById<TextView>(R.id.tvCoupon)
        val usable = selectedCoupon?.takeIf { DemoRepository.couponUsableFor(it, subtotal) }
        if (usable != null) {
            couponText.text = "已减 ¥%.1f".format(usable.value)
            couponText.setTextColor(Design.PRIMARY)
        } else {
            couponText.text = if (DemoRepository.availableCoupons().isEmpty()) {
                "暂无可用"
            } else {
                "选择优惠券（${DemoRepository.availableCoupons().size}张可用）"
            }
            couponText.setTextColor(Design.TEXT_TERTIARY)
        }
        renderGoods()
    }

    private fun choosePayMethod() {
        val methods = arrayOf("微信支付", "支付宝", "余额支付")
        val checked = methods.indexOf(payMethod).coerceAtLeast(0)
        AlertDialog.Builder(this)
            .setTitle("选择支付方式")
            .setSingleChoiceItems(methods, checked) { dialog, which ->
                payMethod = methods[which]
                findViewById<TextView>(R.id.tvPay).text = payMethod
                dialog.dismiss()
            }
            .setNegativeButton("取消", null)
            .show()
    }

    private fun editRemark() {
        val input = EditText(this).apply {
            hint = "口味、送达要求等（50字以内）"
            maxLines = 3
            setText(remark)
            setSelection(text.length)
            setPadding(ctx.dp(16), ctx.dp(8), ctx.dp(16), ctx.dp(8))
        }
        AlertDialog.Builder(this)
            .setTitle("订单备注")
            .setView(input)
            .setNegativeButton("取消", null)
            .setPositiveButton("保存") { _, _ ->
                remark = input.text.toString().trim().take(50)
                findViewById<TextView>(R.id.tvRemark).text = remark.ifBlank { "选填" }
            }
            .show()
    }

    private fun submit() {
        if (submitting) return
        if (DemoRepository.cart.isEmpty() || DemoRepository.cartMerchantId != merchantId) {
            ctx.toast("购物车为空，请先选择商品")
            return
        }
        val address = selectedAddress
        if (address == null) {
            ctx.toast("请先选择收货地址")
            return
        }
        val subtotal = DemoRepository.cartTotal()
        val coupon = selectedCoupon?.takeIf { DemoRepository.couponUsableFor(it, subtotal) }
        val lines = DemoRepository.cart.map {
            OrderLine(it.dish.name, it.dish.price, it.count)
        }
        submitting = true
        findViewById<TextView>(R.id.btnSubmit).isEnabled = false
        findViewById<TextView>(R.id.btnSubmit).text = "提交中…"
        val merchant = DemoRepository.merchantById(merchantId)
        val order = DemoRepository.createFoodOrder(
            merchant = merchant,
            address = "${address.name} ${DemoRepository.maskPhone(address.phone)} · ${address.shortDetail}",
            lines = lines,
            subtotal = subtotal,
            deliveryFee = 3.0,
            packingFee = 1.5,
            discount = 3.0,
            coupon = coupon,
            remark = remark,
            payMethod = payMethod,
        )
        startActivity(Intent(this, PaySuccessActivity::class.java)
            .putExtra("amount", order.total.removePrefix("¥"))
            .putExtra("type", "外卖")
            .putExtra("order_no", order.orderNo)
            .putExtra("order_id", order.id))
        finish()
    }

    private fun row(label: String, value: String, valueColor: Int = Design.TEXT_PRIMARY): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, ctx.dp(6), 0, ctx.dp(6))
            addView(TextView(this@OrderConfirmActivity).apply {
                text = label
                setTextColor(Design.TEXT_SECONDARY)
                textSize = 13f
                layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            })
            addView(TextView(this@OrderConfirmActivity).apply {
                text = value
                setTextColor(valueColor)
                textSize = 13f
            })
        }
    }

    private fun divider(): View = View(this).apply {
        setBackgroundColor(Design.DIVIDER)
        layoutParams = LinearLayout.LayoutParams(LinearLayout.LayoutParams.MATCH_PARENT, 1)
    }
}
