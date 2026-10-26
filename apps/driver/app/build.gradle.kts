import java.util.Properties

plugins {
    alias(libs.plugins.android.application)
}

// 发布签名（keystore/keystore.properties，已 gitignore）：高德 API Key 安全码即该 keystore 的 SHA1
val keystoreProps = Properties().apply {
    val f = rootProject.file("keystore/keystore.properties")
    if (f.exists()) f.inputStream().use { load(it) }
}

// 后端基址：构建时用 -PdriverApiBaseUrl=https://api.example.com 注入，与客户端 `-PclientApiBaseUrl` 同口径。
// 默认值只服务模拟器本机联调；release 包对非本机地址强制 TLS（见 res/xml/network_security_config.xml）。
val driverApiBaseUrl = (project.findProperty("driverApiBaseUrl") as String?)
    ?: "http://10.0.2.2:8080"

android {
    namespace = "com.dawnerrand.driver"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.dawnerrand.driver"
        minSdk = 26
        targetSdk = 34
        versionCode = 1
        versionName = "1.0.0"
        buildConfigField("String", "API_BASE_URL", "\"$driverApiBaseUrl\"")
    }

    signingConfigs {
        create("release") {
            if (keystoreProps.isNotEmpty()) {
                storeFile = rootProject.file(keystoreProps.getProperty("storeFile"))
                storePassword = keystoreProps.getProperty("storePassword")
                keyAlias = keystoreProps.getProperty("keyAlias")
                keyPassword = keystoreProps.getProperty("keyPassword")
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            signingConfig = signingConfigs.getByName("release")
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    buildFeatures {
        buildConfig = true
        viewBinding = true
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.appcompat)
    implementation(libs.androidx.activity.ktx)
    implementation(libs.androidx.fragment.ktx)
    implementation(libs.androidx.lifecycle.viewmodel.ktx)
    implementation(libs.androidx.lifecycle.livedata.ktx)
    implementation(libs.androidx.recyclerview)
    implementation(libs.kotlinx.coroutines.android)
    implementation(libs.amap.map)
}
