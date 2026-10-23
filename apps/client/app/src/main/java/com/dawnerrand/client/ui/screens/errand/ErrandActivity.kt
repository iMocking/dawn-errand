package com.dawnerrand.client.ui.screens.errand

import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity

/** 独立跑腿页（深链/兼容入口）；主页 Tab 内直接使用 [ErrandFragment]。 */
class ErrandActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        if (savedInstanceState == null) {
            supportFragmentManager.beginTransaction()
                .replace(android.R.id.content, ErrandFragment.newInstance(embedded = false))
                .commit()
        }
    }
}
