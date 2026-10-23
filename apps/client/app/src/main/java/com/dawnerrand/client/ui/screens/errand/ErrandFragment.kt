package com.dawnerrand.client.ui.screens.errand

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.EditText
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.fragment.app.Fragment
import com.dawnerrand.client.R
import com.dawnerrand.client.data.Address
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.address.AddressActivity
import com.dawnerrand.client.ui.screens.paysuccess.PaySuccessActivity
import com.dawnerrand.client.ui.theme.Design
import com.dawnerrand.client.ui.util.dp
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars
import com.dawnerrand.client.ui.util.toast

/**
 * 跑腿代办（主页 Tab 内容 / 独立页共用）：
 * 4 类型 Tab + 取/收地址 + 描述 + 费用预估 + 立即下单。
 */
class ErrandFragment : Fragment() {
    private val ctx: android.content.Context get() = requireContext()

    private val types = listOf(
        Triple("帮我买", R.drawable.ic_cart, "代买商品（按清单采购）"),
        Triple("帮我送", R.drawable.ic_hexagon, "代送物品（请电话联系）"),
        Triple("帮我取", R.drawable.ic_mail, "代取物品（请电话联系）"),
        Triple("帮我办", R.drawable.ic_tab_orders, "代办事项（请电话联系）"),
    )
    private val fees = listOf(8.0, 6.0, 5.0, 12.0)
    private val feeLabels = listOf("跑腿费（5.2km）", "跑腿费（3.8km）", "跑腿费（2.6km）", "代办费（按时长）")
    private val hints = listOf(
        "1. 金龙鱼食用油5L装 ×1桶\n2. 东北大米10kg ×1袋\n3. 生抽酱油 ×1瓶",
        "请填写要送出的物品，以及收件人的姓名和电话",
        "请填写取件码、取件地址和物品信息",
        "请描述需要代办的事项，例如：代缴电费、代排队",
    )

    private val tabViews = mutableListOf<LinearLayout>()
    private var current = 0
    private var selectedDrop: Address? = null
    private var pickTitle = "镇中心供销社超市"
    private var pickSub = "联系人：刘老板 139****5678"
    private var dropTitle = ""
    private var dropSub = ""
    private var submitting = false

    private val addressPicker = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == Activity.RESULT_OK) {
            val id = result.data?.getStringExtra(AddressActivity.EXTRA_SELECTED_ID)
            val address = id?.let { DemoRepository.addressById(it) } ?: return@registerForActivityResult
            selectedDrop = address
            dropTitle = address.shortDetail
            dropSub = "${address.name} ${DemoRepository.maskPhone(address.phone)}"
            renderAddresses()
        }
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?,
    ): View {
        val root = inflater.inflate(R.layout.activity_errand, container, false)
        val activity = requireActivity() as AppCompatActivity
        val nav = root.findViewById<View>(R.id.navBar)
        if (embedded()) {
            nav.padTopWithStatusBar()
            nav.setPadding(ctx.dp(16), nav.paddingTop, ctx.dp(16), nav.paddingBottom)
            root.findViewById<View>(R.id.btnBack).visibility = View.GONE
        } else {
            activity.setLightStatusBars(true)
            activity.window.decorView.padTopWithStatusBar()
            root.findViewById<View>(R.id.btnBack).setOnClickListener { activity.finish() }
        }

        if (dropTitle.isBlank()) {
            selectedDrop = DemoRepository.defaultAddress()
            selectedDrop?.let {
                dropTitle = it.shortDetail
                dropSub = "${it.name} ${DemoRepository.maskPhone(it.phone)}"
            }
        }
        if (dropTitle.isBlank()) {
            dropTitle = "请选择收货地址"
            dropSub = "点击选择常用地址"
        }

        tabViews.clear()
        val tabRow = root.findViewById<LinearLayout>(R.id.typeTabs)
        tabRow.removeAllViews()
        types.forEachIndexed { i, (name, icon, _) ->
            tabViews.add(buildTypeTab(i, name, icon))
            tabRow.addView(tabViews[i])
        }
        renderTabs()
        renderAddresses()
        renderFee()
        root.findViewById<EditText>(R.id.etDesc).hint = hints[current]

        root.findViewById<View>(R.id.rowDrop).setOnClickListener {
            addressPicker.launch(Intent(requireContext(), AddressActivity::class.java)
                .putExtra(AddressActivity.EXTRA_PICK, true))
        }
        root.findViewById<View>(R.id.rowPick).setOnClickListener { editPickAddress() }
        root.findViewById<View>(R.id.btnSwap).setOnClickListener { swapAddresses() }
        root.findViewById<View>(R.id.btnSubmit).setOnClickListener { submit() }
        return root
    }

    override fun onHiddenChanged(hidden: Boolean) {
        super.onHiddenChanged(hidden)
        if (!hidden) {
            renderAddresses()
            renderFee()
        }
    }

    private fun embedded(): Boolean = arguments?.getBoolean(ARG_EMBEDDED, false) ?: false

    private fun buildTypeTab(index: Int, name: String, iconRes: Int): LinearLayout {
        val tab = LinearLayout(ctx).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(0, ctx.dp(64), 1f)
            isClickable = true
            setOnClickListener {
                current = index
                renderTabs()
                renderFee()
                view?.findViewById<EditText>(R.id.etDesc)?.hint = hints[current]
            }
        }
        val icon = ImageView(ctx).apply {
            setImageResource(iconRes)
            layoutParams = LinearLayout.LayoutParams(ctx.dp(24), ctx.dp(24))
        }
        val label = TextView(ctx).apply {
            text = name
            textSize = 12f
            gravity = Gravity.CENTER
            setPadding(0, ctx.dp(4), 0, 0)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT,
                LinearLayout.LayoutParams.WRAP_CONTENT,
            )
        }
        tab.addView(icon)
        tab.addView(label)
        return tab
    }

    private fun renderTabs() {
        tabViews.forEachIndexed { i, tab ->
            val icon = tab.getChildAt(0) as ImageView
            val label = tab.getChildAt(1) as TextView
            val active = i == current
            icon.setColorFilter(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            label.setTextColor(if (active) Design.PRIMARY else Design.TEXT_SECONDARY)
            label.typeface = android.graphics.Typeface.create(
                android.graphics.Typeface.DEFAULT,
                if (active) android.graphics.Typeface.BOLD else android.graphics.Typeface.NORMAL,
            )
        }
    }

    private fun renderAddresses() {
        val root = view ?: return
        root.findViewById<TextView>(R.id.tvPick).text = pickTitle
        root.findViewById<TextView>(R.id.tvPickSub).text = pickSub
        root.findViewById<TextView>(R.id.tvDrop).text = dropTitle
        root.findViewById<TextView>(R.id.tvDropSub).text = dropSub
    }

    private fun renderFee() {
        val root = view ?: return
        val fee = fees[current]
        root.findViewById<TextView>(R.id.tvFeeLabel).text = feeLabels[current]
        root.findViewById<TextView>(R.id.tvFee).text = "¥%.1f".format(fee)
        root.findViewById<TextView>(R.id.tvPrepay).text = "¥%.1f".format(fee)
        root.findViewById<TextView>(R.id.tvBottomFee).text = "¥%.1f".format(fee)
    }

    private fun editPickAddress() {
        val input = EditText(ctx).apply {
            setText(pickTitle)
            setSelection(text.length)
            hint = "取货地址"
            setPadding(ctx.dp(16), ctx.dp(8), ctx.dp(16), ctx.dp(8))
        }
        AlertDialog.Builder(ctx)
            .setTitle("修改取货地址")
            .setView(input)
            .setNegativeButton("取消", null)
            .setPositiveButton("保存") { _, _ ->
                val text = input.text.toString().trim()
                if (text.isEmpty()) {
                    ctx.toast("取货地址不能为空")
                    return@setPositiveButton
                }
                pickTitle = text
                pickSub = "联系人：刘老板 139****5678"
                renderAddresses()
            }
            .show()
    }

    private fun swapAddresses() {
        val oldTitle = pickTitle
        val oldSub = pickSub
        pickTitle = dropTitle
        pickSub = dropSub
        dropTitle = oldTitle
        dropSub = oldSub
        renderAddresses()
        ctx.toast("已交换取货与收货地址")
    }

    private fun submit() {
        if (submitting) return
        if (dropTitle.isBlank() || dropTitle == "请选择收货地址") {
            ctx.toast("请先选择收货地址")
            return
        }
        val descView = view?.findViewById<EditText>(R.id.etDesc) ?: return
        val desc = descView.text.toString().trim().ifEmpty { types[current].third }
        submitting = true
        val order = DemoRepository.createErrandOrder(
            type = types[current].first,
            description = desc,
            pickAddress = pickTitle,
            dropAddress = dropTitle,
            fee = fees[current],
        )
        startActivity(Intent(requireContext(), PaySuccessActivity::class.java)
            .putExtra("amount", "%.1f".format(fees[current]))
            .putExtra("type", "跑腿")
            .putExtra("order_no", order.orderNo)
            .putExtra("order_id", order.id))
        descView.setText("")
        submitting = false
        if (!embedded()) requireActivity().finish()
    }

    companion object {
        private const val ARG_EMBEDDED = "embedded"

        fun newInstance(embedded: Boolean): ErrandFragment = ErrandFragment().apply {
            arguments = Bundle().apply { putBoolean(ARG_EMBEDDED, embedded) }
        }
    }
}
