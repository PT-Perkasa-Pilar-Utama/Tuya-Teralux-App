package com.example.whisper_android.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit

class TokenManager(
    context: Context
) {
    private val prefs: SharedPreferences = context.getSharedPreferences(
        "teralux_prefs",
        Context.MODE_PRIVATE
    )

    companion object {
        private const val KEY_ACCESS_TOKEN = "access_token"
    }

    fun saveAccessToken(token: String) {
        prefs.edit { putString(KEY_ACCESS_TOKEN, token) }
    }

    fun getAccessToken(): String? = prefs.getString(KEY_ACCESS_TOKEN, null)

    fun clearToken() {
        prefs.edit { remove(KEY_ACCESS_TOKEN) }
    }
}
