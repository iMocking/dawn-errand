package com.dawnerrand.client.ui.screens.legal

import android.os.Bundle
import android.view.View
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.ui.util.padTopWithStatusBar
import com.dawnerrand.client.ui.util.setLightStatusBars

/** 用户协议 / 隐私政策（静态文本页） */
class LegalActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_legal)
        window.decorView.padTopWithStatusBar()
        setLightStatusBars(true)
        findViewById<View>(R.id.btnBack).setOnClickListener { finish() }

        val type = intent.getStringExtra(EXTRA_TYPE) ?: TYPE_AGREEMENT
        val title = if (type == TYPE_PRIVACY) "隐私政策" else "用户协议"
        findViewById<TextView>(R.id.tvTitle).text = title
        findViewById<TextView>(R.id.tvContent).text =
            if (type == TYPE_PRIVACY) LegalText.PRIVACY else LegalText.AGREEMENT
    }

    companion object {
        const val EXTRA_TYPE = "legal_type"
        const val TYPE_AGREEMENT = "agreement"
        const val TYPE_PRIVACY = "privacy"
    }
}

/** 与产品口径一致的协议文本；正式上线前需由法务替换。 */
object LegalText {
    const val AGREEMENT = """曙光跑腿用户服务协议

更新日期：2026年10月24日
生效日期：2026年10月24日

一、服务内容
曙光跑腿（以下简称“本平台”）为乡镇用户提供外卖代买、跑腿代办、同城配送等信息撮合与履约服务。您在使用本平台服务前，请认真阅读并同意本协议。

二、账号注册与使用
1. 您应使用本人合法持有的手机号注册并登录，不得冒用他人身份。
2. 您应妥善保管账号及验证码，因您保管不善造成的损失由您自行承担。
3. 未满16周岁的用户应在监护人陪同下使用本平台服务。

三、下单与支付
1. 您提交订单前应确认商品、地址、联系方式及费用信息。
2. 订单支付成功后，平台将通知商家或骑手接单；如遇商家休息、运力不足等情况，平台有权取消订单并原路退款。
3. 跑腿订单中的商品垫付费用按实际采购金额结算，多退少补。

四、配送与验收
1. 配送时间受天气、路况、商家出餐等因素影响，页面展示为预计时间。
2. 您应在收货时当面核对商品数量与质量；如发现问题，可通过订单售后或客服入口反馈。

五、评价与信用
1. 您应对评价内容的真实性负责，不得发布违法违规、侮辱诽谤或侵犯他人隐私的信息。
2. 平台有权对恶意下单、恶意退款、虚假评价等行为采取限制账号功能等措施。

六、责任限制
因不可抗力、第三方服务商故障、您提供信息错误等非平台原因造成的损失，平台在法律允许范围内不承担责任。

七、协议变更与终止
平台可根据业务需要更新本协议，更新后将在应用内公示。若您继续使用服务，视为接受更新后的协议。

八、联系我们
如对本协议有疑问，可通过“我的-联系客服”或客服热线 400-888-0000 与我们联系。"""

    const val PRIVACY = """曙光跑腿隐私政策

更新日期：2026年10月24日

我们深知个人信息对您的重要性，并会尽全力保护您的个人信息安全。本政策说明我们如何收集、使用、存储和保护您的信息。

一、我们收集的信息
1. 账号信息：手机号、昵称、头像。
2. 收货信息：联系人姓名、手机号、收货地址。
3. 订单信息：商品/服务内容、金额、支付方式、订单状态。
4. 设备信息：设备型号、操作系统版本、网络状态，用于保障服务稳定与安全。

二、我们如何使用信息
1. 完成账号注册、登录与身份校验。
2. 向商家、骑手提供履约所必需的收货信息。
3. 处理订单、售后、投诉与客服请求。
4. 在您授权的情况下，向您发送订单状态、优惠活动等通知。

三、信息的共享
我们仅在履约必需范围内向商家、骑手共享收货信息；除法律法规要求或您明确授权外，不会向第三方出售您的个人信息。

四、信息的存储与保护
1. 您的信息存储于中华人民共和国境内。
2. 我们采用加密传输、访问控制等措施保护您的信息。
3. 我们仅在实现服务目的所需期限内保留您的信息，超出期限后将删除或匿名化处理。

五、您的权利
您有权查询、更正、删除您的个人信息，撤回授权，注销账号。可通过“我的-设置”或联系客服行使上述权利。

六、未成年人保护
我们非常重视未成年人信息保护。若您为未成年人，请在监护人指导下使用本服务。

七、联系我们
如对本政策有任何疑问、意见或建议，可通过“我的-联系客服”或客服热线 400-888-0000 与我们联系。"""
}
