package com.example.whisperandroid.presentation.dashboard

import android.provider.Settings
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
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
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.icons.outlined.AutoAwesome
import androidx.compose.material.icons.outlined.Groups
import androidx.compose.material.icons.outlined.SmartToy
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.presentation.components.DashboardFeatureCard
import com.example.whisperandroid.util.DeviceUtils
import kotlinx.coroutines.launch

/**
 * Dashboard-specific device classification for responsive layout decisions.
 * Maps from low-level [DeviceUtils] detection to screen-specific layout contracts.
 */
enum class DashboardDeviceClass {
    TERALUX,
    TABLET,
    PHONE
}

/**
 * Layout mode for the AI provider selector.
 */
enum class ProviderSelectorMode {
    /** Single row, chips may wrap naturally */
    SINGLE_ROW,

    /** Multiple rows with balanced distribution */
    WRAPPED_ROWS
}

/**
 * Layout mode for the background assistant card.
 */
enum class BackgroundAssistantLayoutMode {
    /** All content in a single horizontal row */
    INLINE,

    /** Title/switch on one row, details below */
    SPLIT_ROWS
}

/**
 * Complete layout specification for the dashboard responsive variants.
 * Derived from device class, screen dimensions, and orientation.
 */
data class DashboardLayoutSpec(
    val deviceClass: DashboardDeviceClass,
    val isScrollable: Boolean,
    val contentMaxWidth: androidx.compose.ui.unit.Dp,
    val outerHorizontalPadding: androidx.compose.ui.unit.Dp,
    val sectionSpacing: androidx.compose.ui.unit.Dp,
    val featureColumns: Int,
    val providerSelectorMode: ProviderSelectorMode,
    val backgroundAssistantMode: BackgroundAssistantLayoutMode,
    val isLandscape: Boolean,
    val availableWidth: androidx.compose.ui.unit.Dp
)

/**
 * Computes the dashboard layout specification based on device class and screen metrics.
 * This is the single source of truth for responsive layout decisions.
 */
@Composable
private fun rememberDashboardLayoutSpec(): DashboardLayoutSpec {
    val context = LocalContext.current
    val configuration = LocalConfiguration.current
    val screenWidth = configuration.screenWidthDp.dp
    val screenHeight = configuration.screenHeightDp.dp
    val isLandscape = configuration.orientation == android.content.res.Configuration.ORIENTATION_LANDSCAPE

    val deviceClass = when {
        DeviceUtils.isTerminal() -> DashboardDeviceClass.TERALUX
        DeviceUtils.isTablet(context) -> DashboardDeviceClass.TABLET
        else -> DashboardDeviceClass.PHONE
    }

    return when (deviceClass) {
        DashboardDeviceClass.TERALUX -> {
            // Teralux: landscape-first, wide content, two-panel layout
            DashboardLayoutSpec(
                deviceClass = DashboardDeviceClass.TERALUX,
                isScrollable = !isLandscape || screenHeight < 600.dp,
                contentMaxWidth = 1200.dp,
                outerHorizontalPadding = 48.dp,
                sectionSpacing = 32.dp,
                featureColumns = 2,
                providerSelectorMode = if (screenWidth > 800.dp) ProviderSelectorMode.SINGLE_ROW else ProviderSelectorMode.WRAPPED_ROWS,
                backgroundAssistantMode = BackgroundAssistantLayoutMode.INLINE,
                isLandscape = isLandscape,
                availableWidth = screenWidth
            )
        }
        DashboardDeviceClass.TABLET -> {
            // Tablet: medium-wide content, balanced spacing
            DashboardLayoutSpec(
                deviceClass = DashboardDeviceClass.TABLET,
                isScrollable = screenHeight < 700.dp,
                contentMaxWidth = 960.dp,
                outerHorizontalPadding = 32.dp,
                sectionSpacing = 28.dp,
                featureColumns = 2,
                providerSelectorMode = if (screenWidth > 600.dp) ProviderSelectorMode.SINGLE_ROW else ProviderSelectorMode.WRAPPED_ROWS,
                backgroundAssistantMode = BackgroundAssistantLayoutMode.INLINE,
                isLandscape = isLandscape,
                availableWidth = screenWidth
            )
        }
        DashboardDeviceClass.PHONE -> {
            // Phone: compact, stacked, always scrollable
            DashboardLayoutSpec(
                deviceClass = DashboardDeviceClass.PHONE,
                isScrollable = true,
                contentMaxWidth = screenWidth,
                outerHorizontalPadding = 20.dp,
                sectionSpacing = 24.dp,
                featureColumns = 1,
                providerSelectorMode = ProviderSelectorMode.WRAPPED_ROWS,
                backgroundAssistantMode = BackgroundAssistantLayoutMode.SPLIT_ROWS,
                isLandscape = isLandscape,
                availableWidth = screenWidth
            )
        }
    }
}

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
    bootstrapViewModel: com.example.whisperandroid.presentation.bootstrap.AppBootstrapViewModel,
    viewModel: DashboardViewModel =
        androidx.lifecycle.viewmodel.compose.viewModel {
            DashboardViewModel(
                NetworkModule.authenticateUseCase,
                NetworkModule.getTuyaDevicesUseCase,
                NetworkModule.backgroundAssistantModeStore,
                NetworkModule.terminalRepository,
                NetworkModule.tokenManager
            )
        }
) {
    val uiState by viewModel.uiState.collectAsState()
    val bootstrapState by bootstrapViewModel.uiState.collectAsState()
    val context = LocalContext.current

    val lifecycleOwner = LocalLifecycleOwner.current

    androidx.compose.runtime.LaunchedEffect(Unit) {
        viewModel.fetchDevices(force = true)
    }

    var hasMicPermission by remember {
        mutableStateOf(
            androidx.core.content.ContextCompat.checkSelfPermission(
                context,
                android.Manifest.permission.RECORD_AUDIO
            ) == android.content.pm.PackageManager.PERMISSION_GRANTED
        )
    }

    var wasBackgroundModeEnabled by remember { mutableStateOf(uiState.isBackgroundModeEnabled) }

    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_RESUME) {
                viewModel.checkOverlayPermission(context)
                hasMicPermission = androidx.core.content.ContextCompat.checkSelfPermission(
                    context,
                    android.Manifest.permission.RECORD_AUDIO
                ) == android.content.pm.PackageManager.PERMISSION_GRANTED
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
        }
    }

    val launcher =
        androidx.activity.compose.rememberLauncherForActivityResult(
            androidx.activity.result.contract.ActivityResultContracts
                .RequestPermission()
        ) { isGranted ->
            hasMicPermission = isGranted
            if (isGranted) {
                viewModel.setBackgroundMode(context, true)
            }
        }

    val snackbarHostState = remember { androidx.compose.material3.SnackbarHostState() }
    val scope = androidx.compose.runtime.rememberCoroutineScope()

    // Handle redirect to register when terminal not found
    androidx.compose.runtime.LaunchedEffect(uiState.shouldRedirectToRegister) {
        if (uiState.shouldRedirectToRegister) {
            onNavigateToRegister()
        }
    }

    androidx.compose.runtime.LaunchedEffect(uiState.isBackgroundModeEnabled) {
        if (wasBackgroundModeEnabled && !uiState.isBackgroundModeEnabled) {
            val currentMicPermission = androidx.core.content.ContextCompat.checkSelfPermission(
                context,
                android.Manifest.permission.RECORD_AUDIO
            ) == android.content.pm.PackageManager.PERMISSION_GRANTED
            if (!currentMicPermission) {
                snackbarHostState.showSnackbar(
                    message = "Background Assistant turned off because microphone permission is disabled.",
                    duration = androidx.compose.material3.SnackbarDuration.Long
                )
            }
        }
        wasBackgroundModeEnabled = uiState.isBackgroundModeEnabled
    }

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

            if (bootstrapState.isSyncing) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            "Syncing devices...",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f)
                        )
                    }
                }
            } else if (uiState.error != null) {
                // Show error state (e.g., Terminal not found)
                uiState.error?.let { errorMessage ->
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center
                    ) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            modifier = Modifier.padding(24.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.Warning,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.error,
                                modifier = Modifier.size(48.dp)
                            )
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(
                                text = errorMessage,
                                modifier = Modifier.padding(horizontal = 16.dp),
                                color = MaterialTheme.colorScheme.error,
                                style = MaterialTheme.typography.bodyLarge
                            )
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(
                                text = "Redirecting to registration...",
                                modifier = Modifier.padding(horizontal = 16.dp),
                                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                                style = MaterialTheme.typography.bodyMedium
                            )
                        }
                    }
                }
            } else if (bootstrapState.isBootstrapped && uiState.isTuyaSyncReady) {
                // Centered content container - width is now controlled by each layout variant
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.TopCenter
                ) {
                    DashboardContent(
                        onNavigateToStreaming = onNavigateToStreaming,
                        onNavigateToEdge = onNavigateToEdge,
                        isBackgroundModeEnabled = uiState.isBackgroundModeEnabled,
                        isOverlayPermissionGranted = uiState.isOverlayPermissionGranted,
                        onBackgroundModeChange = { enabled ->
                            if (enabled) {
                                if (hasMicPermission) {
                                    viewModel.setBackgroundMode(context, true)
                                } else {
                                    launcher.launch(android.Manifest.permission.RECORD_AUDIO)
                                }
                            } else {
                                viewModel.setBackgroundMode(context, false)
                            }
                        },
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
                        },
                        aiProvider = uiState.aiProvider,
                        isSavingAiProvider = uiState.isSavingAiProvider,
                        onAiProviderChange = { provider ->
                            viewModel.updateAiProvider(provider)
                        }
                    )
                }
            } else if (bootstrapState.error != null) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier.padding(24.dp)
                    ) {
                        Text(
                            bootstrapState.error ?: "Unknown error",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.error,
                            textAlign = androidx.compose.ui.text.style.TextAlign.Center
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        androidx.compose.material3.Button(
                            onClick = { bootstrapViewModel.bootstrap(forceRetry = true) }
                        ) {
                            Text("Retry")
                        }
                    }
                }
            } else {
                // Should not happen, but as a fallback
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
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
    onRequestOverlayPermission: () -> Unit,
    layoutSpec: DashboardLayoutSpec
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
            when (layoutSpec.backgroundAssistantMode) {
                BackgroundAssistantLayoutMode.INLINE -> {
                    // Inline layout for tablets and teralux
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
                        // Full-size switch for better tap target
                        androidx.compose.material3.Switch(
                            checked = isEnabled,
                            onCheckedChange = onEnabledChange
                        )
                    }

                    Text(
                        text = "Runs across apps with wake word support.",
                        style = MaterialTheme.typography.labelSmall.copy(fontSize = 11.sp),
                        color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f),
                        modifier = Modifier.padding(top = DashboardLayoutTokens.Spacing4)
                    )

                    if (isEnabled) {
                        Spacer(modifier = Modifier.height(DashboardLayoutTokens.Spacing12))

                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing8),
                            verticalAlignment = Alignment.Top
                        ) {
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
                BackgroundAssistantLayoutMode.SPLIT_ROWS -> {
                    // Split layout for phones: title/switch on first row, details below
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
                        // Full-size switch for better tap target
                        androidx.compose.material3.Switch(
                            checked = isEnabled,
                            onCheckedChange = onEnabledChange
                        )
                    }

                    Text(
                        text = "Runs across apps with wake word support.",
                        style = MaterialTheme.typography.labelSmall.copy(fontSize = 11.sp),
                        color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f),
                        modifier = Modifier.padding(top = DashboardLayoutTokens.Spacing4)
                    )

                    if (isEnabled) {
                        Spacer(modifier = Modifier.height(DashboardLayoutTokens.Spacing12))

                        // Stack vertically on phones for better readability
                        Column(
                            verticalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing8)
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
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
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(
                                            MaterialTheme.colorScheme.error.copy(alpha = 0.05f),
                                            RoundedCornerShape(DashboardLayoutTokens.Spacing4)
                                        )
                                        .padding(horizontal = 8.dp, vertical = 6.dp),
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
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
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
    }
}

@Composable
private fun FeatureGrid(
    isBackgroundModeEnabled: Boolean,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    onShowDisabledMessage: () -> Unit,
    layoutSpec: DashboardLayoutSpec
) {
    when (layoutSpec.featureColumns) {
        1 -> {
            // Single column for phones
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
        2 -> {
            // Two columns for tablets and teralux
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
private fun AiProviderCard(
    selectedProvider: String?,
    isSaving: Boolean,
    onProviderSelected: (String?) -> Unit,
    layoutSpec: DashboardLayoutSpec
) {
    // User-selectable providers only (excludes 'local' which is fallback-only)
    val providers = listOf("gemini", "openai", "groq", "orion")
    val providerLabels = mapOf(
        "gemini" to "Gemini",
        "openai" to "OpenAI",
        "groq" to "Groq",
        "orion" to "Orion"
    )

    androidx.compose.material3.Surface(
        modifier = Modifier.fillMaxWidth(),
        color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.08f),
        shape = RoundedCornerShape(DashboardLayoutTokens.CardRadiusPrimary),
        border = androidx.compose.foundation.BorderStroke(
            1.dp,
            MaterialTheme.colorScheme.primary.copy(alpha = 0.1f)
        )
    ) {
        Column(modifier = Modifier.padding(DashboardLayoutTokens.Spacing16)) {
            // Header row
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Outlined.AutoAwesome,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.size(20.dp)
                    )
                    Spacer(modifier = Modifier.width(DashboardLayoutTokens.Spacing8))
                    Text(
                        text = "AI Engine",
                        style = MaterialTheme.typography.titleSmall,
                        color = MaterialTheme.colorScheme.onSurface,
                        fontWeight = FontWeight.ExtraBold
                    )
                }
                if (isSaving) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(16.dp),
                        strokeWidth = 2.dp,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }

            // Helper text - explicitly mentions all three features
            Text(
                text = "Used for Meeting Transcriber, AI Assistant, and Background Assistant.",
                style = MaterialTheme.typography.labelSmall.copy(fontSize = 11.sp),
                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f),
                modifier = Modifier.padding(top = DashboardLayoutTokens.Spacing4)
            )

            // Provider selector - stretched chips for balanced width distribution
            when (layoutSpec.providerSelectorMode) {
                ProviderSelectorMode.SINGLE_ROW -> {
                    // Stretched row layout for wider screens
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = DashboardLayoutTokens.Spacing12),
                        horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing8)
                    ) {
                        providers.forEach { provider ->
                            FilterChip(
                                selected = selectedProvider == provider,
                                onClick = {
                                    if (!isSaving) {
                                        onProviderSelected(provider)
                                    }
                                },
                                label = {
                                    Text(
                                        text = providerLabels[provider] ?: provider,
                                        style = MaterialTheme.typography.labelSmall.copy(fontSize = 11.sp),
                                        fontWeight = FontWeight.Medium,
                                        modifier = Modifier.fillMaxWidth()
                                    )
                                },
                                leadingIcon = if (selectedProvider == provider) {
                                    {
                                        Icon(
                                            imageVector = Icons.Filled.Check,
                                            contentDescription = null,
                                            modifier = Modifier.size(FilterChipDefaults.IconSize),
                                            tint = MaterialTheme.colorScheme.primary
                                        )
                                    }
                                } else {
                                    null
                                },
                                enabled = !isSaving,
                                modifier = Modifier.weight(1f)
                            )
                        }
                    }
                }
                ProviderSelectorMode.WRAPPED_ROWS -> {
                    // 2-column stretched grid for phone/narrow screens
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = DashboardLayoutTokens.Spacing12),
                        verticalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing8)
                    ) {
                        // Split providers into rows of 2
                        providers.chunked(2).forEach { rowProviders ->
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.spacedBy(DashboardLayoutTokens.Spacing8)
                            ) {
                                rowProviders.forEach { provider ->
                                    FilterChip(
                                        selected = selectedProvider == provider,
                                        onClick = {
                                            if (!isSaving) {
                                                onProviderSelected(provider)
                                            }
                                        },
                                        label = {
                                            Text(
                                                text = providerLabels[provider] ?: provider,
                                                style = MaterialTheme.typography.labelSmall.copy(fontSize = 11.sp),
                                                fontWeight = FontWeight.Medium,
                                                modifier = Modifier.fillMaxWidth()
                                            )
                                        },
                                        leadingIcon = if (selectedProvider == provider) {
                                            {
                                                Icon(
                                                    imageVector = Icons.Filled.Check,
                                                    contentDescription = null,
                                                    modifier = Modifier.size(FilterChipDefaults.IconSize),
                                                    tint = MaterialTheme.colorScheme.primary
                                                )
                                            }
                                        } else {
                                            null
                                        },
                                        enabled = !isSaving,
                                        modifier = Modifier.weight(1f)
                                    )
                                }
                                // Add spacer if odd number of providers in last row
                                if (rowProviders.size < 2) {
                                    Spacer(modifier = Modifier.weight(1f))
                                }
                            }
                        }
                    }
                }
            }

            // Clear selection option
            if (selectedProvider != null) {
                Text(
                    text = "Reset to system default",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.primary,
                    modifier = Modifier
                        .padding(top = DashboardLayoutTokens.Spacing8)
                        .clickable {
                            if (!isSaving) {
                                onProviderSelected(null)
                            }
                        }
                )
            }
        }
    }
}

/**
 * Main dashboard content container that branches to device-specific layouts.
 */
@Composable
fun DashboardContent(
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    isBackgroundModeEnabled: Boolean,
    isOverlayPermissionGranted: Boolean,
    onBackgroundModeChange: (Boolean) -> Unit,
    onRequestOverlayPermission: () -> Unit,
    onShowDisabledMessage: () -> Unit,
    aiProvider: String? = null,
    isSavingAiProvider: Boolean = false,
    onAiProviderChange: (String?) -> Unit = {}
) {
    val layoutSpec = rememberDashboardLayoutSpec()

    when (layoutSpec.deviceClass) {
        DashboardDeviceClass.PHONE -> DashboardPhoneLayout(
            layoutSpec = layoutSpec,
            onNavigateToStreaming = onNavigateToStreaming,
            onNavigateToEdge = onNavigateToEdge,
            isBackgroundModeEnabled = isBackgroundModeEnabled,
            isOverlayPermissionGranted = isOverlayPermissionGranted,
            onBackgroundModeChange = onBackgroundModeChange,
            onRequestOverlayPermission = onRequestOverlayPermission,
            onShowDisabledMessage = onShowDisabledMessage,
            aiProvider = aiProvider,
            isSavingAiProvider = isSavingAiProvider,
            onAiProviderChange = onAiProviderChange
        )
        DashboardDeviceClass.TABLET -> DashboardTabletLayout(
            layoutSpec = layoutSpec,
            onNavigateToStreaming = onNavigateToStreaming,
            onNavigateToEdge = onNavigateToEdge,
            isBackgroundModeEnabled = isBackgroundModeEnabled,
            isOverlayPermissionGranted = isOverlayPermissionGranted,
            onBackgroundModeChange = onBackgroundModeChange,
            onRequestOverlayPermission = onRequestOverlayPermission,
            onShowDisabledMessage = onShowDisabledMessage,
            aiProvider = aiProvider,
            isSavingAiProvider = isSavingAiProvider,
            onAiProviderChange = onAiProviderChange
        )
        DashboardDeviceClass.TERALUX -> DashboardTeraluxLayout(
            layoutSpec = layoutSpec,
            onNavigateToStreaming = onNavigateToStreaming,
            onNavigateToEdge = onNavigateToEdge,
            isBackgroundModeEnabled = isBackgroundModeEnabled,
            isOverlayPermissionGranted = isOverlayPermissionGranted,
            onBackgroundModeChange = onBackgroundModeChange,
            onRequestOverlayPermission = onRequestOverlayPermission,
            onShowDisabledMessage = onShowDisabledMessage,
            aiProvider = aiProvider,
            isSavingAiProvider = isSavingAiProvider,
            onAiProviderChange = onAiProviderChange
        )
    }
}

/**
 * Phone layout: single-column vertical flow, always scrollable, compact padding.
 */
@Composable
private fun DashboardPhoneLayout(
    layoutSpec: DashboardLayoutSpec,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    isBackgroundModeEnabled: Boolean,
    isOverlayPermissionGranted: Boolean,
    onBackgroundModeChange: (Boolean) -> Unit,
    onRequestOverlayPermission: () -> Unit,
    onShowDisabledMessage: () -> Unit,
    aiProvider: String?,
    isSavingAiProvider: Boolean,
    onAiProviderChange: (String?) -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = layoutSpec.outerHorizontalPadding)
            .widthIn(max = layoutSpec.contentMaxWidth),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(layoutSpec.sectionSpacing)
    ) {
        DashboardHeader()

        BackgroundAssistantCard(
            isEnabled = isBackgroundModeEnabled,
            isOverlayPermissionGranted = isOverlayPermissionGranted,
            onEnabledChange = onBackgroundModeChange,
            onRequestOverlayPermission = onRequestOverlayPermission,
            layoutSpec = layoutSpec
        )

        AiProviderCard(
            selectedProvider = aiProvider,
            isSaving = isSavingAiProvider,
            onProviderSelected = onAiProviderChange,
            layoutSpec = layoutSpec
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
                onShowDisabledMessage = onShowDisabledMessage,
                layoutSpec = layoutSpec
            )
        }

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

/**
 * Tablet layout: medium-wide content, 2-column features, balanced spacing.
 */
@Composable
private fun DashboardTabletLayout(
    layoutSpec: DashboardLayoutSpec,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    isBackgroundModeEnabled: Boolean,
    isOverlayPermissionGranted: Boolean,
    onBackgroundModeChange: (Boolean) -> Unit,
    onRequestOverlayPermission: () -> Unit,
    onShowDisabledMessage: () -> Unit,
    aiProvider: String?,
    isSavingAiProvider: Boolean,
    onAiProviderChange: (String?) -> Unit
) {
    val contentModifier =
        if (layoutSpec.isScrollable) {
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = layoutSpec.outerHorizontalPadding)
        } else {
            Modifier.fillMaxSize().padding(horizontal = layoutSpec.outerHorizontalPadding)
        }

    Column(
        modifier = contentModifier.widthIn(max = layoutSpec.contentMaxWidth),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(layoutSpec.sectionSpacing)
    ) {
        DashboardHeader()

        BackgroundAssistantCard(
            isEnabled = isBackgroundModeEnabled,
            isOverlayPermissionGranted = isOverlayPermissionGranted,
            onEnabledChange = onBackgroundModeChange,
            onRequestOverlayPermission = onRequestOverlayPermission,
            layoutSpec = layoutSpec
        )

        AiProviderCard(
            selectedProvider = aiProvider,
            isSaving = isSavingAiProvider,
            onProviderSelected = onAiProviderChange,
            layoutSpec = layoutSpec
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
                onShowDisabledMessage = onShowDisabledMessage,
                layoutSpec = layoutSpec
            )
        }

        // Footer spacer: only effective in non-scrollable layouts
        if (!layoutSpec.isScrollable) {
            Spacer(modifier = Modifier.weight(1f))
        }

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

/**
 * Teralux layout: landscape-first, two-panel composition with control content on left
 * and workspace features on right. Falls back to stacked layout for portrait/narrow.
 */
@Composable
private fun DashboardTeraluxLayout(
    layoutSpec: DashboardLayoutSpec,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    isBackgroundModeEnabled: Boolean,
    isOverlayPermissionGranted: Boolean,
    onBackgroundModeChange: (Boolean) -> Unit,
    onRequestOverlayPermission: () -> Unit,
    onShowDisabledMessage: () -> Unit,
    aiProvider: String?,
    isSavingAiProvider: Boolean,
    onAiProviderChange: (String?) -> Unit
) {
    // Fallback to stacked layout for portrait or narrow width
    if (!layoutSpec.isLandscape || layoutSpec.availableWidth < 800.dp) {
        DashboardTabletLayout(
            layoutSpec = layoutSpec,
            onNavigateToStreaming = onNavigateToStreaming,
            onNavigateToEdge = onNavigateToEdge,
            isBackgroundModeEnabled = isBackgroundModeEnabled,
            isOverlayPermissionGranted = isOverlayPermissionGranted,
            onBackgroundModeChange = onBackgroundModeChange,
            onRequestOverlayPermission = onRequestOverlayPermission,
            onShowDisabledMessage = onShowDisabledMessage,
            aiProvider = aiProvider,
            isSavingAiProvider = isSavingAiProvider,
            onAiProviderChange = onAiProviderChange
        )
        return
    }

    // Two-panel landscape layout
    val contentModifier =
        if (layoutSpec.isScrollable) {
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = layoutSpec.outerHorizontalPadding)
        } else {
            Modifier.fillMaxSize().padding(horizontal = layoutSpec.outerHorizontalPadding)
        }

    Row(
        modifier = contentModifier.widthIn(max = layoutSpec.contentMaxWidth),
        horizontalArrangement = Arrangement.spacedBy(layoutSpec.sectionSpacing)
    ) {
        // Left panel: control-oriented content
        Column(
            modifier = Modifier.weight(1f),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(layoutSpec.sectionSpacing)
        ) {
            DashboardHeader()

            BackgroundAssistantCard(
                isEnabled = isBackgroundModeEnabled,
                isOverlayPermissionGranted = isOverlayPermissionGranted,
                onEnabledChange = onBackgroundModeChange,
                onRequestOverlayPermission = onRequestOverlayPermission,
                layoutSpec = layoutSpec
            )

            AiProviderCard(
                selectedProvider = aiProvider,
                isSaving = isSavingAiProvider,
                onProviderSelected = onAiProviderChange,
                layoutSpec = layoutSpec
            )

            // Footer spacer: only effective in non-scrollable layouts
            if (!layoutSpec.isScrollable) {
                Spacer(modifier = Modifier.weight(1f))
            }

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

        // Right panel: workspace features
        Column(
            modifier = Modifier.weight(1f),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
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
                    onShowDisabledMessage = onShowDisabledMessage,
                    layoutSpec = layoutSpec
                )
            }
        }
    }
}

// region Previews

@Composable
private fun DashboardPreviewWrapper(
    content: @Composable () -> Unit
) {
    androidx.compose.material3.MaterialTheme(
        colorScheme = androidx.compose.material3.lightColorScheme()
    ) {
        androidx.compose.material3.Surface(
            color = androidx.compose.material3.MaterialTheme.colorScheme.background,
            modifier = Modifier.fillMaxSize()
        ) {
            content()
        }
    }
}

@androidx.compose.ui.tooling.preview.Preview(name = "Phone Portrait", widthDp = 360, heightDp = 740)
@Composable
private fun PreviewDashboardPhone() {
    DashboardPreviewWrapper {
        DashboardPhoneLayout(
            layoutSpec = DashboardLayoutSpec(
                deviceClass = DashboardDeviceClass.PHONE,
                isScrollable = true,
                contentMaxWidth = 360.dp,
                outerHorizontalPadding = 20.dp,
                sectionSpacing = 24.dp,
                featureColumns = 1,
                providerSelectorMode = ProviderSelectorMode.WRAPPED_ROWS,
                backgroundAssistantMode = BackgroundAssistantLayoutMode.SPLIT_ROWS,
                isLandscape = false,
                availableWidth = 360.dp
            ),
            onNavigateToStreaming = {},
            onNavigateToEdge = {},
            isBackgroundModeEnabled = false,
            isOverlayPermissionGranted = true,
            onBackgroundModeChange = {},
            onRequestOverlayPermission = {},
            onShowDisabledMessage = {},
            aiProvider = "gemini",
            isSavingAiProvider = false,
            onAiProviderChange = {}
        )
    }
}

@androidx.compose.ui.tooling.preview.Preview(name = "Tablet Portrait", widthDp = 768, heightDp = 1024)
@Composable
private fun PreviewDashboardTablet() {
    DashboardPreviewWrapper {
        DashboardTabletLayout(
            layoutSpec = DashboardLayoutSpec(
                deviceClass = DashboardDeviceClass.TABLET,
                isScrollable = false,
                contentMaxWidth = 960.dp,
                outerHorizontalPadding = 32.dp,
                sectionSpacing = 28.dp,
                featureColumns = 2,
                providerSelectorMode = ProviderSelectorMode.SINGLE_ROW,
                backgroundAssistantMode = BackgroundAssistantLayoutMode.INLINE,
                isLandscape = false,
                availableWidth = 768.dp
            ),
            onNavigateToStreaming = {},
            onNavigateToEdge = {},
            isBackgroundModeEnabled = true,
            isOverlayPermissionGranted = true,
            onBackgroundModeChange = {},
            onRequestOverlayPermission = {},
            onShowDisabledMessage = {},
            aiProvider = "openai",
            isSavingAiProvider = false,
            onAiProviderChange = {}
        )
    }
}

@androidx.compose.ui.tooling.preview.Preview(name = "Teralux Landscape", widthDp = 1280, heightDp = 720)
@Composable
private fun PreviewDashboardTeralux() {
    DashboardPreviewWrapper {
        DashboardTeraluxLayout(
            layoutSpec = DashboardLayoutSpec(
                deviceClass = DashboardDeviceClass.TERALUX,
                isScrollable = false,
                contentMaxWidth = 1200.dp,
                outerHorizontalPadding = 48.dp,
                sectionSpacing = 32.dp,
                featureColumns = 2,
                providerSelectorMode = ProviderSelectorMode.SINGLE_ROW,
                backgroundAssistantMode = BackgroundAssistantLayoutMode.INLINE,
                isLandscape = true,
                availableWidth = 1280.dp
            ),
            onNavigateToStreaming = {},
            onNavigateToEdge = {},
            isBackgroundModeEnabled = true,
            isOverlayPermissionGranted = true,
            onBackgroundModeChange = {},
            onRequestOverlayPermission = {},
            onShowDisabledMessage = {},
            aiProvider = "gemini",
            isSavingAiProvider = false,
            onAiProviderChange = {}
        )
    }
}

@androidx.compose.ui.tooling.preview.Preview(name = "Phone - Provider Chips Wrap", widthDp = 320, heightDp = 640)
@Composable
private fun PreviewDashboardPhoneNarrow() {
    DashboardPreviewWrapper {
        DashboardPhoneLayout(
            layoutSpec = DashboardLayoutSpec(
                deviceClass = DashboardDeviceClass.PHONE,
                isScrollable = true,
                contentMaxWidth = 320.dp,
                outerHorizontalPadding = 20.dp,
                sectionSpacing = 24.dp,
                featureColumns = 1,
                providerSelectorMode = ProviderSelectorMode.WRAPPED_ROWS,
                backgroundAssistantMode = BackgroundAssistantLayoutMode.SPLIT_ROWS,
                isLandscape = false,
                availableWidth = 320.dp
            ),
            onNavigateToStreaming = {},
            onNavigateToEdge = {},
            isBackgroundModeEnabled = false,
            isOverlayPermissionGranted = false,
            onBackgroundModeChange = {},
            onRequestOverlayPermission = {},
            onShowDisabledMessage = {},
            aiProvider = null,
            isSavingAiProvider = false,
            onAiProviderChange = {}
        )
    }
}

// endregion
