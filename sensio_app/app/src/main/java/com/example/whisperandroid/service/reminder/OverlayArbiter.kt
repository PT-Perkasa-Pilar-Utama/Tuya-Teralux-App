package com.example.whisperandroid.service.reminder

import android.content.Context
import android.util.Log

/**
 * Arbiter for managing overlay conflicts between assistant and reminder overlays.
 *
 * Ensures only one overlay is visible at a time.
 */
class OverlayArbiter(private val context: Context) {
    private var isAssistantOverlayActive = false
    private var isReminderOverlayActive = false
    private var hasOverlayPermission = false
    private val tag = "OverlayArbiter"

    init {
        checkOverlayPermission()
    }

    /**
     * Check and update overlay permission state.
     */
    fun checkOverlayPermission() {
        hasOverlayPermission = android.provider.Settings.canDrawOverlays(context)
        Log.d(tag, "Overlay permission: $hasOverlayPermission")
    }

    /**
     * Mark whether the assistant overlay is currently active.
     *
     * @param active true if assistant overlay is showing
     */
    fun markAssistantOverlayActive(active: Boolean) {
        isAssistantOverlayActive = active
        Log.d(tag, "Assistant overlay active: $active")
    }

    /**
     * Mark whether a reminder overlay is currently active.
     *
     * @param active true if reminder overlay is showing
     */
    fun markReminderOverlayActive(active: Boolean) {
        isReminderOverlayActive = active
        Log.d(tag, "Reminder overlay active: $active")
    }

    /**
     * Check if reminder overlay can be shown.
     *
     * @return true if overlay permission exists and no other overlay is active
     */
    fun canShowReminderOverlay(): Boolean {
        if (!hasOverlayPermission) {
            Log.w(tag, "Cannot show reminder overlay: missing permission")
            return false
        }

        if (isAssistantOverlayActive) {
            Log.w(tag, "Cannot show reminder overlay: assistant overlay active")
            return false
        }

        if (isReminderOverlayActive) {
            Log.w(tag, "Cannot show reminder overlay: another reminder overlay already active")
            return false
        }

        Log.d(tag, "Reminder overlay allowed")
        return true
    }

    /**
     * Check if any overlay is currently active.
     */
    fun isAnyOverlayActive(): Boolean {
        return isAssistantOverlayActive || isReminderOverlayActive
    }
}
