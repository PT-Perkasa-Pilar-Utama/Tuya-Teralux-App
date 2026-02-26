
package com.example.whisper_android.util

import android.content.Context
import android.content.res.Configuration
import android.os.Build

object DeviceUtils {
    fun isTablet(context: Context): Boolean =
        (context.resources.configuration.screenLayout and Configuration.SCREENLAYOUT_SIZE_MASK) >=
            Configuration.SCREENLAYOUT_SIZE_LARGE

    fun isTerminal(): Boolean {
        val model = Build.MODEL
        val manufacturer = Build.MANUFACTURER
        return model.contains("px30_evb", ignoreCase = true) ||
            manufacturer.contains("rockchip", ignoreCase = true) ||
            model.contains("terminal", ignoreCase = true) ||
            model.contains("teralux", ignoreCase = true)
    }

    fun isPhone(context: Context): Boolean = !isTablet(context) && !isTerminal()

    fun getDeviceTypeId(context: Context): String {
        return when {
            isTerminal() -> "2"
            isTablet(context) -> "1"
            else -> "3" // Phone
        }
    }

    fun getDeviceId(context: Context): String {
        // 1. Check SharedPreferences FIRST (Stable across restarts)
        val prefs = context.getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
        var savedId = prefs.getString("device_id", null) ?: prefs.getString("device_uuid", null)

        if (!savedId.isNullOrBlank()) {
            return savedId.uppercase()
        }

        // 2. If first time, try ANDROID_ID
        val androidId =
            android.provider.Settings.Secure.getString(
                context.contentResolver,
                android.provider.Settings.Secure.ANDROID_ID
            )

        val newId = if (!androidId.isNullOrBlank() && androidId != "9774d56d682e549c") {
            androidId.uppercase()
        } else {
            // 3. Fallback to random UUID if ANDROID_ID is missing/broken
            java.util.UUID.randomUUID().toString()
        }

        // Save for future use
        prefs.edit().putString("device_id", newId).apply()

        return newId.uppercase()
    }
}
