package com.dawnerrand.driver.ui.screens.orderdetail

import android.app.AlertDialog
import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.net.Uri
import android.os.Bundle
import android.text.InputFilter
import android.text.InputType
import android.widget.EditText
import android.widget.LinearLayout
import androidx.activity.SystemBarStyle
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.dawnerrand.driver.data.repository.DriverRemoteRepository
import com.dawnerrand.driver.databinding.ActivityOrderDetailBinding
import com.dawnerrand.driver.ui.util.driverMessage
import com.dawnerrand.driver.ui.util.padBottomWithNavBar
import com.dawnerrand.driver.ui.util.padTopWithStatusBar
import com.dawnerrand.driver.ui.util.toast
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/**
 * 配送详情：取送步骤卡 + 订单信息 + 本单收入 + 底部操作栏。
 *
 * 这里不再使用演示数据：进页按 orderId 拉 `GET /api/app/driver/orders/:id`，
 * 「确认取货 / 确认送达」分别调用后端接口，失败原样提示后端 message。
 * 之前这里是一句 `toast("已确认送达，本单收入 ¥7.0 已到账")`——金额写死、后端无记录，
 * 骑手以为钱已经到账。现在可执行动作一律以后端下发的 `actions` 为准，没有动作就不给按钮。
 */
class OrderDetailActivity : AppCompatActivity() {

    private lateinit var binding: ActivityOrderDetailBinding

    private var orderId: String = ""
    private var detail: JSONObject? = null
    private var actionLabel: String = "确认送达"
    private var actionEnabled: Boolean = false
    private var busy: Boolean = false

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT),
        )
        super.onCreate(savedInstanceState)
        binding = ActivityOrderDetailBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.navBar.padTopWithStatusBar()
        binding.bottomBar.padBottomWithNavBar(extra = 0)
        binding.btnBack.setOnClickListener { finish() }

        orderId = intent.getStringExtra(EXTRA_ORDER_ID).orEmpty()
        if (orderId.isBlank()) {
            // 没有订单 ID 就没有可推进的状态：宁可回退并提示，也不再演一次假成功。
            toast("订单信息缺失，请刷新后重试")
            finish()
            return
        }

        binding.btnConfirm.text = "加载中…"
        binding.btnConfirm.isEnabled = false
        binding.btnConfirm.setOnClickListener { onPrimaryAction() }
        load()
    }

    private fun load() {
        setBusy(true)
        lifecycleScope.launch {
            val json = withContext(Dispatchers.IO) { DriverRemoteRepository.orderDetail(orderId) }
            if (isFinishing || isDestroyed) return@launch
            setBusy(false)
            if (json == null) {
                toast("订单加载失败，请检查网络后重试")
                binding.btnConfirm.text = "加载失败，点击重试"
                binding.btnConfirm.isEnabled = true
                binding.btnConfirm.setOnClickListener { load() }
                return@launch
            }
            detail = json
            binding.btnConfirm.setOnClickListener { onPrimaryAction() }
            render(json)
        }
    }

    private fun render(d: JSONObject) {
        val pickup = d.optJSONObject("pickup") ?: JSONObject()
        val dropoff = d.optJSONObject("dropoff") ?: JSONObject()

        binding.tvPickName.text = pickup.optString("name").ifBlank { "商家" }
        binding.tvPickAddr.text = pickup.optString("address").ifBlank { "地址未提供" }
        val pickPhone = pickup.optString("phone")
        binding.tvPickContact.text = pickPhone.ifBlank { "电话未提供" }
        binding.btnCallPick.setOnClickListener { dial(pickPhone) }

        binding.tvDropName.text = dropoff.optString("name").ifBlank { "顾客" }
        binding.tvDropAddr.text = dropoff.optString("address").ifBlank { "地址未提供" }
        val dropPhone = dropoff.optString("phone")
        binding.tvDropContact.text = dropPhone.ifBlank { "电话未提供" }
        binding.btnCallDrop.setOnClickListener { dial(dropPhone) }
        binding.btnContact.setOnClickListener { dial(dropPhone) }

        binding.tvOrderNo.text = d.optString("order_no").ifBlank { orderId }
        binding.tvGoods.text = goodsSummary(d.optJSONArray("items"))
        binding.tvPickDist.text = "—"
        val distance = d.optDouble("distance_km", 0.0)
        binding.tvDeliverDist.text = if (distance > 0) km(distance) else "—"
        val eta = d.optLong("estimated_arrival", 0L)
        binding.tvEta.text = if (eta > 0) clock(eta) + "前" else "—"

        val reward = d.optDouble("reward", 0.0)
        val earning = d.optDouble("rider_earning", 0.0)
        binding.tvFeeIncome.text = money(reward)
        binding.tvSubsidy.text = "—"
        // 未送达前 rider_earning 为 0，此时展示本单配送费作为预期；送达后以后端实际入账为准。
        binding.tvTotalIncome.text = money(if (earning > 0) earning else reward)

        renderAction(d)
    }

    private fun renderAction(d: JSONObject) {
        val actions = actionsOf(d)
        actionLabel = when {
            "deliver" in actions -> "确认送达"
            "pickup" in actions -> "确认取货"
            "grab" in actions -> "请先在接单大厅抢单"
            else -> "当前状态不可操作"
        }
        actionEnabled = "deliver" in actions || "pickup" in actions
        setBusy(false)
    }

    private fun onPrimaryAction() {
        val d = detail ?: return
        val actions = actionsOf(d)
        val isDeliver = "deliver" in actions
        val isPickup = "pickup" in actions
        if (!isDeliver && !isPickup) {
            toast("当前没有可执行的操作")
            return
        }
        // 订单没生成取送码时后端放行空码（历史/种子订单），此时不弹输入框。
        val codeBlank = d.optString(if (isDeliver) "delivery_code" else "pickup_code").isBlank()
        if (codeBlank) {
            submit(isDeliver, "")
            return
        }
        askCode(
            title = if (isDeliver) "输入送达码" else "输入取货码",
            hint = if (isDeliver) "请向顾客索取 4 位送达码" else "请向商家索取 4 位取货码",
        ) { code -> submit(isDeliver, code) }
    }

    private fun submit(isDeliver: Boolean, code: String) {
        if (busy) return
        setBusy(true)
        lifecycleScope.launch {
            val resp = withContext(Dispatchers.IO) {
                if (isDeliver) {
                    DriverRemoteRepository.deliver(orderId, code)
                } else {
                    DriverRemoteRepository.pickup(orderId, code)
                }
            }
            if (isFinishing || isDestroyed) return@launch
            setBusy(false)
            if (!resp.ok) {
                toast(resp.driverMessage())
                // 状态被并发改掉、或校验码输错：重拉一次，界面不要停在过期状态上。
                if (resp.code == 1607 || resp.code == 1804) load()
                return@launch
            }
            if (isDeliver) {
                val earning = resp.body?.optDouble("rider_earning", 0.0) ?: 0.0
                val amount = if (earning > 0) money(earning) else "以实际入账为准"
                toast("已确认送达，本单收入 $amount")
                finish()
            } else {
                toast("已确认取货，请尽快送达")
                load()
            }
        }
    }

    private fun setBusy(value: Boolean) {
        busy = value
        binding.btnConfirm.isEnabled = actionEnabled && !value
        binding.btnConfirm.alpha = if (binding.btnConfirm.isEnabled) 1f else 0.5f
        binding.btnConfirm.text = if (value) "处理中…" else actionLabel
    }

    private fun askCode(title: String, hint: String, onConfirm: (String) -> Unit) {
        val input = EditText(this).apply {
            inputType = InputType.TYPE_CLASS_NUMBER
            filters = arrayOf(InputFilter.LengthFilter(4))
            this.hint = hint
        }
        val wrapper = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(16), dp(4), dp(16), 0)
            addView(input)
        }
        AlertDialog.Builder(this)
            .setTitle(title)
            .setView(wrapper)
            .setNegativeButton("取消", null)
            .setPositiveButton("确定") { _, _ ->
                val code = input.text.toString().trim()
                if (code.length != 4) {
                    toast("请输入 4 位数字验证码")
                    return@setPositiveButton
                }
                onConfirm(code)
            }
            .show()
    }

    private fun dial(phone: String) {
        val digits = phone.filter { it.isDigit() }
        if (digits.length < 5) {
            toast("暂无联系电话")
            return
        }
        runCatching { startActivity(Intent(Intent.ACTION_DIAL, Uri.parse("tel:$digits"))) }
            .onFailure { toast("无法打开拨号盘") }
    }

    private fun actionsOf(d: JSONObject): List<String> {
        val arr = d.optJSONArray("actions") ?: return emptyList()
        return (0 until arr.length()).mapNotNull { i -> arr.optString(i).takeIf { it.isNotBlank() } }
    }

    private fun goodsSummary(items: JSONArray?): String {
        if (items == null || items.length() == 0) return "—"
        val parts = (0 until minOf(items.length(), 2)).map { i ->
            val o = items.optJSONObject(i) ?: JSONObject()
            o.optString("name") + "×" + o.optInt("count", 1)
        }
        val extra = if (items.length() > 2) " 等${items.length()}件" else ""
        return parts.joinToString("、") + extra
    }

    private fun dp(v: Int): Int = (v * resources.displayMetrics.density).toInt()

    private fun money(v: Double): String = "¥" + String.format(Locale.CHINA, "%.2f", v)

    private fun km(v: Double): String = String.format(Locale.CHINA, "%.1fkm", v)

    private fun clock(epochSec: Long): String =
        SimpleDateFormat("HH:mm", Locale.CHINA).format(Date(epochSec * 1000))

    companion object {
        private const val EXTRA_ORDER_ID = "order_id"

        /** 进入详情必须带后端订单 ID：没有 ID 就无法真正推进状态。 */
        fun start(context: Context, orderId: String) {
            context.startActivity(
                Intent(context, OrderDetailActivity::class.java).putExtra(EXTRA_ORDER_ID, orderId),
            )
        }
    }
}
