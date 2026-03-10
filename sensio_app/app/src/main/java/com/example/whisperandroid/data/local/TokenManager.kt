package com.example.whisperandroid.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit

class TokenManager(
    context: Context
) {
    private val prefs: SharedPreferences = context.getSharedPreferences(
        "sensio_prefs",
        Context.MODE_PRIVATE
    )

    companion object {
        private const val KEY_ACCESS_TOKEN = "access_token"
        private const val KEY_TERMINAL_ID = "terminal_id"
        private const val KEY_TUYA_UID = "tuya_uid"
    }

    fun saveAccessToken(token: String) {
        prefs.edit { putString(KEY_ACCESS_TOKEN, token) }
    }

    fun getAccessToken(): String? = prefs.getString(KEY_ACCESS_TOKEN, null)

    fun saveTerminalId(id: String) {
        prefs.edit { putString(KEY_TERMINAL_ID, id) }
    }

    fun getTerminalId(): String? = prefs.getString(KEY_TERMINAL_ID, null)

    fun saveTuyaUid(uid: String) {
        prefs.edit { putString(KEY_TUYA_UID, uid) }
    }

    fun getTuyaUid(): String? = prefs.getString(KEY_TUYA_UID, null)

    fun clearToken() {
        prefs.edit {
            remove(KEY_ACCESS_TOKEN)
            remove(KEY_TUYA_UID)
        }
    }
}
