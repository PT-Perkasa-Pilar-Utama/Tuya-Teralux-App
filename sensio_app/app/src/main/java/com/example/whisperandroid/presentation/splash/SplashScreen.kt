package com.example.whisperandroid.presentation.splash

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.EaseInOut
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
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
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
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
    var animationStarted by remember { mutableStateOf(false) }
    var navigateOut by remember { mutableStateOf(false) }

    // Gradient flow animation - subtle vertical shift
    val infiniteTransition = rememberInfiniteTransition(label = "gradientFlow")
    val gradientOffset by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            tween(3000, easing = LinearEasing),
            RepeatMode.Reverse
        ),
        label = "gradientOffset"
    )

    LaunchedEffect(Unit) {
        animationStarted = true
    }

    // Handle zoom-out transition
    LaunchedEffect(uiState) {
        when (uiState) {
            is SplashUiState.Authenticated -> {
                navigateOut = true
                kotlinx.coroutines.delay(300) // Wait for zoom-out animation
                navController?.navigate(AppRoutes.Dashboard.route) {
                    popUpTo(AppRoutes.Splash.route) { inclusive = true }
                }
            }
            is SplashUiState.NotRegistered -> {
                navigateOut = true
                kotlinx.coroutines.delay(300)
                navController?.navigate(AppRoutes.Register.route) {
                    popUpTo(AppRoutes.Splash.route) { inclusive = true }
                }
            }
            is SplashUiState.Unauthorized -> {
                navigateOut = true
                kotlinx.coroutines.delay(300)
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
            .graphicsLayer {
                scaleX = if (navigateOut) 1.05f else 1f
                scaleY = if (navigateOut) 1.05f else 1f
                alpha = if (navigateOut) 0f else 1f
            }
    ) {
        // Animated gradient background with flow effect
        Canvas(modifier = Modifier.fillMaxSize()) {
            val brush = Brush.verticalGradient(
                colors = listOf(
                    Color(0xFF67E8F9),
                    Color(0xFF0EA5E9),
                    Color(0xFF075985)
                ),
                startY = gradientOffset * size.height,
                endY = size.height + gradientOffset * size.height
            )
            drawRect(brush = brush)
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = SensioSpacing.Lg),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            AnimatedVisibility(
                visible = animationStarted && !navigateOut,
                enter = scaleIn(
                    initialScale = 0.9f,
                    animationSpec = tween(
                        durationMillis = 800,
                        easing = EaseInOut
                    )
                ) + fadeIn(
                    animationSpec = tween(
                        durationMillis = 800,
                        easing = EaseInOut
                    )
                ),
                exit = scaleOut(
                    targetScale = 1.05f,
                    animationSpec = tween(
                        durationMillis = 300,
                        easing = EaseInOut
                    )
                ) + fadeOut(
                    animationSpec = tween(
                        durationMillis = 300,
                        easing = EaseInOut
                    )
                )
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
