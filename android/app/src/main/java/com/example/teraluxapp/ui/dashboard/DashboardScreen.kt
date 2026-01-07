package com.example.teraluxapp.ui.dashboard

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.example.teraluxapp.data.model.Device
import com.example.teraluxapp.data.model.TuyaSyncDeviceDTO
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

import androidx.compose.ui.graphics.vector.rememberVectorPainter
import coil.compose.AsyncImage
import androidx.compose.ui.graphics.ColorMatrix
import androidx.compose.ui.graphics.ColorFilter
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.platform.LocalContext
import com.example.teraluxapp.utils.PreferencesManager

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardScreen(token: String,
                    onDeviceClick: (deviceId: String, category: String, deviceName: String, gatewayId: String?) -> Unit,
                    onSettingsClick: () -> Unit,
                    onBack: () -> Unit = {}) {
    val context = LocalContext.current
    val teraluxId = remember { PreferencesManager.getTeraluxId(context) ?: "" }
    val scope = rememberCoroutineScope()
    var devices by remember { mutableStateOf<List<Device>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }
    val snackbarHostState = remember { SnackbarHostState() }

    val fetchDevices = {
        scope.launch {
            isLoading = true
            error = null
            try {
                val response = RetrofitClient.instance.getDevicesByTeraluxId("Bearer $token", teraluxId)
                if (response.isSuccessful && response.body() != null) {
                    val rawDevices = response.body()!!.data?.devices ?: emptyList()
                    
                    val flatList = rawDevices.flatMap { d ->
                        val parsedCollections = d.getParsedCollections()
                        if (parsedCollections.isEmpty()) listOf(d) else parsedCollections
                    }
                    // Force default online status to true (Requested: Treat all devices as online)
                    devices = flatList.map { it.copy(online = true) }
                } else {
                    error = "Failed: ${response.code()}"
                }
            } catch (e: Exception) {
                error = "Error: ${e.message}"
                e.printStackTrace()
            } finally {
                isLoading = false
            }
        }
    }

    LaunchedEffect(Unit) {
        fetchDevices()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Dashboard") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    // Settings icon
                    IconButton(onClick = onSettingsClick) {
                        Icon(Icons.Default.Settings, contentDescription = "Settings")
                    }
                    // Refresh button
                    IconButton(onClick = {
                        scope.launch {
                            isLoading = true
                            try {
                                // Sync status and get real-time data
                                val syncResponse = RetrofitClient.instance.syncDevices("Bearer $token")
                                if (syncResponse.isSuccessful && syncResponse.body()?.data != null) {
                                    // Sync logic preserved but Status update removed as requested.
                                    // Devices remain Online=true always.
                                    devices = devices.map { it.copy(online = true) }
                                } else {
                                    error = "Sync failed: ${syncResponse.code()}"
                                }
                            } catch (e: Exception) {
                                error = "Sync failed: ${e.message}"
                            } finally {
                                isLoading = false
                            }
                        }
                    }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { paddingValues ->
        Column(modifier = Modifier.padding(paddingValues).fillMaxSize()) {
            if (isLoading) {
                Box(Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            } else if (error != null) {
                Box(Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(text = error!!, color = MaterialTheme.colorScheme.error)
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(onClick = { fetchDevices() }) {
                            Text("Retry")
                        }
                    }
                }
            } else {
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .padding(8.dp), // Reduced padding
                    verticalArrangement = Arrangement.spacedBy(8.dp) // Reduced spacing
                ) {
                    val firstRowDevices = devices.take(3)
                    val secondRowDevices = if (devices.size > 3) devices.drop(3).take(3) else emptyList()

                    // Row 1
                    Row(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp) // Reduced spacing
                    ) {
                        for (i in 0 until 3) {
                            if (i < firstRowDevices.size) {
                                val device = firstRowDevices[i]
                                // For IR devices, use remote_id as deviceId and id as gatewayId
                                val hasRemoteId = !device.remoteId.isNullOrBlank()
                                val actualDeviceId = if (hasRemoteId) device.remoteId!! else device.id
                                val actualGatewayId = if (hasRemoteId) device.id else device.gatewayId
                                val rawCategory = if (!device.remoteCategory.isNullOrBlank()) device.remoteCategory else device.category ?: "unknown"
                                val actualCategory = if (rawCategory.isBlank()) "unknown" else rawCategory
                                DeviceItem(
                                    device = device,
                                    modifier = Modifier.weight(1f),
                                    onClick = {
                                        onDeviceClick(actualDeviceId, actualCategory, device.name, actualGatewayId)
                                    }
                                )
                            } else {
                                Spacer(modifier = Modifier.weight(1f))
                            }
                        }
                    }

                    // Row 2
                    Row(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp) // Reduced spacing
                    ) {
                        for (i in 0 until 3) {
                            if (i < secondRowDevices.size) {
                                val device = secondRowDevices[i]
                                // For IR devices, use remote_id as deviceId and id as gatewayId
                                val hasRemoteId = !device.remoteId.isNullOrBlank()
                                val actualDeviceId = if (hasRemoteId) device.remoteId!! else device.id
                                val actualGatewayId = if (hasRemoteId) device.id else device.gatewayId
                                val rawCategory = if (!device.remoteCategory.isNullOrBlank()) device.remoteCategory else device.category ?: "unknown"
                                val actualCategory = if (rawCategory.isBlank()) "unknown" else rawCategory
                                DeviceItem(
                                    device = device,
                                    modifier = Modifier.weight(1f),
                                    onClick = {
                                        onDeviceClick(actualDeviceId, actualCategory, device.name, actualGatewayId)
                                    }
                                )
                            } else {
                                Spacer(modifier = Modifier.weight(1f))
                            }
                        }
                    }
                }
            }
        }
    }
}


@Composable
fun DeviceItem(device: Device, modifier: Modifier = Modifier, onClick: () -> Unit) {
    val saturationMatrix = remember { ColorMatrix() }
    LaunchedEffect(device.online) {
        saturationMatrix.setToSaturation(if (device.online) 1f else 0f)
    }

    Card(
        modifier = modifier
            .fillMaxWidth()
            .padding(4.dp)
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            Column(
                modifier = Modifier
                    .padding(8.dp)
                    .fillMaxSize()
                    .alpha(if (device.online) 1f else 0.6f), // Dim content if offline
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                Spacer(modifier = Modifier.height(16.dp)) // Space for badge

                // Icon
                val iconUrl = "https://images.tuyacn.com/${device.icon ?: ""}"
                AsyncImage(
                    model = iconUrl,
                    contentDescription = null,
                    modifier = Modifier.size(64.dp),
                    placeholder = rememberVectorPainter(Icons.Default.Home),
                    error = rememberVectorPainter(Icons.Default.Home),
                    colorFilter = ColorFilter.colorMatrix(saturationMatrix)
                )

                Spacer(modifier = Modifier.height(12.dp))

                // Name
                Text(
                    text = device.name,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                    color = if (device.online) Color.Unspecified else Color.Gray
                )
            }

        }
    }
}
