package com.sensio.notification

import java.awt.Image
import java.awt.SystemTray
import java.awt.Toolkit
import java.awt.TrayIcon

actual fun showNotification(
    title: String,
    message: String,
) {
    if (!SystemTray.isSupported()) {
        println("SystemTray is not supported")
        return
    }

    val tray = SystemTray.getSystemTray()
    val image: Image = Toolkit.getDefaultToolkit().createImage("icon.png") // Placeholder or real icon
    val trayIcon = TrayIcon(image, "Sensio Notification")

    trayIcon.isImageAutoSize = true
    trayIcon.toolTip = "Sensio Notification"

    try {
        tray.add(trayIcon)
        trayIcon.displayMessage(title, message, TrayIcon.MessageType.INFO)

        // Remove from tray after some time or immediately after showing
        // For a "pop and exit" app, we can wait a bit or just exit
    } catch (e: Exception) {
        e.printStackTrace()
    }
}
