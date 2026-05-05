package com.example.whisperandroid.presentation.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import android.content.Intent
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.presentation.components.SensioButton
import com.example.whisperandroid.service.mqtt.MqttForegroundService
import com.example.whisperandroid.presentation.components.SensioLogo
import com.example.whisperandroid.ui.theme.SensioBorder
import com.example.whisperandroid.ui.theme.SensioElevation
import com.example.whisperandroid.ui.theme.SensioRadius
import com.example.whisperandroid.ui.theme.SensioSpacing
import com.example.whisperandroid.ui.theme.SensioTypography

@Composable
fun AuthScreen(
    onNavigateToRegister: () -> Unit,
    onNavigateToDashboard: () -> Unit
) {
    val context = LocalContext.current
    val application = context.applicationContext as android.app.Application
    val viewModel: AuthViewModel = viewModel {
        AuthViewModel(
            application,
            NetworkModule.terminalRepository
        )
    }
    val uiState by viewModel.uiState.collectAsState()

    LaunchedEffect(Unit) {
        viewModel.navigationEvent.collect { isRegistered ->
            if (isRegistered) {
                val serviceIntent = Intent(context, MqttForegroundService::class.java)
                context.startForegroundService(serviceIntent)
                onNavigateToDashboard()
            } else {
                onNavigateToRegister()
            }
        }
    }

    // Background gradient
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
        modifier = Modifier
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
                    modifier = Modifier
                        .padding(SensioSpacing.Xl)
                        .fillMaxWidth(0.85f),
                    horizontalArrangement = Arrangement.SpaceEvenly,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(
                        modifier = Modifier.weight(1f),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center
                    ) {
                        AuthLogo()
                        Spacer(modifier = Modifier.height(SensioSpacing.Xxl))
                        AuthHeadlineTablet()
                    }

                    AuthCard(
                        isLoading = uiState.isLoading,
                        error = uiState.error,
                        onRetry = { viewModel.checkMacRegistration() },
                        modifier = Modifier.weight(0.8f)
                    )
                }
            } else {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(
                            top = SensioSpacing.Lg,
                            start = SensioSpacing.Lg,
                            end = SensioSpacing.Lg,
                            bottom = SensioSpacing.Xl
                        ),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    AuthLogo()
                    Spacer(modifier = Modifier.height(SensioSpacing.Xxl))
                    AuthHeadlineMobile()
                    Spacer(modifier = Modifier.height(SensioSpacing.Xxl))
                    AuthCard(
                        isLoading = uiState.isLoading,
                        error = uiState.error,
                        onRetry = { viewModel.checkMacRegistration() }
                    )
                }
            }
        }
    }
}

@Composable
private fun AuthLogo() {
    Box(
        modifier = Modifier.testTag("auth_logo")
    ) {
        SensioLogo()
    }
}

@Composable
private fun AuthHeadlineMobile() {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.fillMaxWidth()
    ) {
        Text(
            text = "Welcome",
            fontSize = SensioTypography.HeadlineMobile,
            lineHeight = 32.sp,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.onBackground,
            textAlign = TextAlign.Center,
            style = MaterialTheme.typography.headlineSmall
        )
        Spacer(modifier = Modifier.height(SensioSpacing.Sm))
        Text(
            text = "Checking device registration...",
            fontSize = SensioTypography.Body,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            fontWeight = FontWeight.Medium,
            textAlign = TextAlign.Center
        )
    }
}

@Composable
private fun AuthHeadlineTablet() {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.fillMaxWidth()
    ) {
        Text(
            text = "Welcome",
            fontSize = SensioTypography.HeadlineTablet,
            lineHeight = 44.sp,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.onBackground,
            style = MaterialTheme.typography.displaySmall
        )
        Spacer(modifier = Modifier.height(SensioSpacing.Md))
        Text(
            text = "Checking device registration...",
            fontSize = SensioTypography.Subtitle,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            lineHeight = 24.sp,
            fontWeight = FontWeight.Medium
        )
    }
}

@Composable
private fun AuthCard(
    isLoading: Boolean,
    error: String?,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier
            .widthIn(max = 400.dp)
            .fillMaxSize()
            .padding(SensioSpacing.Sm),
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
            Spacer(modifier = Modifier.height(SensioSpacing.Xxl))

            when {
                isLoading -> {
                    CircularProgressIndicator(
                        modifier = Modifier.size(48.dp),
                        color = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.height(SensioSpacing.Md))
                    Text(
                        text = "Verifying device...",
                        fontSize = SensioTypography.Body,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
                        textAlign = TextAlign.Center
                    )
                }
                error != null -> {
                    Text(
                        text = error,
                        fontSize = SensioTypography.Body,
                        color = MaterialTheme.colorScheme.error.copy(alpha = 0.8f),
                        textAlign = TextAlign.Center
                    )
                    Spacer(modifier = Modifier.height(SensioSpacing.Md))
                    SensioButton(
                        text = "Retry",
                        onClick = onRetry
                    )
                }
                else -> {
                    CircularProgressIndicator(
                        modifier = Modifier.size(48.dp),
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }

            Spacer(modifier = Modifier.weight(1f))
        }
    }
}