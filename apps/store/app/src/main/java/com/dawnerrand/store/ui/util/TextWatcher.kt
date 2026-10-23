package com.dawnerrand.store.ui.util

import android.text.Editable
import android.text.TextWatcher

/** 极简 TextWatcher：只关心文本变化 */
class TextWatcher(private val onChange: (String) -> Unit) : TextWatcher {
    override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) = Unit
    override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) = Unit
    override fun afterTextChanged(s: Editable?) = onChange(s?.toString().orEmpty())
}
