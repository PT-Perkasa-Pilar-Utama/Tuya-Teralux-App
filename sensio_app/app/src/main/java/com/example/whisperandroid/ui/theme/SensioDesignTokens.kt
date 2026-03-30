package com.example.whisperandroid.ui.theme

import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * Sensio Design Tokens
 * * Centralized design system for consistent spacing, typography, and radii across the app.
 * Based on 4dp grid system for vertical rhythm and visual harmony.
 */
object SensioSpacing {
    val Xs = 4.dp
    val Sm = 8.dp
    val Md = 16.dp
    val Lg = 24.dp
    val Xl = 32.dp
    val Xxl = 40.dp
    val Xxxl = 48.dp
}

object SensioRadius {
    val Sm = 8.dp
    val Md = 12.dp
    val Lg = 16.dp
    val Xl = 20.dp
    val Xxl = 24.dp
    val Xxxl = 28.dp
    val Full = 9999.dp
}

object SensioTypography {
    val HeadlineMobile = 24.sp
    val HeadlineTablet = 36.sp
    val HeadlineDesktop = 48.sp
    val Subtitle = 16.sp
    val Body = 14.sp
    val Caption = 12.sp
    val ButtonText = 16.sp
    val InputLabel = 14.sp
}

object SensioElevation {
    val None = 0.dp
    val Sm = 2.dp
    val Md = 4.dp
    val Lg = 8.dp
    val Xl = 12.dp
}

object SensioBorder {
    val None = 0.dp
    val Sm = 0.5.dp
    val Md = 1.dp
    val Lg = 1.5.dp
    val Xl = 2.dp
}
