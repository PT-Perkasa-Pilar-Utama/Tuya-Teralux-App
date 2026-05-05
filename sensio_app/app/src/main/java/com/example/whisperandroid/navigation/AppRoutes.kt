package com.example.whisperandroid.navigation

sealed class AppRoutes(val route: String) {
    object Register : AppRoutes("register")
    object Dashboard : AppRoutes("dashboard")
    object Meeting : AppRoutes("meeting")
    object Assistant : AppRoutes("assistant")
    object Auth : AppRoutes("auth")
}
