package com.example.whisperandroid.presentation.dashboard

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.asPaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.icons.outlined.Groups
import androidx.compose.material.icons.outlined.SmartToy
import android.provider.Settings
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.runtime.DisposableEffect
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.presentation.components.DashboardFeatureCard
import kotlinx.coroutines.launch

private object DashboardLayoutTokens {
    val Spacing4 = 4.dp
    val Spacing8 = 8.dp
    val Spacing12 = 12.dp
    val Spacing16 = 16.dp
    val Spacing20 = 20.dp
    val Spacing24 = 24.dp
    val Spacing32 = 32.dp
    
    val CardRadiusPrimary = 20.dp
    val CardRadiusFeature = 24.dp
    val MaxContentWidth = 960.dp
}

@Composable
fun DashboardScreen(
    onNavigateToRegister: () -> Unit,
    onNavigateToUpload: () -> Unit,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    viewModel: DashboardViewModel =
        androidx.lifecycle.viewmodel.compose.viewModel {
            DashboardViewModel(
                NetworkModule.authenticateUseCase,
                NetworkModule.getTuyaDevicesUseCase,
                NetworkModule.backgroundAssistantModeStore
            )
        }
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current

    val lifecycleOwner = LocalLifecycleOwner.current

    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.fetchDevices()
    }

    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_RESUME) {
                viewModel.checkOverlayPermission(context)
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
        }
    }
    var hasMicPermission by remember {
        mutableStateOf(
            androidx.core.content.ContextCompat.checkSelfPermission(
                context,
                android.Manifest.permission.RECORD_AUDIO
            ) == android.content.pm.PackageManager.PERMISSION_GRANTED
        )
    }

    val launcher =
        androidx.activity.compose.rememberLauncherForActivityResult(
            androidx.activity.result.contract.ActivityResultContracts
                .RequestPermission()
        ) { isGranted ->
            hasMicPermission = isGranted
        }

    val snackbarHostState = remember { androidx.compose.material3.SnackbarHostState() }
    val scope = androidx.compose.runtime.rememberCoroutineScope()

    androidx.compose.material3.Scaffold(
        snackbarHost = { androidx.compose.material3.SnackbarHost(snackbarHostState) },
        containerColor = Color.Transparent,
        contentWindowInsets = WindowInsets(0, 0, 0, 0)
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.background)
                .padding(innerPadding)
                .padding(WindowInsets.systemBars.asPaddingValues())
        ) {
            // Background gradient
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.radialGradient(
                            colors = listOf(
                                MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.15f),
                                Color.Transparent
                            ),
                            center = androidx.compose.ui.geometry.Offset(0f, 0f),
                            radius = 2000f
                        )
                    )
            )

            if (uiState.isLoading) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
                }
            } else if (uiState.isAuthenticated) {
                // Centered content container
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.TopCenter
                ) {
                    Column(
                        modifier = Modifier
                            .widthIn(max = DashboardLayoutTokens.MaxContentWidth)
                            .fillMaxWidth()
                            .padding(horizontal = DashboardLayoutTokens.Spacing24)
                    ) {
                        DashboardContent(
                            onNavigateToStreaming = onNavigateToStreaming,
                            onNavigateToEdge = onNavigateToEdge,
                            isBackgroundModeEnabled = uiState.isBackgroundModeEnabled,
                            isOverlayPermissionGranted = uiState.isOverlayPermissionGranted,
                            onBackgroundModeChange = { viewModel.setBackgroundMode(it) },
                            onRequestOverlayPermission = {
                                val intent = android.content.Intent(
                                    Settings.ACTION_MANAGE_OVERLAY_PERMISSION,
                                    android.net.Uri.parse("package:${context.packageName}")
                                )
                                context.startActivity(intent)
                            },
                            onShowDisabledMessage = {
                                scope.launch {
                                    snackbarHostState.showSnackbar(
                                        "Unavailable while Background Assistant is active."
                                    )
                                }
                            }
                        )
                    }
                }
            } else {
                Column(
                    modifier = Modifier
                        .align(Alignment.Center)
                        .padding(DashboardLayoutTokens.Spacing24),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing16)
                ) {
                    Text(
                        text = uiState.error ?: "Authentication Failed",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White,
                        textAlign = TextAlign.Center
                    )
                    Button(onClick = { viewModel.authenticate() }) {
                        Text("Retry Login")
                    }
                }
            }
        }
    }
}

@Composable
private fun DashboardHeader() {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.padding(top = DashboardLayoutTokens.Spacing24)
    ) {
        Text(
            text = "Select Workspace",
            style = MaterialTheme.typography.headlineMedium.copy(
                fontWeight = FontWeight.Black,
                letterSpacing = (-0.5).sp
            ),
            color = MaterialTheme.colorScheme.onBackground,
            textAlign = TextAlign.Center,
            modifier = Modifier.testTag("dashboard_header")
        )
        Spacer(modifier = Modifier.height(DashboardLayoutTokens.Spacing8))
        Text(
            text = "Choose your AI-powered environment",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
            fontWeight = FontWeight.Medium,
            textAlign = TextAlign.Center
        )
    }
}

@Composable
private fun BackgroundAssistantCard(
    isEnabled: Boolean,
    isOverlayPermissionGranted: Boolean,
    onEnabledChange: (Boolean) -> Unit,
    onRequestOverlayPermission: () -> Unit
) {
    androidx.compose.material3.Surface(
        modifier = Modifier.fillMaxWidth(),
        color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.08f),
        shape = RoundedCornerShape(DashboardLayoutTokens.CardRadiusPrimary),
        border = androidx.compose.foundation.BorderStroke(
            1.dp,
            MaterialTheme.colorScheme.primary.copy(alpha = if (isEnabled) 0.15f else 0.04f)
        )
    ) {
        Column(modifier = Modifier.padding(DashboardLayoutTokens.Spacing16)) {
            // Row 1: Title + Badge + Switch
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = "Background Assistant",
                        style = MaterialTheme.typography.titleSmall,
                        color = MaterialTheme.colorScheme.onSurface,
                        fontWeight = FontWeight.ExtraBold
                    )
                    Spacer(modifier = Modifier.width(DashboardLayoutTokens.Spacing8))
                    if (isEnabled) {
                        // Status Badge
                        Box(
                            modifier = Modifier
                                .background(
                                    MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                                    CircleShape
                                )
                                .padding(horizontal = 6.dp, vertical = 2.dp)
                        ) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Box(
                                    modifier = Modifier
                                        .size(5.dp)
                                        .background(MaterialTheme.colorScheme.primary, CircleShape)
                                )
                                Spacer(modifier = Modifier.width(4.dp))
                                Text(
                                    text = "Active",
                                    style = MaterialTheme.typography.labelSmall.copy(fontSize = 9.sp),
                                    color = MaterialTheme.colorScheme.primary,
                                    fontWeight = FontWeight.Bold
                                )
                            }
                        }
                    } else {
                        Text(
                            text = "Inactive",
                            style = MaterialTheme.typography.labelSmall.copy(fontSize = 9.sp),
                            color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f),
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
                androidx.compose.material3.Switch(
                    checked = isEnabled,
                    onCheckedChange = onEnabledChange,
                    modifier = Modifier.scale(0.7f)
                )
            }

            // Row 2: Helper Text
            Text(
                text = "Runs across apps with wake word support.",
                style = MaterialTheme.typography.labelSmall.copy(fontSize = 11.sp),
                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f),
                modifier = Modifier.padding(top = 0.dp)
            )

            if (isEnabled) {
                Spacer(modifier = Modifier.height(DashboardLayoutTokens.Spacing12))
                
                Row(
                   modifier = Modifier.fillMaxWidth(),
                   horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing8),
                   verticalAlignment = Alignment.Top
                ) {
                    // Feature Lock Impact (Condensed)
                    Row(
                        modifier = Modifier
                            .weight(1.1f)
                            .background(
                                MaterialTheme.colorScheme.primary.copy(alpha = 0.05f),
                                RoundedCornerShape(DashboardLayoutTokens.Spacing4)
                            )
                            .padding(horizontal = 8.dp, vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing4)
                    ) {
                        Icon(
                            imageVector = Icons.Filled.Lock,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.primary.copy(alpha = 0.7f),
                            modifier = Modifier.size(12.dp)
                        )
                        Text(
                            text = "Features locked while active.",
                            style = MaterialTheme.typography.labelSmall.copy(fontSize = 10.sp),
                            color = MaterialTheme.colorScheme.primary,
                            fontWeight = FontWeight.Medium
                        )
                    }

                    if (!isOverlayPermissionGranted) {
                        // Permission Warning (Condensed)
                        Row(
                            modifier = Modifier
                                .weight(1f)
                                .background(
                                    MaterialTheme.colorScheme.error.copy(alpha = 0.05f),
                                    RoundedCornerShape(DashboardLayoutTokens.Spacing4)
                                )
                                .padding(horizontal = 8.dp, vertical = 2.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing4)
                            ) {
                                Icon(
                                    imageVector = Icons.Filled.Warning,
                                    contentDescription = null,
                                    tint = MaterialTheme.colorScheme.error,
                                    modifier = Modifier.size(12.dp)
                                )
                                Text(
                                    text = "Overlay Required",
                                    style = MaterialTheme.typography.labelSmall.copy(fontSize = 10.sp),
                                    color = MaterialTheme.colorScheme.error,
                                    fontWeight = FontWeight.Bold
                                )
                            }
                            androidx.compose.material3.TextButton(
                                onClick = onRequestOverlayPermission,
                                contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 4.dp),
                                modifier = Modifier.height(24.dp)
                            ) {
                                Text("Grant", fontSize = 10.sp, fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.error)
                            }
                        }
                    } else {
                        // Status: OK (Condensed)
                        Row(
                            modifier = Modifier
                                .weight(1f)
                                .padding(horizontal = 4.dp, vertical = 6.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing4)
                        ) {
                            Icon(
                                imageVector = Icons.Filled.Check,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(12.dp)
                            )
                            Text(
                                text = "Overlay active",
                                style = MaterialTheme.typography.labelSmall.copy(fontSize = 10.sp),
                                color = MaterialTheme.colorScheme.primary.copy(alpha = 0.8f),
                                fontWeight = FontWeight.Medium
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun FeatureGrid(
    isBackgroundModeEnabled: Boolean,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    onShowDisabledMessage: () -> Unit
) {
    BoxWithConstraints(modifier = Modifier.fillMaxWidth()) {
        val isWide = maxWidth > 600.dp
        
        if (isWide) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing16)
            ) {
                FeatureCard(
                    title = "Meeting Transcriber",
                    description = "Live transcription and AI summaries.",
                    icon = Icons.Outlined.Groups,
                    onClick = onNavigateToStreaming,
                    isEnabled = !isBackgroundModeEnabled,
                    onDisabledClick = onShowDisabledMessage,
                    modifier = Modifier.weight(1f)
                )
                FeatureCard(
                    title = "AI Assistant",
                    description = "Direct conversational AI interaction.",
                    icon = Icons.Outlined.SmartToy,
                    onClick = onNavigateToEdge,
                    isEnabled = !isBackgroundModeEnabled,
                    onDisabledClick = onShowDisabledMessage,
                    modifier = Modifier.weight(1f)
                )
            }
        } else {
            Column(
                verticalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing16)
            ) {
                FeatureCard(
                    title = "Meeting Transcriber",
                    description = "Live transcription and AI summaries.",
                    icon = Icons.Outlined.Groups,
                    onClick = onNavigateToStreaming,
                    isEnabled = !isBackgroundModeEnabled,
                    onDisabledClick = onShowDisabledMessage
                )
                FeatureCard(
                    title = "AI Assistant",
                    description = "Direct conversational AI interaction.",
                    icon = Icons.Outlined.SmartToy,
                    onClick = onNavigateToEdge,
                    isEnabled = !isBackgroundModeEnabled,
                    onDisabledClick = onShowDisabledMessage
                )
            }
        }
    }
}

@Composable
private fun FeatureCard(
    title: String,
    description: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit,
    isEnabled: Boolean,
    onDisabledClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    DashboardFeatureCard(
        title = title,
        description = description,
        icon = {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.primary,
                modifier = Modifier.size(48.dp)
            )
        },
        onClick = {
            if (isEnabled) onClick() else onDisabledClick()
        },
        enabled = isEnabled,
        modifier = modifier.height(200.dp)
    )
}

@Composable
fun DashboardContent(
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    isBackgroundModeEnabled: Boolean,
    isOverlayPermissionGranted: Boolean,
    onBackgroundModeChange: (Boolean) -> Unit,
    onRequestOverlayPermission: () -> Unit,
    onShowDisabledMessage: () -> Unit
) {
    Column(
        modifier = Modifier.fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing32)
    ) {
        DashboardHeader()

        BackgroundAssistantCard(
            isEnabled = isBackgroundModeEnabled,
            isOverlayPermissionGranted = isOverlayPermissionGranted,
            onEnabledChange = onBackgroundModeChange,
            onRequestOverlayPermission = onRequestOverlayPermission
        )

        Column(verticalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing20)) {
            Text(
                text = "WORKSPACE FEATURES",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.4f),
                fontWeight = FontWeight.Bold,
                letterSpacing = 1.sp
            )
            FeatureGrid(
                isBackgroundModeEnabled = isBackgroundModeEnabled,
                onNavigateToStreaming = onNavigateToStreaming,
                onNavigateToEdge = onNavigateToEdge,
                onShowDisabledMessage = onShowDisabledMessage
            )
        }

        Spacer(modifier = Modifier.weight(1f))

        // Footer
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing8),
            modifier = Modifier.padding(bottom = DashboardLayoutTokens.Spacing16)
        ) {
            Box(
                modifier = Modifier
                    .size(4.dp)
                    .background(MaterialTheme.colorScheme.primary, CircleShape)
            )
            Text(
                text = "Powered by Sensio",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.4f),
                fontWeight = FontWeight.SemiBold,
                letterSpacing = 1.sp
            )
        }
    }
}
