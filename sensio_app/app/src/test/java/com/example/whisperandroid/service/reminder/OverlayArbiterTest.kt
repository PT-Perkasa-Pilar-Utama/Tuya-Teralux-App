package com.example.whisperandroid.service.reminder

import android.content.Context
import io.mockk.every
import io.mockk.mockk
import io.mockk.mockkStatic
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

/**
 * Unit tests for OverlayArbiter.
 */
class OverlayArbiterTest {

    private lateinit var context: Context
    private lateinit var arbiter: OverlayArbiter

    @Before
    fun setup() {
        context = mockk()

        // Mock Settings.canDrawOverlays
        mockkStatic("android.provider.Settings")
        every { android.provider.Settings.canDrawOverlays(any()) } returns true

        arbiter = OverlayArbiter(context)
    }

    @Test
    fun canShowReminderOverlay_permissionGranted_noActiveOverlay_returnsTrue() {
        every { android.provider.Settings.canDrawOverlays(any()) } returns true
        arbiter.checkOverlayPermission()
        arbiter.markAssistantOverlayActive(false)
        arbiter.markReminderOverlayActive(false)

        val result = arbiter.canShowReminderOverlay()

        assertTrue(result)
    }

    @Test
    fun canShowReminderOverlay_permissionMissing_returnsFalse() {
        every { android.provider.Settings.canDrawOverlays(any()) } returns false
        arbiter.checkOverlayPermission()

        val result = arbiter.canShowReminderOverlay()

        assertFalse(result)
    }

    @Test
    fun canShowReminderOverlay_assistantOverlayActive_returnsFalse() {
        arbiter.markAssistantOverlayActive(true)
        arbiter.markReminderOverlayActive(false)

        val result = arbiter.canShowReminderOverlay()

        assertFalse(result)
    }

    @Test
    fun canShowReminderOverlay_reminderOverlayAlreadyActive_returnsFalse() {
        arbiter.markAssistantOverlayActive(false)
        arbiter.markReminderOverlayActive(true)

        val result = arbiter.canShowReminderOverlay()

        assertFalse(result)
    }

    @Test
    fun canShowReminderOverlay_bothOverlaysActive_returnsFalse() {
        arbiter.markAssistantOverlayActive(true)
        arbiter.markReminderOverlayActive(true)

        val result = arbiter.canShowReminderOverlay()

        assertFalse(result)
    }

    @Test
    fun isAnyOverlayActive_assistantOverlayActive_returnsTrue() {
        arbiter.markAssistantOverlayActive(true)
        arbiter.markReminderOverlayActive(false)

        val result = arbiter.isAnyOverlayActive()

        assertTrue(result)
    }

    @Test
    fun isAnyOverlayActive_reminderOverlayActive_returnsTrue() {
        arbiter.markAssistantOverlayActive(false)
        arbiter.markReminderOverlayActive(true)

        val result = arbiter.isAnyOverlayActive()

        assertTrue(result)
    }

    @Test
    fun isAnyOverlayActive_noActiveOverlay_returnsFalse() {
        arbiter.markAssistantOverlayActive(false)
        arbiter.markReminderOverlayActive(false)

        val result = arbiter.isAnyOverlayActive()

        assertFalse(result)
    }
}
