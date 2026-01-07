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
    
    // Pagination State
    var currentPage by remember { mutableIntStateOf(1) }
    var totalItems by remember { mutableIntStateOf(0) }
    val itemsPerPage = 6 // 2 rows x 3 columns
    
    val snackbarHostState = remember { SnackbarHostState() }

    val fetchDevices = { page: Int ->
        scope.launch {
            isLoading = true
            error = null
            try {
                // Pass page and limit to the API
                val response = RetrofitClient.instance.getDevicesByTeraluxId(
                    "Bearer $token", 
                    teraluxId,
                    page = page,
                    limit = itemsPerPage
                )
                if (response.isSuccessful && response.body() != null) {
                    val respData = response.body()!!.data
                    val rawDevices = respData?.devices ?: emptyList()
                    
                    val flatList = rawDevices.flatMap { d ->
                        val parsedCollections = d.getParsedCollections()
                        if (parsedCollections.isEmpty()) listOf(d) else parsedCollections
                    }
                    // Force default online status to true (Requested: Treat all devices as online)
                    devices = flatList.map { it.copy(online = true) }
                    
                    // Update metadata
                    totalItems = respData?.total ?: 0
                    currentPage = respData?.page ?: page
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
        fetchDevices(1)
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
                        // Reset to page 1 on refresh
                        currentPage = 1
                        fetchDevices(1)
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
                        Button(onClick = { fetchDevices(currentPage) }) {
                            Text("Retry")
                        }
                    }
                }
            } else {
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .padding(8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    // Render Grid Dynamically based on current 'devices' list (which is already paginated by backend)
                    // We expect up to 6 items. Break them into 2 rows of 3.
                    val rows = devices.chunked(3)
                    
                    // Always render 2 rows to keep layout stable, even if empty
                    for (i in 0 until 2) {
                         if (i < rows.size) {
                            val rowDevices = rows[i]
                            Row(
                                modifier = Modifier.weight(1f).fillMaxWidth(),
                                horizontalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                for (j in 0 until 3) {
                                    if (j < rowDevices.size) {
                                        val device = rowDevices[j]
                                        // Logic for ID mapping
                                        val hasRemoteId = !device.remoteId.isNullOrBlank()
                                        val actualDeviceId = device.id
                                        val actualGatewayId = if (hasRemoteId) device.remoteId else device.gatewayId
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
                                        // Empty slot
                                        Spacer(modifier = Modifier.weight(1f))
                                    }
                                }
                            }
                        } else {
                            // Empty Row
                             Row(modifier = Modifier.weight(1f).fillMaxWidth()) {
                                 Spacer(modifier = Modifier.weight(1f))
                             }
                        }
                    }

                    // Pagination Control Bar
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = 8.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Button(
                            onClick = {
                                if (currentPage > 1) {
                                    val newPage = currentPage - 1
                                    fetchDevices(newPage)
                                }
                            },
                            enabled = currentPage > 1
                        ) {
                            Text("Previous")
                        }

                        Text(
                            text = "Page $currentPage / ${ kotlin.math.max(1, (totalItems + itemsPerPage - 1) / itemsPerPage) }",
                            style = MaterialTheme.typography.bodyMedium
                        )

                        val maxPage = (totalItems + itemsPerPage - 1) / itemsPerPage
                        Button(
                            onClick = {
                                if (currentPage < maxPage) {
                                    val newPage = currentPage + 1
                                    fetchDevices(newPage)
                                }
                            },
                            enabled = currentPage < maxPage
                        ) {
                            Text("Next")
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
