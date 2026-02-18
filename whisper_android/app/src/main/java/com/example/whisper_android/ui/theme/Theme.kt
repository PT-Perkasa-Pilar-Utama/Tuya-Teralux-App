package com.example.whisper_android.ui.theme

import android.app.Activity
import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.dynamicDarkColorScheme
import androidx.compose.material3.dynamicLightColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.SideEffect
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

private val DarkColorScheme = darkColorScheme(
    primary = Cyan600,
    onPrimary = Slate50,
    primaryContainer = Cyan900,
    onPrimaryContainer = Cyan100,
    
    secondary = Slate400,
    onSecondary = Slate950,
    secondaryContainer = Slate800,
    onSecondaryContainer = Slate100,
    
    tertiary = Cyan400,
    onTertiary = Slate950,
    tertiaryContainer = Cyan800,
    onTertiaryContainer = Cyan200,
    
    background = Slate950,
    onBackground = Slate50,
    surface = Slate900,
    onSurface = Slate50,
    surfaceVariant = Slate800,
    onSurfaceVariant = Slate200
)

private val LightColorScheme = lightColorScheme(
    primary = Cyan600,
    onPrimary = Slate50,
    primaryContainer = Cyan100,
    onPrimaryContainer = Cyan900,
    
    secondary = Slate500,
    onSecondary = Slate50,
    secondaryContainer = Slate100,
    onSecondaryContainer = Slate900,
    
    tertiary = Cyan700,
    onTertiary = Slate50,
    tertiaryContainer = Cyan200,
    onTertiaryContainer = Cyan900,
    
    background = Slate50,
    onBackground = Slate950,
    surface = Slate100,
    onSurface = Slate950,
    surfaceVariant = Slate200,
    onSurfaceVariant = Slate700
)

@Composable
fun SmartMeetingRoomWhisperDemoTheme(
    darkTheme: Boolean = false, // Always force light mode
    // Dynamic color is available on Android 12+
    dynamicColor: Boolean = false,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColorScheme
        else -> LightColorScheme
    }
    
    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            WindowCompat.setDecorFitsSystemWindows(window, false)
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = !darkTheme
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        content = content
    )
}
