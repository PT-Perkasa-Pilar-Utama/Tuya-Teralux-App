package com.sensio.notification

import platform.Foundation.NSUUID
import platform.UserNotifications.UNAuthorizationOptionAlert
import platform.UserNotifications.UNAuthorizationOptionSound
import platform.UserNotifications.UNMutableNotificationContent
import platform.UserNotifications.UNNotificationAction
import platform.UserNotifications.UNNotificationActionOptionForeground
import platform.UserNotifications.UNNotificationCategory
import platform.UserNotifications.UNNotificationCategoryOptionNone
import platform.UserNotifications.UNNotificationRequest
import platform.UserNotifications.UNTimeIntervalNotificationTrigger
import platform.UserNotifications.UNUserNotificationCenter

actual fun showNotification(
    title: String,
    message: String,
) {
    val center = UNUserNotificationCenter.currentNotificationCenter()

    // Define Actions
    val extendAction =
        UNNotificationAction.actionWithIdentifier(
            identifier = "EXTEND_ACTION",
            title = "Extend 10m",
            options = UNNotificationActionOptionForeground,
        )
    val dismissAction =
        UNNotificationAction.actionWithIdentifier(
            identifier = "DISMISS_ACTION",
            title = "Dismiss",
            options = UNNotificationCategoryOptionNone,
        )

    // Define Category
    val category =
        UNNotificationCategory.categoryWithIdentifier(
            identifier = "MEETING_CATEGORY",
            actions = listOf(extendAction, dismissAction),
            intentIdentifiers = emptyList<String>(),
            options = UNNotificationCategoryOptionNone,
        )

    center.setNotificationCategories(setOf(category))

    center.requestAuthorizationWithOptions(UNAuthorizationOptionAlert or UNAuthorizationOptionSound) { granted, error ->
        if (granted) {
            val content =
                UNMutableNotificationContent().apply {
                    setTitle(title)
                    setBody(message)
                    setCategoryIdentifier("MEETING_CATEGORY")
                }

            val trigger = UNTimeIntervalNotificationTrigger.triggerWithTimeInterval(1.0, false)
            val request =
                UNNotificationRequest.requestWithIdentifier(
                    NSUUID().UUIDString(),
                    content,
                    trigger,
                )

            center.addNotificationRequest(request) { error ->
                if (error != null) {
                    println("Error showing notification: ${error.localizedDescription}")
                }
            }
        }
    }
}
