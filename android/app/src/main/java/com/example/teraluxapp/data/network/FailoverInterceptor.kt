package com.example.teraluxapp.data.network

import com.example.teraluxapp.config.ApiConfig
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.Interceptor
import okhttp3.Response
import java.io.IOException

class FailoverInterceptor : Interceptor {

    // Store the index of the currently successful URL.
    // We try to stick to one URL as long as it works.
    @Volatile private var currentUrlIndex = 0

    override fun intercept(chain: Interceptor.Chain): Response {
        var request = chain.request()
        var exception: IOException? = null

        // Try all available URLs starting from the current successful one
        for (i in ApiConfig.BASE_URLS.indices) {
            val urlIndexToTry = (currentUrlIndex + i) % ApiConfig.BASE_URLS.size
            val baseUrlStr = ApiConfig.BASE_URLS[urlIndexToTry]
            
            // Safe parsing of the base URL
            val baseUrl = baseUrlStr.toHttpUrlOrNull()
            
            if (baseUrl == null) {
                // Should not happen if config is correct, but safe to skip
                continue
            }

            // Reconstruct the request URL with the new scheme/host/port
            val newUrl = request.url.newBuilder()
                .scheme(baseUrl.scheme)
                .host(baseUrl.host)
                .port(baseUrl.port)
                .build()

            val newRequest = request.newBuilder()
                .url(newUrl)
                .build()

            try {
                val response = chain.proceed(newRequest)
                
                // If we get here, the network call completed (even if 4xx/5xx).
                // "Down" usually means unreachable, so we consider this a success for the host connection.
                // We update the sticky index to this working host.
                currentUrlIndex = urlIndexToTry
                return response
            } catch (e: IOException) {
                // Network failure (timeout, unreachable, etc.)
                // Save exception to throw if valid all attempts fail
                exception = e
                // Continue loop to try next URL
            }
        }

        // If all attempts failed, throw the last exception
        throw exception ?: IOException("All Base URLs failed to connect")
    }
}
