plugins {
    alias(libs.plugins.android.application)
}

// 后端基址：构建时用 -PstoreApiBaseUrl=https://api.example.com 注入，与客户端 `-PclientApiBaseUrl` 同口径。
// 默认值只服务模拟器本机联调；release 包对非本机地址强制 TLS（见 res/xml/network_security_config.xml）。
val storeApiBaseUrl = (project.findProperty("storeApiBaseUrl") as String?)
    ?: "http://10.0.2.2:8080"

android {
    namespace = "com.dawnerrand.store"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.dawnerrand.store"
        minSdk = 26
        targetSdk = 34
        versionCode = 1
        versionName = "1.0.0"
        buildConfigField("String", "API_BASE_URL", "\"$storeApiBaseUrl\"")
    }

    buildTypes {
        release {
            isMinifyEnabled = false
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
    implementation(libs.androidx.constraintlayout)
    implementation(libs.kotlinx.coroutines.android)
}
