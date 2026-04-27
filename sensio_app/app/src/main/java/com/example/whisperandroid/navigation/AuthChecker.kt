package com.example.whisperandroid.navigation

import com.example.whisperandroid.data.auth.AuthStateManager
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.navigation.AppRoutes

object AuthChecker {
    fun getStartDestination(): String {
        if (!AuthStateManager.isInitialized()) {
            AuthStateManager.init(NetworkModule.commonApi, NetworkModule.appContext)
        }
        return AppRoutes.Authenticating.route
    }
}
