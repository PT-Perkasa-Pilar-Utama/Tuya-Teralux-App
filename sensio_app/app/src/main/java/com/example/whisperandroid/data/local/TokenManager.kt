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
        private const val KEY_MAC_ADDRESS = "mac_address"
        private const val KEY_TOKEN_EXPIRES_AT = "token_expires_at"
    }

    fun saveTokenWithExpiry(token: String, expiresAtMillis: Long) {
        prefs.edit {
            putString(KEY_ACCESS_TOKEN, token)
            putLong(KEY_TOKEN_EXPIRES_AT, expiresAtMillis)
        }
    }

    fun saveRenewedToken(token: String, expiresAtMillis: Long) {
        prefs.edit {
            putString(KEY_ACCESS_TOKEN, token)
            putLong(KEY_TOKEN_EXPIRES_AT, expiresAtMillis)
        }
    }

    fun isTokenExpired(): Boolean {
        val token = getAccessToken() ?: return true
        val expiresAt = prefs.getLong(KEY_TOKEN_EXPIRES_AT, 0L)
        return expiresAt == 0L || System.currentTimeMillis() >= expiresAt
    }

    fun saveAccessToken(token: String) {
        prefs.edit { putString(KEY_ACCESS_TOKEN, token) }
    }

    fun getAccessToken(): String? {
        return try {
            prefs.getString(KEY_ACCESS_TOKEN, null)
        } catch (e: Exception) {
            null
        }
    }

    fun saveTerminalId(id: String) {
        prefs.edit { putString(KEY_TERMINAL_ID, id) }
    }

    fun getTerminalId(): String? {
        return try {
            prefs.getString(KEY_TERMINAL_ID, null)
        } catch (e: Exception) {
            null
        }
    }

    fun saveTuyaUid(uid: String) {
        prefs.edit { putString(KEY_TUYA_UID, uid) }
    }

    fun getTuyaUid(): String? {
        return try {
            prefs.getString(KEY_TUYA_UID, null)
        } catch (e: Exception) {
            null
        }
    }

    fun saveMacAddress(mac: String) {
        prefs.edit { putString(KEY_MAC_ADDRESS, mac) }
    }

    fun getMacAddress(): String? {
        return try {
            prefs.getString(KEY_MAC_ADDRESS, null)
        } catch (e: Exception) {
            null
        }
    }

    fun clearToken() {
        prefs.edit {
            remove(KEY_ACCESS_TOKEN)
            remove(KEY_TUYA_UID)
            remove(KEY_TERMINAL_ID)
            remove(KEY_MAC_ADDRESS)
        }
    }
}
