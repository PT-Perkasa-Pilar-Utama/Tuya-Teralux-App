package com.example.whisperandroid.presentation.splash

import androidx.compose.animation.core.EaseInOut
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding

import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController
import com.example.whisperandroid.R
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.components.SensioButton
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

    val logoScale by animateFloatAsState(
        targetValue = 1f,
        animationSpec = tween(
            durationMillis = 800,
            easing = EaseInOut
        ),
        label = "logoScale"
    )

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
        modifier = modifier.fillMaxSize()
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            drawRect(
                brush = Brush.verticalGradient(
                    colors = listOf(
                        Color(0xFF67E8F9),
                        Color(0xFF0EA5E9),
                        Color(0xFF075985)
                    )
                )
            )
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = SensioSpacing.Lg),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Box(
                modifier = Modifier.graphicsLayer {
                    scaleX = logoScale
                    scaleY = logoScale
                }
            ) {
                androidx.compose.foundation.Image(
                    painter = painterResource(id = R.drawable.ic_launcher_foreground),
                    contentDescription = "Sensio Logo",
                    modifier = Modifier.height(108.dp)
                )
            }

            Spacer(modifier = Modifier.height(SensioSpacing.Lg))

            Text(
                text = "Sensio",
                fontSize = SensioTypography.HeadlineMobile,
                fontWeight = FontWeight.Bold,
                color = Color.White,
                textAlign = TextAlign.Center
            )
        }

        val state = uiState
        when (state) {
            is SplashUiState.Error -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(SensioSpacing.Md)
                    ) {
                        Text(
                            text = state.message,
                            color = Color.White,
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
            }
            else -> {}
        }
    }
}