package com.example.whisperandroid

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.dashboard.DashboardScreen
import com.example.whisperandroid.presentation.register.RegisterScreen
import com.example.whisperandroid.ui.theme.SensioTheme

class MainActivity : ComponentActivity() {
    companion object {
        lateinit var appContext: android.content.Context
            private set
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        appContext = applicationContext

        // Initialize dependency injection (Manual)
        com.example.whisperandroid.data.di.NetworkModule
            .init(this)

        // Dynamic Orientation Locking
        if (com.example.whisperandroid.util.DeviceUtils
            .isTerminal()
        ) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
        } else if (com.example.whisperandroid.util.DeviceUtils
            .isPhone(this)
        ) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        } else {
            // Tablet: Allow rotation
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
        }

        androidx.core.view.WindowCompat
            .setDecorFitsSystemWindows(window, false)
        enableEdgeToEdge()
        setContent {
            SensioTheme {
                val context = androidx.compose.ui.platform.LocalContext.current
                var permissionsGranted by remember {
                    androidx.compose.runtime.mutableStateOf(false)
                }

                val launcher =
                    androidx.activity.compose.rememberLauncherForActivityResult(
                        contract =
                        androidx.activity.result.contract.ActivityResultContracts
                            .RequestMultiplePermissions()
                    ) { permissions ->
                        // Check if all permissions are granted
                        val allGranted = permissions.entries.all { it.value }
                        permissionsGranted = allGranted
                    }

                androidx.compose.runtime.LaunchedEffect(Unit) {
                    val permissionsToRequest =
                        mutableListOf(
                            android.Manifest.permission.RECORD_AUDIO
                        )

                    val isTiramisu =
                        android.os.Build.VERSION.SDK_INT >=
                            android.os.Build.VERSION_CODES.TIRAMISU
                    if (isTiramisu) {
                        permissionsToRequest.add(android.Manifest.permission.READ_MEDIA_AUDIO)
                    } else {
                        permissionsToRequest.add(android.Manifest.permission.READ_EXTERNAL_STORAGE)
                        permissionsToRequest.add(android.Manifest.permission.WRITE_EXTERNAL_STORAGE)
                    }

                    launcher.launch(permissionsToRequest.toTypedArray())
                }

                // Navigation implementation
                val navController = rememberNavController()
                val tm = com.example.whisperandroid.data.di.NetworkModule.tokenManager
                val token = tm.getAccessToken()
                val startDestination = remember {
                    if (token != null) {
                        AppRoutes.Dashboard.route
                    } else {
                        AppRoutes.Register.route
                    }
                }

                NavHost(
                    navController = navController,
                    startDestination = startDestination
                ) {
                    composable(AppRoutes.Register.route) {
                        RegisterScreen(onNavigateToDashboard = {
                            navController.navigate(AppRoutes.Dashboard.route) {
                                popUpTo(AppRoutes.Register.route) {
                                    inclusive = true
                                }
                                launchSingleTop = true
                                restoreState = true
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
                            onNavigateToUpload = { /* Deprecated */ },
                            onNavigateToStreaming = {
                                navController.navigate(AppRoutes.Meeting.route) {
                                    popUpTo(AppRoutes.Dashboard.route) { saveState = true }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            },
                            onNavigateToEdge = {
                                navController.navigate(AppRoutes.Assistant.route) {
                                    popUpTo(AppRoutes.Dashboard.route) { saveState = true }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            }
                        )
                    }

                    composable(AppRoutes.Meeting.route) {
                        com.example.whisperandroid.presentation.meeting.MeetingTranscriberScreen(
                            onNavigateBack = { navController.popBackStack() }
                        )
                    }

                    composable(AppRoutes.Assistant.route) {
                        com.example.whisperandroid.presentation.assistant.AiAssistantScreen(
                            onNavigateBack = { navController.popBackStack() }
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun Greeting(
    name: String,
    modifier: Modifier = Modifier
) {
    Text(
        text = "Hello $name!",
        style = MaterialTheme.typography.displayMedium,
        modifier = modifier
    )
}

@Preview(showBackground = true)
@Composable
fun GreetingPreview() {
    SensioTheme {
        Greeting("Android")
    }
}
