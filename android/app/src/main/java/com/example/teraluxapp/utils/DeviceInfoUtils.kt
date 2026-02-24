package com.example.teraluxapp.utils

import android.content.Context
import android.net.wifi.WifiManager
import android.os.Build
import android.provider.Settings
import java.net.NetworkInterface
import java.util.Collections

object DeviceInfoUtils {

    fun getMacAddress(context: Context): String {
        // Try to get from NetworkInterface (works for some Android versions)
        try {
            val all: List<NetworkInterface> = Collections.list(NetworkInterface.getNetworkInterfaces())
            for (nif in all) {
                if (!nif.name.equals("wlan0", ignoreCase = true)) continue

                val macBytes = nif.hardwareAddress ?: return ""

                val res1 = StringBuilder()
                for (b in macBytes) {
                    res1.append(String.format("%02X:", b))
                }

                if (res1.isNotEmpty()) {
                    res1.deleteCharAt(res1.length - 1)
                }
                return res1.toString()
            }
        } catch (ex: Exception) {
            // Ignore
        }


        // Try to get from WifiManager (Legacy approach, often returns 02:00:00:00:00:00 on newer Androids)
        @Suppress("DEPRECATION")
        try {
            val wifiManager = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as WifiManager
            val wInfo = wifiManager.connectionInfo
            val macAddress = wInfo.macAddress
            if (macAddress != "02:00:00:00:00:00") {
                return macAddress
            }
        } catch (ex: Exception) {
             // Ignore
        }

        // If generic MAC is returned (Android 6.0+), or access restricted, fallback to Android ID which is a unique enough identifier for most device management needs if MAC fails.
        // However, user specifically asked for MAC. We will return a message indicating restriction if we can't find it.
        return "MAC Unavailable (Restricted > Android 10)"
    }
    
    // Optional helper if we decide to use Android ID as fallback
    fun getAndroidId(context: Context): String {
        return Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID)
    }
}
