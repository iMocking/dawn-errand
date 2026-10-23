package com.dawnerrand.client.ui.screens.orders

import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity

/** 独立订单页（深链/兼容入口）；主页 Tab 内直接使用 [OrdersFragment]。 */
class OrdersActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        if (savedInstanceState == null) {
            supportFragmentManager.beginTransaction()
                .replace(
                    android.R.id.content,
                    OrdersFragment.newInstance(
                        embedded = false,
                        filterStatus = intent.getStringExtra(EXTRA_FILTER_STATUS),
                        tab = intent.getIntExtra(EXTRA_TAB, 0),
                    ),
                )
                .commit()
        }
    }

    companion object {
        const val EXTRA_TAB = "tab_index"
        const val EXTRA_FILTER_STATUS = "filter_status"
    }
}
