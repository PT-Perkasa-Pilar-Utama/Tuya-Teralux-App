package com.example.whisperandroid

import android.content.Intent
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.dashboard.DashboardScreen
import com.example.whisperandroid.presentation.register.RegisterScreen
import com.example.whisperandroid.service.BackgroundAssistantService
import com.example.whisperandroid.ui.theme.SensioTheme
import com.example.whisperandroid.util.FeatureAvailabilityGuard
import com.example.whisperandroid.presentation.assistant.SensioWakeWordManager
import android.widget.Toast
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.compose.LocalLifecycleOwner

class MainActivity : ComponentActivity() {
    companion object {
        lateinit var appContext: android.content.Context
            private set
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        appContext = applicationContext

        if (com.example.whisperandroid.util.DeviceUtils.isTerminal()) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
        } else if (com.example.whisperandroid.util.DeviceUtils.isPhone(this)) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        }

        enableEdgeToEdge()
        setContent {
            SensioTheme {
                MainScreen()
            }
        }
    }
}

@Composable
fun MainScreen() {
    val context = LocalContext.current
    val navController = rememberNavController()
    
    // Permission Handling
    val launcher = androidx.activity.compose.rememberLauncherForActivityResult(
        contract = androidx.activity.result.contract.ActivityResultContracts.RequestMultiplePermissions()
    ) { _ -> }

    LaunchedEffect(Unit) {
        val permissions = mutableListOf(android.Manifest.permission.RECORD_AUDIO)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            permissions.add(android.Manifest.permission.READ_MEDIA_AUDIO)
        } else {
            permissions.add(android.Manifest.permission.READ_EXTERNAL_STORAGE)
            permissions.add(android.Manifest.permission.WRITE_EXTERNAL_STORAGE)
        }
        launcher.launch(permissions.toTypedArray())
    }

    val backgroundModeEnabled by NetworkModule.backgroundAssistantModeStore.isEnabled.collectAsState()
    val coordinator = remember { NetworkModule.backgroundAssistantCoordinator }
    val scope = rememberCoroutineScope()
    
    val wakeWordManager = remember {
        SensioWakeWordManager(context) {
            coordinator.onWakeDetected()
        }
    }

    DisposableEffect(Unit) {
        coordinator.start(scope)
        coordinator.onDismissed = {
            if (NetworkModule.backgroundAssistantModeStore.isEnabled.value) {
                wakeWordManager.startListening()
            }
        }
        onDispose {
            coordinator.stop()
            wakeWordManager.destroy()
        }
    }

    // Lifecycle - Start/Stop Vosk only in foreground
    val lifecycleOwner = LocalLifecycleOwner.current
    LaunchedEffect(lifecycleOwner, backgroundModeEnabled) {
        lifecycleOwner.lifecycle.addObserver(object : DefaultLifecycleObserver {
            override fun onResume(owner: LifecycleOwner) {
                if (backgroundModeEnabled && androidx.core.content.ContextCompat.checkSelfPermission(
                        context, android.Manifest.permission.RECORD_AUDIO
                    ) == android.content.pm.PackageManager.PERMISSION_GRANTED
                ) {
                    wakeWordManager.startListening()
                }
            }
            override fun onPause(owner: LifecycleOwner) {
                wakeWordManager.stopListening()
            }
        })
    }

    LaunchedEffect(backgroundModeEnabled) {
        if (backgroundModeEnabled) {
            // Stop service if it was running (transition to foreground-only)
            context.stopService(Intent(context, BackgroundAssistantService::class.java))

            val currentRoute = navController.currentBackStackEntry?.destination?.route
            if (currentRoute == AppRoutes.Meeting.route || currentRoute == AppRoutes.Assistant.route) {
                navController.navigate(AppRoutes.Dashboard.route) {
                    popUpTo(AppRoutes.Dashboard.route) { inclusive = false }
                }
            }
        } else {
            wakeWordManager.stopListening()
        }
    }

    // App Navigation & Overlay
    val token = remember { NetworkModule.tokenManager.getAccessToken() }
    val startDestination = if (token != null) AppRoutes.Dashboard.route else AppRoutes.Register.route

    Box(modifier = Modifier.fillMaxSize()) {
        NavHost(
            navController = navController,
            startDestination = startDestination
        ) {
            composable(AppRoutes.Register.route) {
                RegisterScreen(onNavigateToDashboard = {
                    navController.navigate(AppRoutes.Dashboard.route) {
                        popUpTo(AppRoutes.Register.route) { inclusive = true }
                    }
                })
            }

            composable(AppRoutes.Dashboard.route) {
                DashboardScreen(
                    onNavigateToRegister = {
                        navController.navigate(AppRoutes.Register.route) {
                            popUpTo(AppRoutes.Dashboard.route) { inclusive = true }
                        }
                    },
                    onNavigateToUpload = { },
                    onNavigateToStreaming = {
                        if (FeatureAvailabilityGuard.canOpenInteractiveScreens(backgroundModeEnabled)) {
                            navController.navigate(AppRoutes.Meeting.route)
                        }
                    },
                    onNavigateToEdge = {
                        if (FeatureAvailabilityGuard.canOpenInteractiveScreens(backgroundModeEnabled)) {
                            navController.navigate(AppRoutes.Assistant.route)
                        }
                    }
                )
            }

            composable(AppRoutes.Meeting.route) {
                LaunchedEffect(backgroundModeEnabled) {
                    if (backgroundModeEnabled) {
                        Toast.makeText(context, "Feature disabled in Background Mode", Toast.LENGTH_SHORT).show()
                        navController.navigate(AppRoutes.Dashboard.route) {
                            popUpTo(AppRoutes.Meeting.route) { inclusive = true }
                        }
                    }
                }

                if (!backgroundModeEnabled) {
                    com.example.whisperandroid.presentation.meeting.MeetingTranscriberScreen(
                        onNavigateBack = { navController.popBackStack() }
                    )
                }
            }

            composable(AppRoutes.Assistant.route) {
                LaunchedEffect(backgroundModeEnabled) {
                    if (backgroundModeEnabled) {
                        Toast.makeText(context, "Feature disabled in Background Mode", Toast.LENGTH_SHORT).show()
                        navController.navigate(AppRoutes.Dashboard.route) {
                            popUpTo(AppRoutes.Assistant.route) { inclusive = true }
                        }
                    }
                }

                if (!backgroundModeEnabled) {
                    com.example.whisperandroid.presentation.assistant.AiAssistantScreen(
                        onNavigateBack = { navController.popBackStack() }
                    )
                }
            }
        }

        // Foreground-Only Background Assistant Modal Overlay
        com.example.whisperandroid.presentation.assistant.BackgroundAssistantModalHost(
            coordinator = coordinator
        )
    }
}
