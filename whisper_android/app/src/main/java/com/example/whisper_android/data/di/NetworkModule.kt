package com.example.whisper_android.data.di

import com.example.whisper_android.data.remote.api.TeraluxApi
import com.example.whisper_android.data.repository.TeraluxRepositoryImpl
import com.example.whisper_android.domain.repository.TeraluxRepository
import com.example.whisper_android.domain.usecase.RegisterTeraluxUseCase
import com.example.whisper_android.domain.usecase.GetTeraluxByMacUseCase
import com.example.whisper_android.domain.usecase.AuthenticateUseCase
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import okhttp3.Dns
import java.net.InetAddress

object NetworkModule {
    // Updated to HTTPS domain (Standard Port 443) as verified by user curl
    private const val BASE_URL = "https://teralux.farismunir.my.id/" 
    private const val API_KEY = "REDACTED_SECRET" // From backend .env

    // Custom DNS to bypass device DNS scaling issues
    private class CustomDns : Dns {
        override fun lookup(hostname: String): List<InetAddress> {
            if (hostname == "teralux.farismunir.my.id") {
                return try {
                    // Hardcoded IP from `dig` command on host machine
                    // Cloudflare IPs: 172.67.136.115, 104.21.46.81
                    listOf(InetAddress.getByName("104.21.46.81")) 
                } catch (e: Exception) {
                    Dns.SYSTEM.lookup(hostname)
                }
            }
            return Dns.SYSTEM.lookup(hostname)
        }
    }

    private val client by lazy {
        val logging = HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        }
        OkHttpClient.Builder()
            .dns(CustomDns()) // Apply Custom DNS
            .addInterceptor(logging)
            .build()
    }

    private val retrofit by lazy {
        Retrofit.Builder()
            .baseUrl(BASE_URL)
            .client(client)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    lateinit var tokenManager: com.example.whisper_android.data.local.TokenManager

    fun init(context: android.content.Context) {
        tokenManager = com.example.whisper_android.data.local.TokenManager(context)
    }

    private val api: TeraluxApi by lazy {
        retrofit.create(TeraluxApi::class.java)
    }

    private val tuyaApi: com.example.whisper_android.data.remote.api.TuyaApi by lazy {
        retrofit.create(com.example.whisper_android.data.remote.api.TuyaApi::class.java)
    }

    private val speechApi: com.example.whisper_android.data.remote.api.SpeechApi by lazy {
        retrofit.create(com.example.whisper_android.data.remote.api.SpeechApi::class.java)
    }

    val repository: TeraluxRepository by lazy {
        // Ensure init() is called before accessing this
        TeraluxRepositoryImpl(api, API_KEY)
    }

    val tuyaRepository: com.example.whisper_android.domain.repository.TuyaRepository by lazy {
        com.example.whisper_android.data.repository.TuyaRepositoryImpl(tuyaApi, API_KEY, tokenManager)
    }

    val speechRepository: com.example.whisper_android.data.repository.SpeechRepository by lazy {
        com.example.whisper_android.data.repository.SpeechRepository(speechApi)
    }
    
    val registerUseCase: RegisterTeraluxUseCase by lazy {
        RegisterTeraluxUseCase(repository)
    }

    val getTeraluxByMacUseCase: GetTeraluxByMacUseCase by lazy {
        GetTeraluxByMacUseCase(repository)
    }

    val authenticateUseCase: AuthenticateUseCase by lazy {
        AuthenticateUseCase(tuyaRepository)
    }
}
