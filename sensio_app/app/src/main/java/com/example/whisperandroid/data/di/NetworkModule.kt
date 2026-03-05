package com.example.whisperandroid.data.di

import com.example.whisperandroid.data.remote.api.TerminalApi
import com.example.whisperandroid.data.repository.TerminalRepositoryImpl
import com.example.whisperandroid.domain.repository.TerminalRepository
import com.example.whisperandroid.domain.usecase.AuthenticateUseCase
import com.example.whisperandroid.domain.usecase.GetTerminalByMacUseCase
import com.example.whisperandroid.domain.usecase.RegisterTerminalUseCase
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

object NetworkModule {
    // Read from BuildConfig (local.properties)
    val BASE_URL = com.example.whisperandroid.BuildConfig.BASE_URL
    private val BASE_HOSTNAME = com.example.whisperandroid.BuildConfig.BASE_HOSTNAME
    private val API_KEY = com.example.whisperandroid.BuildConfig.SENSIO_API_KEY
    private val client by lazy {
        val logging =
            HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BODY
            }
        OkHttpClient.Builder()
            .addInterceptor(logging)
            .connectTimeout(10, java.util.concurrent.TimeUnit.MINUTES)
            .readTimeout(10, java.util.concurrent.TimeUnit.MINUTES)
            .writeTimeout(10, java.util.concurrent.TimeUnit.MINUTES)
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

    lateinit var tokenManager: com.example.whisperandroid.data.local.TokenManager
    lateinit var mqttHelper: com.example.whisperandroid.util.MqttHelper
    lateinit var appContext: android.content.Context

    fun init(context: android.content.Context) {
        appContext = context.applicationContext
        tokenManager =
            com.example.whisperandroid.data.local
                .TokenManager(context)
        mqttHelper = com.example.whisperandroid.util.MqttHelper(context)
    }

    private val api: TerminalApi by lazy {
        retrofit.create(TerminalApi::class.java)
    }

    private val tuyaApi: com.example.whisperandroid.data.remote.api.TuyaApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.TuyaApi::class.java)
    }

    private val speechApi: com.example.whisperandroid.data.remote.api.SpeechApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.SpeechApi::class.java)
    }

    val repository: TerminalRepository by lazy {
        // Ensure init() is called before accessing this
        TerminalRepositoryImpl(api, API_KEY, tokenManager)
    }

    val tuyaRepository: com.example.whisperandroid.domain.repository.TuyaRepository by lazy {
        com.example.whisperandroid.data.repository
            .TuyaRepositoryImpl(tuyaApi, API_KEY, tokenManager)
    }

    private val ragApi: com.example.whisperandroid.data.remote.api.RAGApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.RAGApi::class.java)
    }

    private val emailApi: com.example.whisperandroid.data.remote.api.EmailApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.EmailApi::class.java)
    }

    val speechRepository: com.example.whisperandroid.domain.repository.SpeechRepository by lazy {
        com.example.whisperandroid.data.repository
            .SpeechRepositoryImpl(speechApi)
    }

    val ragRepository: com.example.whisperandroid.domain.repository.RagRepository by lazy {
        com.example.whisperandroid.data.repository
            .RagRepositoryImpl(ragApi)
    }

    private val pipelineApi: com.example.whisperandroid.data.remote.api.PipelineApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.PipelineApi::class.java)
    }

    val pipelineRepository: com.example.whisperandroid.domain.repository.PipelineRepository by lazy {
        com.example.whisperandroid.data.repository.PipelineRepositoryImpl(pipelineApi)
    }

    val uploadRepository: com.example.whisperandroid.domain.repository.UploadRepository by lazy {
        com.example.whisperandroid.data.repository.UploadRepositoryImpl(speechApi)
    }

    val registerUseCase: RegisterTerminalUseCase by lazy {
        RegisterTerminalUseCase(repository)
    }

    val getTerminalByMacUseCase: GetTerminalByMacUseCase by lazy {
        GetTerminalByMacUseCase(repository)
    }

    val authenticateUseCase: AuthenticateUseCase by lazy {
        AuthenticateUseCase(tuyaRepository)
    }

    val transcribeAudioUseCase: com.example.whisperandroid.domain.usecase.TranscribeAudioUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .TranscribeAudioUseCase(speechRepository)
    }

    val translateTextUseCase: com.example.whisperandroid.domain.usecase.TranslateTextUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .TranslateTextUseCase(ragRepository)
    }

    val summarizeTextUseCase: com.example.whisperandroid.domain.usecase.SummarizeTextUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .SummarizeTextUseCase(ragRepository)
    }

    val processMeetingUseCase: com.example.whisperandroid.domain.usecase.ProcessMeetingUseCase by lazy {
        val prefs = appContext.getSharedPreferences("upload_sessions", android.content.Context.MODE_PRIVATE)
        com.example.whisperandroid.domain.usecase.ProcessMeetingUseCase(
            pipelineRepository,
            uploadRepository,
            mqttHelper,
            prefs
        )
    }

    val emailRepository: com.example.whisperandroid.domain.repository.EmailRepository by lazy {
        com.example.whisperandroid.data.repository
            .EmailRepositoryImpl(emailApi)
    }

    val sendEmailUseCase: com.example.whisperandroid.domain.usecase.SendEmailUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .SendEmailUseCase(emailRepository)
    }

    val sendEmailByMacUseCase: com.example.whisperandroid.domain.usecase.SendEmailByMacUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .SendEmailByMacUseCase(emailRepository)
    }

    val getTuyaDevicesUseCase: com.example.whisperandroid.domain.usecase.GetTuyaDevicesUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .GetTuyaDevicesUseCase(tuyaRepository)
    }
}
