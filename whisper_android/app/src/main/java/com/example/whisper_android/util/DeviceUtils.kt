
package com.example.whisper_android.util

import android.content.Context
import android.content.res.Configuration
import android.os.Build

object DeviceUtils {
    fun isTablet(context: Context): Boolean =
        (context.resources.configuration.screenLayout and Configuration.SCREENLAYOUT_SIZE_MASK) >=
            Configuration.SCREENLAYOUT_SIZE_LARGE

    fun isTeralux(): Boolean {
        val model = Build.MODEL
        val manufacturer = Build.MANUFACTURER
        return model.contains("px30_evb", ignoreCase = true) ||
            manufacturer.contains("rockchip", ignoreCase = true) ||
            model.contains("teralux", ignoreCase = true)
    }

    fun isPhone(context: Context): Boolean = !isTablet(context) && !isTeralux()

    fun getDeviceId(context: Context): String {
        // 1. Try ANDROID_ID (Stable for lifetime of device reset)
        val androidId =
            android.provider.Settings.Secure.getString(
                context.contentResolver,
                android.provider.Settings.Secure.ANDROID_ID,
            )

        if (!androidId.isNullOrBlank() && androidId != "9774d56d682e549c") { // Known broken ID on some emulators
            return androidId.uppercase()
        }

        // 2. Fallback to stored UUID (Persists until app uninstall/clear data)
        val prefs = context.getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
        var uuid = prefs.getString("device_uuid", null)

        if (uuid == null) {
            uuid =
                java.util.UUID
                    .randomUUID()
                    .toString()
            prefs.edit().putString("device_uuid", uuid).apply()
        }

        return uuid!!
    }
}
