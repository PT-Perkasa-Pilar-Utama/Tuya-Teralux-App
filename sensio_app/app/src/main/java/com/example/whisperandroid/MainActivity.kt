package com.example.whisperandroid

import android.os.Build
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
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
import com.example.whisperandroid.navigation.AuthChecker
import androidx.navigation.compose.rememberNavController
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.auth.AuthStateManager
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.authenticating.AuthenticatingScreen
import com.example.whisperandroid.presentation.authenticating.AuthenticatingViewModel
import com.example.whisperandroid.presentation.dashboard.DashboardScreen
import com.example.whisperandroid.presentation.register.RegisterScreen

import com.example.whisperandroid.ui.theme.SensioTheme
import com.example.whisperandroid.utils.FeatureAvailabilityGuard

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

        if (com.example.whisperandroid.utils.DeviceUtils.isTerminal()) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
        } else if (com.example.whisperandroid.utils.DeviceUtils.isPhone(this)) {
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
    val startDestination = AuthChecker.getStartDestination()

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

            composable(AppRoutes.Authenticating.route) {
                val viewModel: AuthenticatingViewModel = androidx.lifecycle.viewmodel.compose.viewModel(
                    factory = AuthenticatingViewModelFactory(
                        application = context.applicationContext as android.app.Application,
                        onNavigateToDashboard = {
                            navController.navigate(AppRoutes.Dashboard.route) {
                                popUpTo(AppRoutes.Authenticating.route) { inclusive = true }
                            }
                        },
                        onNavigateToRegister = {
                            navController.navigate(AppRoutes.Register.route) {
                                popUpTo(AppRoutes.Authenticating.route) { inclusive = true }
                            }
                        },
                        onAuthError = { error ->
                            Toast.makeText(context, error, Toast.LENGTH_LONG).show()
                        }
                    )
                )
                AuthenticatingScreen(
                    errorMessage = viewModel.uiState.value.errorMessage,
                    onRetry = if (viewModel.uiState.value.isRetryEnabled) {
                        { viewModel.retry() }
                    } else null,
                    onAuthError = { error ->
                        Toast.makeText(context, error, Toast.LENGTH_LONG).show()
                    },
                    onNavigateToDashboard = {
                        navController.navigate(AppRoutes.Dashboard.route) {
                            popUpTo(AppRoutes.Authenticating.route) { inclusive = true }
                        }
                    },
                    onNavigateToRegister = {
                        navController.navigate(AppRoutes.Register.route) {
                            popUpTo(AppRoutes.Authenticating.route) { inclusive = true }
                        }
                    }
                )
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
                    },
                    onNavigateToAuth = {
                        navController.navigate(AppRoutes.Authenticating.route) {
                            popUpTo(AppRoutes.Dashboard.route) { inclusive = true }
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

class AuthenticatingViewModelFactory(
    private val application: android.app.Application,
    private val onNavigateToDashboard: () -> Unit,
    private val onNavigateToRegister: () -> Unit,
    private val onAuthError: (String) -> Unit
) : androidx.lifecycle.ViewModelProvider.Factory {
    @Suppress("UNCHECKED_CAST")
    override fun <T : androidx.lifecycle.ViewModel> create(modelClass: Class<T>): T {
        if (modelClass.isAssignableFrom(AuthenticatingViewModel::class.java)) {
            return AuthenticatingViewModel(
                application = application,
                onNavigateToDashboard = onNavigateToDashboard,
                onNavigateToRegister = onNavigateToRegister,
                onAuthError = onAuthError
            ) as T
        }
        throw IllegalArgumentException("Unknown ViewModel class")
    }
}
