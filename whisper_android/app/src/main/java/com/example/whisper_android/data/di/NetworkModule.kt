package com.example.whisper_android.data.di

import com.example.whisper_android.data.remote.api.TeraluxApi
import com.example.whisper_android.data.repository.TeraluxRepositoryImpl
import com.example.whisper_android.domain.repository.TeraluxRepository
import com.example.whisper_android.domain.usecase.AuthenticateUseCase
import com.example.whisper_android.domain.usecase.GetTeraluxByMacUseCase
import com.example.whisper_android.domain.usecase.RegisterTeraluxUseCase
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

object NetworkModule {
    // Read from BuildConfig (local.properties)
    val BASE_URL = com.example.whisper_android.BuildConfig.BASE_URL
    private val BASE_HOSTNAME = com.example.whisper_android.BuildConfig.BASE_HOSTNAME
    private val API_KEY = com.example.whisper_android.BuildConfig.TERALUX_API_KEY
    private val client by lazy {
        val logging =
            HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BODY
            }
        OkHttpClient
            .Builder()
            .addInterceptor(logging)
            .build()
    }

    private val retrofit by lazy {
        Retrofit
            .Builder()
            .baseUrl(BASE_URL)
            .client(client)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    lateinit var tokenManager: com.example.whisper_android.data.local.TokenManager

    fun init(context: android.content.Context) {
        tokenManager =
            com.example.whisper_android.data.local
                .TokenManager(context)
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
        com.example.whisper_android.data.repository
            .TuyaRepositoryImpl(tuyaApi, API_KEY, tokenManager)
    }

    private val ragApi: com.example.whisper_android.data.remote.api.RAGApi by lazy {
        retrofit.create(com.example.whisper_android.data.remote.api.RAGApi::class.java)
    }

    private val emailApi: com.example.whisper_android.data.remote.api.EmailApi by lazy {
        retrofit.create(com.example.whisper_android.data.remote.api.EmailApi::class.java)
    }

    val speechRepository: com.example.whisper_android.domain.repository.SpeechRepository by lazy {
        com.example.whisper_android.data.repository
            .SpeechRepositoryImpl(speechApi)
    }

    val ragRepository: com.example.whisper_android.domain.repository.RagRepository by lazy {
        com.example.whisper_android.data.repository
            .RagRepositoryImpl(ragApi)
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

    val transcribeAudioUseCase: com.example.whisper_android.domain.usecase.TranscribeAudioUseCase by lazy {
        com.example.whisper_android.domain.usecase
            .TranscribeAudioUseCase(speechRepository)
    }

    val translateTextUseCase: com.example.whisper_android.domain.usecase.TranslateTextUseCase by lazy {
        com.example.whisper_android.domain.usecase
            .TranslateTextUseCase(ragRepository)
    }

    val summarizeTextUseCase: com.example.whisper_android.domain.usecase.SummarizeTextUseCase by lazy {
        com.example.whisper_android.domain.usecase
            .SummarizeTextUseCase(ragRepository)
    }

    val processMeetingUseCase: com.example.whisper_android.domain.usecase.ProcessMeetingUseCase by lazy {
        com.example.whisper_android.domain.usecase.ProcessMeetingUseCase(
            transcribeAudioUseCase,
            translateTextUseCase,
            summarizeTextUseCase,
        )
    }

    val emailRepository: com.example.whisper_android.domain.repository.EmailRepository by lazy {
        com.example.whisper_android.data.repository
            .EmailRepositoryImpl(emailApi)
    }

    val sendEmailUseCase: com.example.whisper_android.domain.usecase.SendEmailUseCase by lazy {
        com.example.whisper_android.domain.usecase
            .SendEmailUseCase(emailRepository)
    }
}
