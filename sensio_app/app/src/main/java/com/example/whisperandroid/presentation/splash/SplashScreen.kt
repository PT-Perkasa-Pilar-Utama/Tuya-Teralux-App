package com.example.whisperandroid.presentation.splash

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Box
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material3.MaterialTheme
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
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController
import kotlinx.coroutines.delay
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.components.SensioButton
import com.example.whisperandroid.ui.theme.SensioSpacing
import com.example.whisperandroid.ui.theme.SensioTypography

@Composable
private fun mapToEmpatheticMessage(rawMessage: String): String {
    val lower = rawMessage.lowercase()
    return when {
        lower.contains("530") || lower.contains("server error") -> "Layanan sedang sibuk, coba sebentar lagi"
        lower.contains("timeout") || lower.contains("timed out") -> "Waktu tunggu habis, periksa koneksi internet"
        lower.contains("network") || lower.contains("connection") ||
            lower.contains("socket") || lower.contains("mqtt") -> "Koneksi internet bermasalah"
        lower.contains("unauthorized") || lower.contains("session") ||
            lower.contains("401") -> "Sesi berakhir, silakan masuk lagi"
        lower.contains("no internet") || lower.contains("no_internet") -> "Tidak ada koneksi internet"
        lower.contains("failed") || lower.contains("error") -> "Terjadi kesalahan, coba lagi"
        else -> rawMessage
    }
}

@Composable
fun SplashScreen(
    viewModel: SplashViewModel,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
    navController: NavController? = null
) {
    val uiState by viewModel.uiState.collectAsState()
    var animationStarted by remember { mutableStateOf(false) }

    val alphaAnimation by animateFloatAsState(
        targetValue = if (animationStarted) 1f else 0f,
        animationSpec = tween(durationMillis = 1000),
        label = "fadeIn"
    )

    LaunchedEffect(Unit) {
        animationStarted = true
    }

    LaunchedEffect(uiState) {
        delay(1000)

        when (uiState) {
            is SplashUiState.Authenticated -> {
                navController?.navigate(AppRoutes.Dashboard.route) {
                    popUpTo(AppRoutes.Splash.route) { inclusive = true }
                }
            }
            is SplashUiState.NotRegistered -> {
                navController?.navigate(AppRoutes.Register.route) {
                    popUpTo(AppRoutes.Splash.route) { inclusive = true }
                }
            }
            is SplashUiState.Unauthorized -> {
                navController?.navigate(AppRoutes.Register.route) {
                    popUpTo(AppRoutes.Splash.route) { inclusive = true }
                }
            }
            else -> {}
        }
    }

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
        contentAlignment = Alignment.Center
    ) {
        when (val state = uiState) {
            is SplashUiState.Error -> {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Column(
                        modifier = Modifier.weight(1f),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center
                    ) {
                        Text(
                            text = mapToEmpatheticMessage(state.message),
                            color = MaterialTheme.colorScheme.primary,
                            fontSize = SensioTypography.Body,
                            textAlign = TextAlign.Center,
                            modifier = Modifier.padding(horizontal = SensioSpacing.Lg)
                        )
                        Spacer(modifier = Modifier.height(SensioSpacing.Md))
                        SensioButton(
                            text = "Coba Lagi",
                            onClick = onRetry,
                            modifier = Modifier
                                .fillMaxWidth()
                                .widthIn(max = 360.dp)
                                .height(48.dp)
                        )
                    }
                }
            }
            else -> {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(SensioSpacing.Md),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "Sensio",
                        fontSize = SensioTypography.HeadlineMobile,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.primary,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.graphicsLayer { this.alpha = alphaAnimation }
                    )
                    CircularProgressIndicator(
                        modifier = Modifier.graphicsLayer { this.alpha = alphaAnimation },
                        color = MaterialTheme.colorScheme.primary,
                        strokeWidth = 2.dp
                    )
                }
            }
        }
    }
}
