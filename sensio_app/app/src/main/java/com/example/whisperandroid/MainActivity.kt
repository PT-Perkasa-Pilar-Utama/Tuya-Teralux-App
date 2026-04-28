package com.example.whisperandroid

import android.os.Build
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.ui.unit.dp
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.bootstrap.AppBootstrapViewModel
import com.example.whisperandroid.presentation.dashboard.DashboardScreen
import com.example.whisperandroid.presentation.register.RegisterScreen
import com.example.whisperandroid.ui.theme.SensioTheme
import com.example.whisperandroid.util.FeatureAvailabilityGuard

class MainActivity : ComponentActivity() {
    companion object {
        lateinit var appContext: android.content.Context
            private set
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        appContext = applicationContext

        // Ensure NetworkModule is initialized
        NetworkModule.ensureInitialized(this)

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
fun MainScreen(
    bootstrapViewModel: AppBootstrapViewModel = viewModel {
        AppBootstrapViewModel(
            NetworkModule.authenticateUseCase,
            NetworkModule.terminalRepository,
            NetworkModule.tokenManager
        )
    }
) {
    val context = LocalContext.current
    val navController = rememberNavController()

    val bootstrapState by bootstrapViewModel.uiState.collectAsState()

    LaunchedEffect(Unit) {
        bootstrapViewModel.bootstrap()
    }

    // Permission Handling
    val launcher = androidx.activity.compose.rememberLauncherForActivityResult(
        contract = androidx.activity.result.contract.ActivityResultContracts.RequestMultiplePermissions()
    ) { _ -> }

    LaunchedEffect(Unit) {
        val permissions = mutableListOf(android.Manifest.permission.RECORD_AUDIO)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            permissions.add(android.Manifest.permission.READ_MEDIA_AUDIO)
            permissions.add(android.Manifest.permission.POST_NOTIFICATIONS)
        } else {
            permissions.add(android.Manifest.permission.READ_EXTERNAL_STORAGE)
            permissions.add(android.Manifest.permission.WRITE_EXTERNAL_STORAGE)
        }
        launcher.launch(permissions.toTypedArray())
    }

    val backgroundModeEnabled by NetworkModule.backgroundAssistantModeStore.isEnabled.collectAsState()
    val coordinator = remember { NetworkModule.backgroundAssistantCoordinator }

    // Stop local listeners if background mode is enabled
    // This allows the Service to own the runtime
    LaunchedEffect(backgroundModeEnabled) {
        if (backgroundModeEnabled) {
            val currentRoute = navController.currentBackStackEntry?.destination?.route
            if (currentRoute == AppRoutes.Meeting.route || currentRoute == AppRoutes.Assistant.route) {
                navController.navigate(AppRoutes.Dashboard.route) {
                    popUpTo(AppRoutes.Dashboard.route) { inclusive = false }
                }
            }
        }
    }

    // App Navigation & Overlay
    val token = remember { NetworkModule.tokenManager.getAccessToken() }

    Box(modifier = Modifier.fillMaxSize()) {
        if (bootstrapState.isBootstrapped) {
            val startDestination = if (bootstrapState.shouldRedirectToRegister) AppRoutes.Register.route
                else if (token != null) AppRoutes.Dashboard.route
                else AppRoutes.Register.route

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
                        if (!bootstrapState.isBootstrapped) {
                            Toast.makeText(context, "Syncing devices, please wait...", Toast.LENGTH_SHORT).show()
                        } else if (FeatureAvailabilityGuard.canOpenInteractiveScreens(backgroundModeEnabled)) {
                            navController.navigate(AppRoutes.Meeting.route)
                        }
                    },
                    onNavigateToEdge = {
                        if (!bootstrapState.isBootstrapped) {
                            Toast.makeText(context, "Syncing devices, please wait...", Toast.LENGTH_SHORT).show()
                        } else if (FeatureAvailabilityGuard.canOpenInteractiveScreens(backgroundModeEnabled)) {
                            navController.navigate(AppRoutes.Assistant.route)
                        }
                    },
                    bootstrapViewModel = bootstrapViewModel
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
        } else {
            if (bootstrapState.error != null) {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = androidx.compose.ui.Alignment.CenterHorizontally
                ) {
                    Text(
                        text = bootstrapState.error ?: "Unknown error",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.error
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    androidx.compose.material3.Button(
                        onClick = { bootstrapViewModel.bootstrap(forceRetry = true) }
                    ) {
                        Text("Retry")
                    }
                }
            } else {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = androidx.compose.ui.Alignment.CenterHorizontally
                ) {
                    CircularProgressIndicator()
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        text = "Checking device...",
                        style = MaterialTheme.typography.bodyMedium
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
