package com.dawnerrand.client.ui.screens.splash

import android.content.Intent
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import androidx.appcompat.app.AppCompatActivity
import com.dawnerrand.client.R
import com.dawnerrand.client.data.DemoRepository
import com.dawnerrand.client.ui.screens.login.LoginActivity
import com.dawnerrand.client.ui.screens.main.MainActivity

class SplashActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_splash)
        Handler(Looper.getMainLooper()).postDelayed({
            val target = if (DemoRepository.isLoggedIn(this))
                MainActivity::class.java else LoginActivity::class.java
            startActivity(Intent(this, target))
            finish()
        }, 1400)
    }
}
