package com.dawnerrand.store.ui.screens.profile

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.recyclerview.widget.LinearLayoutManager
import com.dawnerrand.store.databinding.FragmentProfileBinding
import com.dawnerrand.store.ui.adapter.ProfileMenuAdapter
import com.dawnerrand.store.ui.util.padTopWithStatusBar
import com.dawnerrand.store.ui.util.toast

/**
 * 商家中心：店铺信息 + 资质/结算/经营/评价/客服/设置（profile 稿）。
 */
class ProfileFragment : Fragment() {

    private var _b: FragmentProfileBinding? = null
    private val b get() = _b!!
    private val vm: ProfileViewModel by viewModels()

    private val adapter = ProfileMenuAdapter { item ->
        when (item.label) {
            "经营数据" -> requireContext().toast("经营数据：请从工作台进入")
            else -> requireContext().toast("「${item.label}」演示环境暂未开放")
        }
    }

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View {
        _b = FragmentProfileBinding.inflate(inflater, container, false)
        return b.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        b.headerBox.padTopWithStatusBar()

        b.shopName.text = vm.shop.name
        b.shopDesc.text = vm.shop.desc

        b.menuList.layoutManager = LinearLayoutManager(requireContext())
        b.menuList.adapter = adapter
        adapter.submit(vm.menus)
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _b = null
    }
}
