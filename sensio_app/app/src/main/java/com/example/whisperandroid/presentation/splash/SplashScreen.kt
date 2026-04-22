package com.example.whisperandroid.presentation.splash

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.components.SensioButton
import com.example.whisperandroid.presentation.components.SensioLogo
import com.example.whisperandroid.ui.theme.SensioSpacing
import com.example.whisperandroid.ui.theme.SensioTypography

@Composable
fun SplashScreen(
    viewModel: SplashViewModel,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
    navController: NavController? = null
) {
    val uiState by viewModel.uiState.collectAsState()

    LaunchedEffect(uiState) {
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
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            AnimatedVisibility(
                visible = true,
                enter = fadeIn(animationSpec = androidx.compose.animation.core.tween(500))
            ) {
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    SensioLogo()
                    Spacer(modifier = Modifier.height(SensioSpacing.Md))
                    Text(
                        text = "Sensio",
                        fontSize = SensioTypography.HeadlineMobile,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
                        textAlign = TextAlign.Center
                    )
                }
            }

            Spacer(modifier = Modifier.height(SensioSpacing.Xxl))

            val state = uiState
            when (state) {
                is SplashUiState.Loading -> {
                    CircularProgressIndicator(
                        modifier = Modifier.size(48.dp),
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                is SplashUiState.Error -> {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(SensioSpacing.Md)
                    ) {
                        Text(
                            text = state.message,
                            color = MaterialTheme.colorScheme.error,
                            fontSize = SensioTypography.Body,
                            textAlign = TextAlign.Center,
                            modifier = Modifier.padding(horizontal = SensioSpacing.Lg)
                        )
                        SensioButton(
                            text = "Retry",
                            onClick = onRetry,
                            modifier = Modifier.height(48.dp)
                        )
                    }
                }
                else -> {}
            }
        }
    }
}