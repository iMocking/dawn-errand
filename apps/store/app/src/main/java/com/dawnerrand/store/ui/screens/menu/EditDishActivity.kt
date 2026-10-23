package com.dawnerrand.store.ui.screens.menu

import android.os.Bundle
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.store.databinding.ActivityEditDishBinding
import com.dawnerrand.store.ui.util.TextWatcher
import com.dawnerrand.store.ui.util.dp
import com.dawnerrand.store.ui.util.toast

/**
 * 新增/编辑菜品：图片/名称/分类/价格/库存/描述/上架（edit-dish 稿）。
 * 演示态保存后返回并提示。
 */
class EditDishActivity : AppCompatActivity() {

    private lateinit var b: ActivityEditDishBinding
    private val vm: EditDishViewModel by viewModels {
        object : androidx.lifecycle.ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : androidx.lifecycle.ViewModel> create(modelClass: Class<T>): T =
                EditDishViewModel(intent.getStringExtra("dishId")) as T
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        b = ActivityEditDishBinding.inflate(layoutInflater)
        setContentView(b.root)

        b.title.text = vm.title
        b.backBtn.setOnClickListener { finish() }

        // 编辑态：已有菜品图替换上传占位
        if (intent.getStringExtra("dishId") != null) {
            b.uploadArea.setBackgroundResource(vm.imageBgRes)
            b.uploadIcon.setImageResource(vm.imageRes)
            b.uploadIcon.setColorFilter(vm.imageColor)
            val lp = b.uploadIcon.layoutParams
            lp.width = dp(28)
            lp.height = dp(28)
            b.uploadIcon.layoutParams = lp
        }

        // 表单绑定
        b.nameInput.setText(vm.name.value)
        b.categoryInput.setText(vm.category.value)
        b.priceInput.setText(vm.price.value)
        b.originalPriceInput.setText(vm.originalPrice.value)
        b.stockInput.setText(vm.stock.value)
        b.monthlySalesInput.setText(vm.monthlySales.value)
        b.descInput.setText(vm.desc.value)
        b.onSaleSwitch.isOn = vm.onSale.value ?: true

        b.nameInput.addTextChangedListener(TextWatcher { vm.name.value = it })
        b.categoryInput.addTextChangedListener(TextWatcher { vm.category.value = it })
        b.priceInput.addTextChangedListener(TextWatcher { vm.price.value = it })
        b.originalPriceInput.addTextChangedListener(TextWatcher { vm.originalPrice.value = it })
        b.stockInput.addTextChangedListener(TextWatcher { vm.stock.value = it })
        b.descInput.addTextChangedListener(TextWatcher { vm.desc.value = it })
        b.onSaleSwitch.setOnCheckedChangeListener { vm.onSale.value = it }

        b.uploadArea.setOnClickListener { toast("演示环境：图片上传接口待接入") }
        b.saveBtn.setOnClickListener {
            if (vm.name.value.isNullOrBlank() || vm.price.value.isNullOrBlank()) {
                toast("请填写菜品名称和售价")
                return@setOnClickListener
            }
            vm.save { finish() }
        }
    }
}
