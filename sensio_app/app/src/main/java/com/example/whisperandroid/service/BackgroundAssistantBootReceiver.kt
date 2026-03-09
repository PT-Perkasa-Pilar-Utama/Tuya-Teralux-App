package com.example.whisperandroid.service

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.os.Build
import android.util.Log
import com.example.whisperandroid.data.di.NetworkModule

class BackgroundAssistantBootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED || intent.action == "android.intent.action.QUICKBOOT_POWERON") {
            Log.d("SensioBoot", "Boot completed received")

            // Access store directly
            val store = NetworkModule.backgroundAssistantModeStore
            if (store.isEnabled.value) {
                Log.i("SensioBoot", "Background assistant enabled, starting service...")
                val serviceIntent = Intent(context, BackgroundAssistantService::class.java).apply {
                    action = "ACTION_START_ASSISTANT"
                }

                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                    context.startForegroundService(serviceIntent)
                } else {
                    context.startService(serviceIntent)
                }
            } else {
                Log.d("SensioBoot", "Background assistant disabled, skipping start.")
            }
        }
    }
}
