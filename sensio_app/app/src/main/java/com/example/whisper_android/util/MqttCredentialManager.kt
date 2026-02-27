package com.example.whisper_android.util

import android.content.Context
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

class MqttCredentialManager(context: Context) {
    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val sharedPreferences = EncryptedSharedPreferences.create(
        context,
        "mqtt_creds",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    fun saveCredentials(username: String, password: String) {
        sharedPreferences.edit().apply {
            putString("mqtt_username", username)
            putString("mqtt_password", password)
            apply()
        }
    }

    fun getUsername(): String? = sharedPreferences.getString("mqtt_username", null)
    fun getPassword(): String? = sharedPreferences.getString("mqtt_password", null)

    fun clearCredentials() {
        sharedPreferences.edit().clear().apply()
    }
}
