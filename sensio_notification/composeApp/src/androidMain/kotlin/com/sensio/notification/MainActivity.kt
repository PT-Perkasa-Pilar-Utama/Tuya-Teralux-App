package com.sensio.notification

import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Initialize app context for notification
        appContext = applicationContext

        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.TIRAMISU) {
            val permission = android.Manifest.permission.POST_NOTIFICATIONS
            if (checkSelfPermission(permission) != android.content.pm.PackageManager.PERMISSION_GRANTED) {
                requestPermissions(arrayOf(permission), 1)
                // We don't finish yet, wait for user to grant permission
                return
            }
        }

        // Start monitoring service
        val serviceIntent = Intent(this, MeetingMonitorService::class.java)
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.O) {
            startForegroundService(serviceIntent)
        } else {
            startService(serviceIntent)
        }

        // Show feedback toast
        showNotification("Sensio Notification", "Meeting monitor started")

        // Close the activity immediately to achieve "background app" feel
        finish()
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<String>,
        grantResults: IntArray,
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == 1) {
            // No matter the result, show notification (Toast will still work) and finish
            showNotification("Sensio Notification", "Hello World")
            finish()
        }
    }
}
