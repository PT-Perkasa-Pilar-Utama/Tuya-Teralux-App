package com.example.teraluxapp.ui

import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import androidx.compose.runtime.LaunchedEffect
import com.example.teraluxapp.utils.SessionManager
import com.example.teraluxapp.ui.devices.SmartACScreen
import com.example.teraluxapp.ui.devices.SwitchDeviceScreen
import com.example.teraluxapp.ui.devices.SensorDeviceScreen
import com.example.teraluxapp.ui.devices.DeprecatedDeviceListScreen
import com.example.teraluxapp.ui.dashboard.DashboardScreen
import com.example.teraluxapp.ui.settings.SettingsScreen
import com.example.teraluxapp.ui.teralux.StartupScreen
import com.example.teraluxapp.ui.teralux.RegisterTeraluxScreen
import com.example.teraluxapp.ui.login.LoginScreen
import java.net.URLDecoder
import java.net.URLEncoder

@Composable
fun AppNavigation() {
    val navController = rememberNavController()

    LaunchedEffect(Unit) {
        SessionManager.logoutEvent.collect {
            navController.navigate("startup") {
                popUpTo(0) { inclusive = true }
            }
        }
    }

    NavHost(navController = navController, startDestination = "startup") {
        composable("startup") {
            StartupScreen(
                onDeviceRegistered = {
                    navController.navigate("login") {
                        popUpTo("startup") { inclusive = true }
                    }
                },
                onDeviceNotRegistered = { mac ->
                    navController.navigate("register_teralux/$mac") {
                        popUpTo("startup") { inclusive = true }
                    }
                }
            )
        }
        composable(
            route = "register_teralux/{mac}",
            arguments = listOf(navArgument("mac") { type = NavType.StringType })
        ) { backStackEntry ->
            val mac = backStackEntry.arguments?.getString("mac") ?: ""
            RegisterTeraluxScreen(
                macAddress = mac,
                onRegistrationSuccess = {
                    navController.navigate("login") {
                        popUpTo("register_teralux/$mac") { inclusive = true }
                    }
                }
            )
        }
        composable("login") {
            LoginScreen(
                onLoginSuccess = { token, uid ->
                    navController.navigate("room_status?token=$token&uid=$uid") {
                        popUpTo("login") { inclusive = true }
                    }
                }
            )
        }
        composable(
            route = "room_status?token={token}&uid={uid}",
            arguments = listOf(
                navArgument("token") { type = NavType.StringType },
                navArgument("uid") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val token = backStackEntry.arguments?.getString("token") ?: ""
            val uid = backStackEntry.arguments?.getString("uid") ?: ""
            com.example.teraluxapp.ui.room.RoomStatusScreen(
                token = token,
                uid = uid,
                onNavigateToDashboard = { t, u ->
                    navController.navigate("dashboard?token=$t&uid=$u") {
                        popUpTo("room_status?token=$token&uid=$uid") { inclusive = true }
                    }
                }
            )
        }
        composable(
            route = "dashboard?token={token}&uid={uid}",
            arguments = listOf(
                navArgument("token") { type = NavType.StringType },
                navArgument("uid") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val token = backStackEntry.arguments?.getString("token") ?: ""
            val uid = backStackEntry.arguments?.getString("uid") ?: ""
            DashboardScreen(
                token = token,
                onDeviceClick = { deviceId, category, deviceName, gatewayId ->
                    val encodedName = URLEncoder.encode(deviceName, "UTF-8")
                    val safeGatewayId = gatewayId ?: ""
                    navController.navigate("device/$deviceId/$category/$encodedName?token=$token&gatewayId=$safeGatewayId")
                },
                onSettingsClick = {
                    navController.navigate("settings?token=$token")
                },
                onVoiceControlClick = {
                    navController.navigate("voice_control")
                },
                onBack = {
                    navController.navigate("room_status?token=$token&uid=$uid") {
                        popUpTo("dashboard?token=$token&uid=$uid") { inclusive = true }
                    }
                }
            )
        }
        composable(
            route = "settings?token={token}",
            arguments = listOf(
                navArgument("token") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val token = backStackEntry.arguments?.getString("token") ?: ""
            SettingsScreen(
                token = token,
                onBack = { navController.popBackStack() }
            )
        }
        composable(
            route = "devices?token={token}&uid={uid}",
            arguments = listOf(
                navArgument("token") { type = NavType.StringType },
                navArgument("uid") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val token = backStackEntry.arguments?.getString("token") ?: ""
            val uid = backStackEntry.arguments?.getString("uid") ?: ""
            DeprecatedDeviceListScreen(
                token = token,
                onDeviceClick = { deviceId, category, deviceName, gatewayId ->
                    val encodedName = URLEncoder.encode(deviceName, "UTF-8")
                    val safeGatewayId = gatewayId ?: ""
                    navController.navigate("device/$deviceId/$category/$encodedName?token=$token&gatewayId=$safeGatewayId")
                },
                onBack = {
                    navController.navigate("room_status?token=$token&uid=$uid") {
                        popUpTo("devices?token=$token&uid=$uid") { inclusive = true }
                    }
                }
            )
        }
        
        // Device Detail Route with category-based rendering
        composable(
            route = "device/{deviceId}/{category}/{name}?token={token}&gatewayId={gatewayId}",
            arguments = listOf(
                navArgument("deviceId") { type = NavType.StringType },
                navArgument("category") { type = NavType.StringType },
                navArgument("name") { type = NavType.StringType },
                navArgument("token") { type = NavType.StringType },
                navArgument("gatewayId") { type = NavType.StringType; defaultValue = "" }
            )
        ) { backStackEntry ->
            val deviceId = backStackEntry.arguments?.getString("deviceId") ?: ""
            val category = backStackEntry.arguments?.getString("category") ?: ""
            val name = URLDecoder.decode(backStackEntry.arguments?.getString("name") ?: "Device", "UTF-8")
            val gatewayId = backStackEntry.arguments?.getString("gatewayId") ?: ""
            val token = backStackEntry.arguments?.getString("token") ?: ""
            
            // DEBUG: Log the category
            android.util.Log.d("AppNavigation", "Device: $name, Category: $category, ID: $deviceId")
            
            // Route based on category (from actual Tuya API response)
            // - dgnzk = Multi-function controller (Smart Central Control Panel) -> Switch
            // - kg, cz, pc, clkg, cjkg, tdq = Various switches -> Switch
            // - infrared_ac = IR AC remote -> SmartACScreen
            // - wnykq = Universal IR remote (Smart IR) -> SmartACScreen
            // - kt, ktkzq = AC controller -> SmartACScreen
            when {
                // Switch/Multi-function categories -> SwitchDeviceScreen
                category in listOf("dgnzk", "kg", "cz", "pc", "clkg", "cjkg", "tdq", "kgq", "tgkg", "tgq", "dj", "dd", "dlq") -> {
                    SwitchDeviceScreen(
                        deviceId = deviceId,
                        deviceName = name,
                        token = token,
                        onBack = { navController.popBackStack() }
                    )
                }
                // Sensor categories
                category in listOf("wsdcg") -> {
                    SensorDeviceScreen(
                        deviceId = deviceId,
                        deviceName = name,
                        token = token,
                        onBack = { navController.popBackStack() }
                    )
                }
                // AC/IR categories -> SmartACScreen (IR AC control)
                else -> {
                    SmartACScreen(
                        deviceId = deviceId,
                        deviceName = name,
                        token = token,
                        infraredId = if (gatewayId.isNotEmpty()) gatewayId else "a36d8e212f67a0ea2dbgnl", // Use gatewayId if present, else fallback
                        onBack = { navController.popBackStack() }
                    )
                }
            }
        }
        
        // Voice Control Route
        composable("voice_control") {
             com.example.teraluxapp.ui.voice.VoiceControlScreen(navController = navController)
        }
    }
}
