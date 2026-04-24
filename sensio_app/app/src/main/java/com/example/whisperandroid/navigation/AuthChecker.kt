package com.example.whisperandroid.navigation

import com.example.whisperandroid.data.auth.AuthStateManager
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.navigation.AppRoutes

object AuthChecker {
    fun getStartDestination(): String {
        // Initialize if needed
        if (!AuthStateManager.isInitialized()) {
            AuthStateManager.init(NetworkModule.commonApi, NetworkModule.appContext)
        }
        
        val authState = AuthStateManager.checkAuthOnStart()
        
        return when (authState) {
            AuthStateManager.AuthState.Authenticated -> AppRoutes.Dashboard.route
            AuthStateManager.AuthState.NotRegistered,
            AuthStateManager.AuthState.Unauthorized -> AppRoutes.Register.route
            AuthStateManager.AuthState.Checking -> AppRoutes.Dashboard.route
        }
    }
}
