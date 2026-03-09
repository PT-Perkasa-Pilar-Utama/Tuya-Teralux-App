package com.example.whisperandroid.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class BackgroundAssistantModeStore(context: Context) {
    private val prefs: SharedPreferences = context.getSharedPreferences(
        PREFS_NAME,
        Context.MODE_PRIVATE
    )

    private val _isEnabled = MutableStateFlow(prefs.getBoolean(KEY_BACKGROUND_MODE, false))
    val isEnabled: StateFlow<Boolean> = _isEnabled.asStateFlow()

    fun setEnabled(enabled: Boolean) {
        prefs.edit { putBoolean(KEY_BACKGROUND_MODE, enabled) }
        _isEnabled.value = enabled
    }

    companion object {
        private const val PREFS_NAME = "background_assistant_prefs"
        private const val KEY_BACKGROUND_MODE = "background_mode_enabled"
    }
}
