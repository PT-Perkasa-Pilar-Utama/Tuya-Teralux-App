import java.util.Properties
plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
}

android {
    namespace = "com.example.whisper_android"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.example.whisper_android"
        minSdk = 26
        targetSdk = 34
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
        vectorDrawables {
            useSupportLibrary = true
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro",
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    kotlinOptions {
        jvmTarget = "11"
    }
    buildFeatures {
        compose = true
        buildConfig = true
    }

    val localProperties = Properties()
    val localPropertiesFile = rootProject.file("local.properties")
    if (localPropertiesFile.exists()) {
        localProperties.load(localPropertiesFile.inputStream())
    }

    defaultConfig {
        val baseUrl = localProperties.getProperty("api.base_url") ?: "https://teralux.farismunir.my.id/"
        val baseHostname = baseUrl.removePrefix("https://").removePrefix("http://").substringBefore("/")

        buildConfigField("String", "MQTT_BROKER_URL", "\"${localProperties.getProperty("mqtt.broker_url") ?: ""}\"")
        buildConfigField("String", "MQTT_USERNAME", "\"${localProperties.getProperty("mqtt.username") ?: ""}\"")
        buildConfigField("String", "MQTT_PASSWORD", "\"${localProperties.getProperty("mqtt.password") ?: ""}\"")
        buildConfigField("String", "TERALUX_API_KEY", "\"${localProperties.getProperty("teralux.api_key") ?: ""}\"")
        buildConfigField("String", "BASE_URL", "\"$baseUrl\"")
        buildConfigField("String", "BASE_HOSTNAME", "\"$baseHostname\"")
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.graphics)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.compose.material3)
    implementation("androidx.compose.material:material-icons-extended:1.7.5")

    // Required for XML themes referenced in AndroidManifest
    implementation("com.google.android.material:material:1.12.0")

    // ViewModel utilities for Compose
    implementation(libs.androidx.lifecycle.viewmodel.compose)

    // Networking
    implementation(libs.retrofit)
    implementation(libs.converter.gson)
    implementation(libs.logging.interceptor)
    implementation(libs.compose.markdown)

    // Offline Speech Recognition (Vosk)
    implementation("com.alphacephei:vosk-android:0.3.75")

    // MQTT
    implementation(libs.paho.mqtt)
    implementation(libs.paho.android)
    implementation(libs.localbroadcastmanager)
    implementation("androidx.legacy:legacy-support-v4:1.0.0")

    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.compose.ui.test.junit4)
    debugImplementation(libs.androidx.compose.ui.tooling)
    debugImplementation(libs.androidx.compose.ui.test.manifest)
}
