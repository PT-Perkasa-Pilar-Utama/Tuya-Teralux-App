package com.example.whisperandroid.util

object FeatureAvailabilityGuard {
    /**
     * Returns true if interactive screens like Meeting or Assistant can be opened.
     * These screens are disabled when Background Assistant Mode is active to prevent
     * mic contention and ensure hands-free UX consistency.
     */
    fun canOpenInteractiveScreens(backgroundModeEnabled: Boolean): Boolean {
        return !backgroundModeEnabled
    }
}
