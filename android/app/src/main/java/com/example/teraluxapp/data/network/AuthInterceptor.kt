package com.example.teraluxapp.data.network

import com.example.teraluxapp.utils.SessionManager
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.Response

class AuthInterceptor : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val response = chain.proceed(chain.request())

        if (response.code == 401) {
            runBlocking {
                SessionManager.triggerLogout()
            }
        }
        
        return response
    }
}
