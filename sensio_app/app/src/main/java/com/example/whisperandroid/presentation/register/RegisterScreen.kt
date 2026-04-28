package com.example.whisperandroid.presentation.register

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.wrapContentHeight
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.presentation.components.SensioButton
import com.example.whisperandroid.presentation.components.SensioLogo
import com.example.whisperandroid.presentation.components.SensioTextField
import com.example.whisperandroid.presentation.components.ToastObserver
import com.example.whisperandroid.ui.theme.SensioBorder
import com.example.whisperandroid.ui.theme.SensioElevation
import com.example.whisperandroid.ui.theme.SensioRadius
import com.example.whisperandroid.ui.theme.SensioSpacing
import com.example.whisperandroid.ui.theme.SensioTypography

@Composable
fun RegisterScreen(onNavigateToDashboard: () -> Unit) {
    val context = LocalContext.current
    val application = context.applicationContext as android.app.Application
    val viewModel: RegisterViewModel =
        viewModel {
            RegisterViewModel(
                application,
                NetworkModule.registerUseCase,
                NetworkModule.authenticateUseCase
            )
        }
    val uiState by viewModel.uiState.collectAsState()
    var name by remember { mutableStateOf("") }
    var roomId by remember { mutableStateOf("") }

    // Permission Launcher
    val permissionLauncher =
        rememberLauncherForActivityResult(
            contract = ActivityResultContracts.RequestMultiplePermissions()
        ) { permissions ->
            val recordGranted = permissions[Manifest.permission.RECORD_AUDIO] ?: false
            val storageGranted = permissions[Manifest.permission.WRITE_EXTERNAL_STORAGE] ?: false
        }

    // Proactive Request on Launch
    LaunchedEffect(Unit) {
        val permissionsToRequest = mutableListOf<String>()
        val hasRecordAuth = ContextCompat.checkSelfPermission(
            context,
            Manifest.permission.RECORD_AUDIO
        ) != PackageManager.PERMISSION_GRANTED
        if (hasRecordAuth) {
            permissionsToRequest.add(Manifest.permission.RECORD_AUDIO)
        }
        val hasStorageAuth = ContextCompat.checkSelfPermission(
            context,
            Manifest.permission.WRITE_EXTERNAL_STORAGE
        ) != PackageManager.PERMISSION_GRANTED
        if (hasStorageAuth) {
            permissionsToRequest.add(Manifest.permission.WRITE_EXTERNAL_STORAGE)
        }

        if (permissionsToRequest.isNotEmpty()) {
            permissionLauncher.launch(permissionsToRequest.toTypedArray())
        }
    }

    // Reusable Toast Observer
    ToastObserver(
        message = uiState.message,
        onToastShown = { viewModel.clearMessage() }
    )

    // Side effect for navigation on success
    LaunchedEffect(uiState.isSuccess) {
        if (uiState.isSuccess) {
            onNavigateToDashboard()
        }
    }

    // Enhanced background gradient with subtle radial design
    val bgGradient =
        Brush.radialGradient(
            colors =
            listOf(
                MaterialTheme.colorScheme.primary.copy(alpha = 0.06f),
                MaterialTheme.colorScheme.surface.copy(alpha = 0.5f),
                MaterialTheme.colorScheme.background
            ),
            center = Offset(800f, 400f),
            radius = 1200f
        )

    Box(
        modifier =
        Modifier
            .fillMaxSize()
            .background(bgGradient)
    ) {
        BoxWithConstraints(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            val isTablet = maxWidth > 600.dp

            if (isTablet) {
                Row(
                    modifier =
                    Modifier
                        .padding(SensioSpacing.Xl)
                        .fillMaxWidth(0.85f),
                    horizontalArrangement = Arrangement.SpaceEvenly,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Left Side: Branding & Microcopy
                    Column(
                        modifier =
                        Modifier
                            .weight(1f)
                            .padding(end = SensioSpacing.Xxl),
                        horizontalAlignment = Alignment.Start,
                        verticalArrangement = Arrangement.Center
                    ) {
                        // Logo is clickable
                        Box(
                            modifier = Modifier
                                .testTag("register_logo")
                                .clickable { }
                        ) {
                            SensioLogo()
                        }

                        // Text is NOT clickable
                        Spacer(modifier = Modifier.height(SensioSpacing.Xxl))
                        Text(
                            text = "Secure Your\nConversation",
                            fontSize = SensioTypography.HeadlineTablet,
                            lineHeight = 44.sp,
                            fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.onBackground,
                            style = MaterialTheme.typography.displaySmall
                        )
                        Spacer(modifier = Modifier.height(SensioSpacing.Md))
                        Text(
                            text = "Private, high-quality transcription for enterprise teams.",
                            fontSize = SensioTypography.Subtitle,
                            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
                            lineHeight = 24.sp,
                            fontWeight = FontWeight.Medium
                        )
                    }

                    // Right Side: Card Form
                    RegisterCard(
                        name = name,
                        onNameChange = {
                            name = it
                            viewModel.clearError()
                        },
                        roomId = roomId,
                        onRoomIdChange = {
                            roomId = it
                            viewModel.clearError()
                        },
                        isLoading = uiState.isLoading,
                        error = uiState.error,
                        onRegisterClick = { viewModel.register(name, roomId) },
                        modifier = Modifier.weight(0.8f)
                    )
                }
            } else {
                // Mobile Layout (Vertical Stack)
                Column(
                    modifier =
                    Modifier
                        .fillMaxSize()
                        .padding(top = SensioSpacing.Lg, start = SensioSpacing.Lg, end = SensioSpacing.Lg, bottom = SensioSpacing.Xl)
                        .padding(WindowInsets.statusBars.asPaddingValues()),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    // Wrapper for Logo (Clickable)
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier
                            .testTag("register_logo")
                            .clickable { }
                    ) {
                        SensioLogo()
                    }

                    Spacer(modifier = Modifier.height(SensioSpacing.Xl))

                    // Text (Not Clickable) - Improved headline for mobile
                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "Secure Your Conversation",
                            fontSize = SensioTypography.HeadlineMobile,
                            lineHeight = 32.sp,
                            fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.onBackground,
                            textAlign = TextAlign.Center,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis,
                            style = MaterialTheme.typography.headlineSmall
                        )
                        Spacer(modifier = Modifier.height(SensioSpacing.Sm))
                        Text(
                            text = "Private meeting transcription",
                            fontSize = SensioTypography.Body,
                            lineHeight = 22.sp,
                            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
                            fontWeight = FontWeight.Medium,
                            textAlign = TextAlign.Center
                        )
                    }
                    Spacer(modifier = Modifier.height(SensioSpacing.Xxl))

                    RegisterCard(
                        name = name,
                        onNameChange = {
                            name = it
                            viewModel.clearError()
                        },
                        roomId = roomId,
                        onRoomIdChange = {
                            roomId = it
                            viewModel.clearError()
                        },
                        isLoading = uiState.isLoading,
                        error = uiState.error,
                        onRegisterClick = { viewModel.register(name, roomId) }
                    )
                }
            }
        }
    }
}

@Composable
fun RegisterCard(
    name: String,
    onNameChange: (String) -> Unit,
    roomId: String,
    onRoomIdChange: (String) -> Unit,
    isLoading: Boolean,
    error: String?,
    onRegisterClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier =
        modifier
            .widthIn(max = 400.dp)
            .fillMaxWidth()
            .padding(SensioSpacing.Sm)
            .wrapContentHeight(),
        shape = RoundedCornerShape(SensioRadius.Xxl),
        color = MaterialTheme.colorScheme.surface.copy(alpha = 0.9f),
        shadowElevation = SensioElevation.Md,
        border =
        androidx.compose.foundation.BorderStroke(
            width = SensioBorder.Md,
            color = MaterialTheme.colorScheme.primary.copy(alpha = 0.1f)
        )
    ) {
        Column(
            modifier = Modifier.padding(SensioSpacing.Lg),
            verticalArrangement = Arrangement.spacedBy(SensioSpacing.Md),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            SensioTextField(
                value = name,
                onValueChange = onNameChange,
                label = "Name"
            )

            SensioTextField(
                value = roomId,
                onValueChange = onRoomIdChange,
                label = "Room ID",
                isRoomId = true
            )

            if (isLoading) {
                CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
            } else {
                SensioButton(
                    text = "Register",
                    onClick = onRegisterClick,
                    enabled = name.isNotBlank() && roomId.isNotBlank()
                )
            }

            if (error != null) {
                Text(
                    text = error,
                    color = MaterialTheme.colorScheme.error.copy(alpha = 0.8f),
                    fontSize = SensioTypography.Caption,
                    textAlign = TextAlign.Center
                )
            }
        }
    }
}
