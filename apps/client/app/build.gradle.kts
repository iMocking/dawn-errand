plugins {
    alias(libs.plugins.android.application)
}

val clientApiBaseUrl = (project.findProperty("clientApiBaseUrl") as String?)
    ?: "http://10.0.2.2:8080"

// 高德开放平台 Key（Android 平台 SDK，绑定包名 com.dawnerrand.client + 签名 SHA1）
// 申请地址：https://lbs.amap.com/api/android-sdk/guide/create-project/getting-started
// 构建时用 -PamapKey=... 注入；未配置时地图不渲染（仅鉴权失败提示）
val amapKey = (project.findProperty("amapKey") as String?) ?: ""

android {
    namespace = "com.dawnerrand.client"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.dawnerrand.client"
        minSdk = 26
        targetSdk = 34
        versionCode = 4
        versionName = "1.2.1"
        buildConfigField("String", "API_BASE_URL", "\"$clientApiBaseUrl\"")
        buildConfigField("String", "AMAP_KEY", "\"$amapKey\"")
        manifestPlaceholders["amapKey"] = amapKey
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
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.recyclerview)
    implementation(libs.kotlinx.coroutines.android)
    // 高德地图合包：3D 地图 V11.3.100 + 定位 V11.3.101 + 搜索 V9.8.1（Maven Central）
    implementation(libs.amap.map.location.search)
}
