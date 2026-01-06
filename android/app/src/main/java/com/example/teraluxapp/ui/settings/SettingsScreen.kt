package com.example.teraluxapp.ui.settings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.example.teraluxapp.data.model.*
import com.example.teraluxapp.data.network.RetrofitClient
import com.example.teraluxapp.utils.DeviceInfoUtils
import com.example.teraluxapp.utils.PreferencesManager
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    token: String,
    onBack: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val snackbarHostState = remember { SnackbarHostState() }
    
    // State
    val teraluxId = remember { PreferencesManager.getTeraluxId(context) ?: "" }
    val macAddress = remember { DeviceInfoUtils.getMacAddress(context) }
    var teraluxData by remember { mutableStateOf<Teralux?>(null) }
    var linkedDevices by remember { mutableStateOf<List<Device>>(emptyList()) }
    var roomId by remember { mutableStateOf("") }
    var teraluxName by remember { mutableStateOf("") }
    var isLoading by remember { mutableStateOf(true) }
    var showAddDeviceDialog by remember { mutableStateOf(false) }
    var showDeleteConfirmDialog by remember { mutableStateOf<Device?>(null) }
    var showClearCacheDialog by remember { mutableStateOf(false) }
    var allTuyaDevices by remember { mutableStateOf<List<Device>>(emptyList()) }
    
    // Helper function to fetch linked devices
    suspend fun fetchLinkedDevices() {
        try {
            val response = RetrofitClient.instance.getDevicesByTeraluxId("Bearer $token", teraluxId)
            if (response.isSuccessful && response.body() != null) {
                linkedDevices = response.body()!!.data?.devices ?: emptyList()
            }
        } catch (e: Exception) {
            // Log error or show snackbar if critical
        }
    }

    // Fetch Teralux data
    LaunchedEffect(Unit) {
        scope.launch {
            try {
                // Fetch Teralux Data
                val responseInfo = RetrofitClient.instance.getTeraluxById("Bearer $token", teraluxId)
                if (responseInfo.isSuccessful && responseInfo.body() != null) {
                    teraluxData = responseInfo.body()!!.data?.teralux
                    roomId = teraluxData?.roomId ?: ""
                    teraluxName = teraluxData?.name ?: ""
                }

                // Fetch Linked Devices
                fetchLinkedDevices()
                
                // Fetch All Tuya Devices (with flattening logic)
                val responseDevices = RetrofitClient.instance.getDevices("Bearer $token", page = 1, limit = 100)
                if (responseDevices.isSuccessful && responseDevices.body() != null) {
                    val rawDevices = responseDevices.body()!!.data?.devices ?: emptyList()
                    
                    // Flatten the list: If device has collections (IR Hub), add collections instead of the hub
                    val flatList = ArrayList<Device>()
                    for (d in rawDevices) {
                        val parsedCollections = d.getParsedCollections()
                        if (parsedCollections.isEmpty()) {
                            flatList.add(d)
                        } else {
                            flatList.addAll(parsedCollections)
                        }
                    }
                    allTuyaDevices = flatList
                    
                    // Automatically show add dialog if check implies it, but here we just load data
                }
            } catch (e: Exception) {
                snackbarHostState.showSnackbar("Error loading settings: ${e.message}")
            } finally {
                isLoading = false
            }
        }
    }
    
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Settings") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = Color(0xFF8B5CF6),
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
                )
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { paddingValues ->
        if (isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator(color = Color(0xFF8B5CF6))
            }
        } else {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues)
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                // Device Information Section
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = "📱 Device Information",
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFF1E293B)
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        
                        OutlinedTextField(
                            value = teraluxId,
                            onValueChange = {},
                            label = { Text("Teralux ID") },
                            enabled = false,
                            modifier = Modifier.fillMaxWidth(),
                            colors = OutlinedTextFieldDefaults.colors(
                                disabledTextColor = Color(0xFF64748B),
                                disabledBorderColor = Color(0xFFCBD5E1),
                                disabledLabelColor = Color(0xFF94A3B8)
                            )
                        )
                        
                        Spacer(modifier = Modifier.height(8.dp))
                        
                        OutlinedTextField(
                            value = macAddress,
                            onValueChange = {},
                            label = { Text("MAC Address") },
                            enabled = false,
                            modifier = Modifier.fillMaxWidth(),
                            colors = OutlinedTextFieldDefaults.colors(
                                disabledTextColor = Color(0xFF64748B),
                                disabledBorderColor = Color(0xFFCBD5E1),
                                disabledLabelColor = Color(0xFF94A3B8)
                            )
                        )
                    }
                }
                
                // Teralux Configuration Section
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = "🛠️ Teralux Configuration",
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFF1E293B)
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        
                        OutlinedTextField(
                            value = teraluxName,
                            onValueChange = { teraluxName = it },
                            label = { Text("Device Name") },
                            modifier = Modifier.fillMaxWidth(),
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = Color(0xFF8B5CF6),
                                focusedLabelColor = Color(0xFF8B5CF6)
                            )
                        )
                        
                        Spacer(modifier = Modifier.height(12.dp))
                        
                        OutlinedTextField(
                            value = roomId,
                            onValueChange = { roomId = it },
                            label = { Text("Room ID") },
                            modifier = Modifier.fillMaxWidth(),
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = Color(0xFF8B5CF6),
                                focusedLabelColor = Color(0xFF8B5CF6)
                            )
                        )
                        
                        Spacer(modifier = Modifier.height(16.dp))
                        
                        Button(
                            onClick = {
                                scope.launch {
                                    try {
                                        val response = RetrofitClient.instance.updateTeralux(
                                            "Bearer $token",
                                            teraluxId,
                                            UpdateTeraluxRequest(
                                                roomId = roomId,
                                                name = teraluxName
                                            )
                                        )
                                        if (response.isSuccessful) {
                                            snackbarHostState.showSnackbar("Configuration updated successfully")
                                        } else {
                                            snackbarHostState.showSnackbar("Failed to update: ${response.code()}")
                                        }
                                    } catch (e: Exception) {
                                        snackbarHostState.showSnackbar("Error: ${e.message}")
                                    }
                                }
                            },
                            modifier = Modifier.fillMaxWidth(),
                            colors = ButtonDefaults.buttonColors(
                                containerColor = Color(0xFF8B5CF6)
                            )
                        ) {
                            Text("Update Configuration")
                        }
                    }
                }
                
                // Linked Devices Section
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = "🔌 Linked Devices",
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFF1E293B)
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        
                        // List of linked devices
                        linkedDevices.forEach { device ->
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 4.dp),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "• ${device.name}",
                                    modifier = Modifier.weight(1f),
                                    color = Color(0xFF1E293B)
                                )
                                IconButton(onClick = { showDeleteConfirmDialog = device }) {
                                    Icon(
                                        Icons.Default.Delete,
                                        contentDescription = "Delete",
                                        tint = Color(0xFFEF4444)
                                    )
                                }
                            }
                        }
                        
                        if (linkedDevices.isEmpty()) {
                            Text(
                                text = "No devices linked yet",
                                color = Color(0xFF94A3B8),
                                fontSize = 14.sp
                            )
                        }
                        
                        Spacer(modifier = Modifier.height(12.dp))
                        
                        OutlinedButton(
                            onClick = {
                                scope.launch {
                                    try {
                                        val response = RetrofitClient.instance.getDevices("Bearer $token", page = 1, limit = 100)
                                        if (response.isSuccessful && response.body() != null) {
                                            val rawDevices = response.body()!!.data?.devices ?: emptyList()
                                            
                                            val flatList = ArrayList<Device>()
                                            for (d in rawDevices) {
                                                val parsedCollections = d.getParsedCollections()
                                                if (parsedCollections.isEmpty()) {
                                                    flatList.add(d)
                                                } else {
                                                    flatList.addAll(parsedCollections)
                                                }
                                            }
                                            allTuyaDevices = flatList
                                            showAddDeviceDialog = true
                                        }
                                    } catch (e: Exception) {
                                        snackbarHostState.showSnackbar("Error loading devices: ${e.message}")
                                    }
                                }
                            },
                            modifier = Modifier.fillMaxWidth(),
                            colors = ButtonDefaults.outlinedButtonColors(
                                contentColor = Color(0xFF8B5CF6)
                            )
                        ) {
                            Icon(Icons.Default.Add, contentDescription = null)
                            Spacer(Modifier.width(8.dp))
                            Text("Add Device")
                        }
                    }
                }
                
                // System Section
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            text = "⚙️ System",
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFF1E293B)
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        
                        OutlinedButton(
                            onClick = { showClearCacheDialog = true },
                            modifier = Modifier.fillMaxWidth(),
                            colors = ButtonDefaults.outlinedButtonColors(
                                contentColor = Color(0xFFEF4444)
                            )
                        ) {
                            Text("Clear Cache")
                        }
                    }
                }
            }
        }
    }
    
    // Add Device Dialog
    if (showAddDeviceDialog) {
        Dialog(onDismissRequest = { showAddDeviceDialog = false }) {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .heightIn(max = 600.dp),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White)
            ) {
                Column(modifier = Modifier.padding(24.dp)) {
                    Text(
                        text = "Add Device",
                        fontSize = 20.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFF1E293B)
                    )
                    Text(
                        text = "Select a device to link to this room",
                        fontSize = 14.sp,
                        color = Color(0xFF64748B)
                    )
                    
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    val availableDevices = allTuyaDevices.filter { device ->
                        linkedDevices.none { it.id == device.id }
                    }

                    if (availableDevices.isEmpty()) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(32.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = "No available devices found",
                                color = Color(0xFF94A3B8)
                            )
                        }
                    } else {
                        Column(
                            modifier = Modifier
                                .weight(1f, fill = false)
                                .verticalScroll(rememberScrollState())
                        ) {
                            availableDevices.forEach { device ->
                                Card(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(vertical = 4.dp),
                                    colors = CardDefaults.outlinedCardColors(
                                        containerColor = Color(0xFFF8FAFC),
                                        contentColor = Color(0xFF1E293B)
                                    ),
                                    border = androidx.compose.foundation.BorderStroke(1.dp, Color(0xFFE2E8F0)),
                                    onClick = {
                                        scope.launch {
                                            var isAdding = false
                                            try {
                                                isAdding = true
                                                val response = RetrofitClient.instance.createDevice(
                                                    "Bearer $token",
                                                    CreateDeviceRequest(
                                                        id = device.id,
                                                        teraluxId = teraluxId,
                                                        name = device.name
                                                    )
                                                )
                                                if (response.isSuccessful) {
                                                    // Refresh teralux data and linked devices
                                                    fetchLinkedDevices()
                                                    val refreshResponse = RetrofitClient.instance.getTeraluxById("Bearer $token", teraluxId)
                                                    if (refreshResponse.isSuccessful) {
                                                        teraluxData = refreshResponse.body()!!.data?.teralux
                                                    }
                                                    
                                                    snackbarHostState.showSnackbar("Device added successfully")
                                                    showAddDeviceDialog = false
                                                } else {
                                                    snackbarHostState.showSnackbar("Failed to add device")
                                                }
                                            } catch (e: Exception) {
                                                snackbarHostState.showSnackbar("Error: ${e.message}")
                                            } finally {
                                                isAdding = false
                                            }
                                        }
                                    }
                                ) {
                                    Row(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .padding(16.dp),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        // Device Icon Placeholder
                                        Surface(
                                            modifier = Modifier.size(40.dp),
                                            shape = CircleShape,
                                            color = Color(0xFFEDE9FE)
                                        ) {
                                            Box(contentAlignment = Alignment.Center) {
                                                Icon(
                                                    Icons.Default.Add, // Could be device type icon
                                                    contentDescription = null,
                                                    tint = Color(0xFF8B5CF6),
                                                    modifier = Modifier.size(20.dp)
                                                )
                                            }
                                        }
                                        
                                        Spacer(modifier = Modifier.width(16.dp))
                                        
                                        Column {
                                            Text(
                                                text = device.name,
                                                fontWeight = FontWeight.SemiBold,
                                                fontSize = 16.sp,
                                                color = Color(0xFF1E293B)
                                            )
                                            Text(
                                                text = "ID: ${device.id}",
                                                fontSize = 12.sp,
                                                color = Color(0xFF64748B),
                                                fontFamily = androidx.compose.ui.text.font.FontFamily.Monospace
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                    
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    TextButton(
                        onClick = { showAddDeviceDialog = false },
                        modifier = Modifier.align(Alignment.End),
                        colors = ButtonDefaults.textButtonColors(contentColor = Color(0xFF64748B))
                    ) {
                        Text("Cancel")
                    }
                }
            }
        }
    }
    
    // Delete Confirmation Dialog
    showDeleteConfirmDialog?.let { device ->
        AlertDialog(
            onDismissRequest = { showDeleteConfirmDialog = null },
            title = { Text("Delete Device") },
            text = { Text("Are you sure you want to remove ${device.name} from this Teralux?") },
            confirmButton = {
                TextButton(
                    onClick = {
                        scope.launch {
                            try {
                                val response = RetrofitClient.instance.deleteDevice("Bearer $token", device.id)
                                if (response.isSuccessful) {
                                    snackbarHostState.showSnackbar("Device removed successfully")
                                    showDeleteConfirmDialog = null
                                    // Refresh data
                                    fetchLinkedDevices()
                                    val refreshResponse = RetrofitClient.instance.getTeraluxById("Bearer $token", teraluxId)
                                    if (refreshResponse.isSuccessful) {
                                        teraluxData = refreshResponse.body()!!.data?.teralux
                                    }
                                } else {
                                    snackbarHostState.showSnackbar("Failed to remove device")
                                }
                            } catch (e: Exception) {
                                snackbarHostState.showSnackbar("Error: ${e.message}")
                            }
                        }
                    },
                    colors = ButtonDefaults.textButtonColors(contentColor = Color(0xFFEF4444))
                ) {
                    Text("Delete")
                }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteConfirmDialog = null }) {
                    Text("Cancel")
                }
            }
        )
    }
    
    // Clear Cache Confirmation Dialog
    if (showClearCacheDialog) {
        AlertDialog(
            onDismissRequest = { showClearCacheDialog = false },
            title = { Text("Clear Cache") },
            text = { Text("Are you sure you want to clear the cache? This will refresh all data.") },
            confirmButton = {
                TextButton(
                    onClick = {
                        scope.launch {
                            try {
                                val response = RetrofitClient.instance.flushCache("Bearer $token")
                                if (response.isSuccessful) {
                                    snackbarHostState.showSnackbar("Cache cleared successfully")
                                    showClearCacheDialog = false
                                } else {
                                    snackbarHostState.showSnackbar("Failed to clear cache")
                                }
                            } catch (e: Exception) {
                                snackbarHostState.showSnackbar("Error: ${e.message}")
                            }
                        }
                    },
                    colors = ButtonDefaults.textButtonColors(contentColor = Color(0xFFEF4444))
                ) {
                    Text("Clear")
                }
            },
            dismissButton = {
                TextButton(onClick = { showClearCacheDialog = false }) {
                    Text("Cancel")
                }
            }
        )
    }
}
