package com.example.whisperandroid

import android.app.Application
import android.content.Intent
import android.os.Build
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.service.reminder.MeetingReminderService

class SensioApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        NetworkModule.init(this)

        // Start meeting reminder service independently from assistant mode
        startReminderService()
    }

    private fun startReminderService() {
        val serviceIntent = Intent(this, MeetingReminderService::class.java).apply {
            action = MeetingReminderService.ACTION_START_REMINDER
        }

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(serviceIntent)
        } else {
            startService(serviceIntent)
        }
    }
}
