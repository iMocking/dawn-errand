plugins {
    alias(libs.plugins.android.application)
}

val clientApiBaseUrl = (project.findProperty("clientApiBaseUrl") as String?)
    ?: "http://10.0.2.2:8080"

android {
    namespace = "com.dawnerrand.client"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.dawnerrand.client"
        minSdk = 26
        targetSdk = 34
        versionCode = 2
        versionName = "1.1.0"
        buildConfigField("String", "API_BASE_URL", "\"$clientApiBaseUrl\"")
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
}
