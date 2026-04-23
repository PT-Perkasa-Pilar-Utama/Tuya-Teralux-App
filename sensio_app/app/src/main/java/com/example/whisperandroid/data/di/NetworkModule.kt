package com.example.whisperandroid.data.di

import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.remote.api.PipelineApi
import com.example.whisperandroid.data.remote.api.TerminalApi
import com.example.whisperandroid.data.repository.PipelineRepositoryImpl
import com.example.whisperandroid.data.repository.TerminalRepositoryImpl
import com.example.whisperandroid.data.repository.UploadRepositoryImpl
import com.example.whisperandroid.domain.repository.PipelineRepository
import com.example.whisperandroid.domain.repository.TerminalRepository
import com.example.whisperandroid.domain.repository.UploadRepository
import com.example.whisperandroid.domain.usecase.AuthenticateUseCase
import com.example.whisperandroid.domain.usecase.GetTerminalByMacUseCase
import com.example.whisperandroid.domain.usecase.GetTuyaDevicesUseCase
import com.example.whisperandroid.domain.usecase.ProcessMeetingUseCase
import com.example.whisperandroid.domain.usecase.RegisterTerminalUseCase
import com.example.whisperandroid.domain.usecase.SendEmailByMacUseCase
import com.example.whisperandroid.domain.usecase.SummarizeTextUseCase
import com.example.whisperandroid.domain.usecase.TranscribeAudioUseCase
import com.example.whisperandroid.domain.usecase.TranslateTextUseCase
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import okhttp3.Cookie
import okhttp3.CookieJar
import okhttp3.HttpUrl
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

object NetworkModule {
    // Read from BuildConfig (local.properties)
    val BASE_URL = com.example.whisperandroid.BuildConfig.BASE_URL
    private val BASE_HOSTNAME = com.example.whisperandroid.BuildConfig.BASE_HOSTNAME
    private val API_KEY = com.example.whisperandroid.BuildConfig.SENSIO_API_KEY

    private val cookieJar by lazy {
        object : CookieJar {
            private val cookieStore = mutableMapOf<String, MutableList<Cookie>>()

            override fun saveFromResponse(url: HttpUrl, cookies: List<Cookie>) {
                val host = url.host
                val existing = cookieStore[host] ?: mutableListOf()
                cookies.forEach { cookie ->
                    existing.removeAll { it.name == cookie.name && it.domain == cookie.domain }
                    existing.add(cookie)
                }
                cookieStore[host] = existing
            }

            override fun loadForRequest(url: HttpUrl): List<Cookie> {
                return cookieStore[url.host]?.filter { cookie ->
                    val matcher = okhttp3.Cookie.Builder()
                        .name(cookie.name)
                        .domain(cookie.domain)
                        .build()
                    matcher.matches(url)
                } ?: emptyList()
            }
        }
    }

    // Shared client for general API calls
    private val client by lazy {
        val logging =
            HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BASIC
            }
        OkHttpClient.Builder()
            .cookieJar(cookieJar)
            .addInterceptor(logging)
            .connectTimeout(45, java.util.concurrent.TimeUnit.SECONDS)
            .readTimeout(45, java.util.concurrent.TimeUnit.SECONDS)
            .writeTimeout(45, java.util.concurrent.TimeUnit.SECONDS)
            .build()
    }

    // Dedicated upload client with bounded wall-clock timeout
    // Timeouts are configured to handle slow uplink conditions:
    // - 8MB chunk at 100 KB/s = ~80 seconds, so we use 5 minutes as safe buffer
    // - callTimeout is set to 6 minutes to allow for retry overhead
    private val uploadClient by lazy {
        val logging =
            HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BASIC
            }
        OkHttpClient.Builder()
            .addInterceptor(logging)
            .connectTimeout(30, java.util.concurrent.TimeUnit.SECONDS)
            .readTimeout(5, java.util.concurrent.TimeUnit.MINUTES)
            .writeTimeout(5, java.util.concurrent.TimeUnit.MINUTES)
            .callTimeout(6, java.util.concurrent.TimeUnit.MINUTES)
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

    // Dedicated Retrofit for upload APIs with bounded timeout client
    private val uploadRetrofit by lazy {
        Retrofit
            .Builder()
            .baseUrl(BASE_URL)
            .client(uploadClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    lateinit var tokenManager: com.example.whisperandroid.data.local.TokenManager
    lateinit var mqttHelper: com.example.whisperandroid.utils.MqttHelper
    lateinit var backgroundAssistantModeStore: com.example.whisperandroid.data.local.BackgroundAssistantModeStore

    // Meeting reminder components
    lateinit var meetingReminderStore: com.example.whisperandroid.data.local.reminder.MeetingReminderStore
    lateinit var meetingReminderScheduler: com.example.whisperandroid.service.reminder.MeetingReminderScheduler
    lateinit var meetingReminderNotifier: com.example.whisperandroid.service.reminder.MeetingReminderNotifier
    lateinit var overlayArbiter: com.example.whisperandroid.service.reminder.OverlayArbiter
    lateinit var meetingReminderOverlayController: com.example.whisperandroid.service.reminder.MeetingReminderOverlayController
    lateinit var meetingReminderRuntimeCoordinator: com.example.whisperandroid.service.reminder.MeetingReminderRuntimeCoordinator
    lateinit var backgroundAssistantCoordinator: com.example.whisperandroid.presentation.assistant.BackgroundAssistantCoordinator

    lateinit var appContext: android.content.Context

    private val _isTuyaSyncReady = MutableStateFlow(false)
    val isTuyaSyncReady = _isTuyaSyncReady.asStateFlow()

    fun setTuyaSyncReady(ready: Boolean) {
        _isTuyaSyncReady.value = ready
    }

    fun init(context: android.content.Context) {
        if (::appContext.isInitialized) return
        appContext = context.applicationContext
        tokenManager =
            com.example.whisperandroid.data.local
                .TokenManager(appContext)
        mqttHelper = com.example.whisperandroid.utils.MqttHelper(appContext)
        backgroundAssistantModeStore = com.example.whisperandroid.data.local.BackgroundAssistantModeStore(appContext)

        // Initialize meeting reminder components
        meetingReminderStore = com.example.whisperandroid.data.local.reminder.MeetingReminderStore(appContext)
        meetingReminderScheduler = com.example.whisperandroid.service.reminder.MeetingReminderScheduler(appContext)
        meetingReminderNotifier = com.example.whisperandroid.service.reminder.MeetingReminderNotifier(appContext)
        overlayArbiter = com.example.whisperandroid.service.reminder.OverlayArbiter(appContext)
        meetingReminderOverlayController = com.example.whisperandroid.service.reminder.MeetingReminderOverlayController(
            context = appContext,
            arbiter = overlayArbiter
        )
        meetingReminderRuntimeCoordinator = com.example.whisperandroid.service.reminder.MeetingReminderRuntimeCoordinator(
            context = appContext,
            store = meetingReminderStore,
            scheduler = meetingReminderScheduler,
            notifier = meetingReminderNotifier,
            overlayController = meetingReminderOverlayController,
            arbiter = overlayArbiter
        )
        backgroundAssistantCoordinator = com.example.whisperandroid.presentation.assistant.BackgroundAssistantCoordinator(
            appContext as android.app.Application
        )
    }

    fun ensureInitialized(context: android.content.Context) {
        init(context)
    }

    private val api: TerminalApi by lazy {
        retrofit.create(TerminalApi::class.java)
    }

    val commonApi: CommonApi by lazy {
        retrofit.create(CommonApi::class.java)
    }

    private val tuyaApi: com.example.whisperandroid.data.remote.api.TuyaApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.TuyaApi::class.java)
    }

    private val whisperApi: com.example.whisperandroid.data.remote.api.WhisperApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.WhisperApi::class.java)
    }

    // Dedicated WhisperApi for upload operations with bounded timeout
    private val uploadWhisperApi: com.example.whisperandroid.data.remote.api.WhisperApi by lazy {
        uploadRetrofit.create(com.example.whisperandroid.data.remote.api.WhisperApi::class.java)
    }

    val repository: TerminalRepository by lazy {
        // Ensure init() is called before accessing this
        TerminalRepositoryImpl(api, API_KEY, tokenManager)
    }

    val terminalRepository: TerminalRepository by lazy {
        repository
    }

    val tuyaRepository: com.example.whisperandroid.domain.repository.TuyaRepository by lazy {
        com.example.whisperandroid.data.repository
            .TuyaRepositoryImpl(tuyaApi, tokenManager, API_KEY)
    }

    private val ragApi: com.example.whisperandroid.data.remote.api.RAGApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.RAGApi::class.java)
    }

    private val emailApi: com.example.whisperandroid.data.remote.api.EmailApi by lazy {
        retrofit.create(com.example.whisperandroid.data.remote.api.EmailApi::class.java)
    }

    val whisperRepository: com.example.whisperandroid.domain.repository.WhisperRepository by lazy {
        com.example.whisperandroid.data.repository
            .WhisperRepositoryImpl(whisperApi)
    }

    val ragRepository: com.example.whisperandroid.domain.repository.RagRepository by lazy {
        com.example.whisperandroid.data.repository
            .RagRepositoryImpl(ragApi)
    }

    private val pipelineApi: PipelineApi by lazy {
        retrofit.create(PipelineApi::class.java)
    }

    val pipelineRepository: PipelineRepository by lazy {
        PipelineRepositoryImpl(pipelineApi)
    }

    val uploadRepository: UploadRepository by lazy {
        UploadRepositoryImpl(
            uploadWhisperApi,
            signedUploadModeStore.isEnabled.value
        )
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

    val transcribeAudioUseCase: TranscribeAudioUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .TranscribeAudioUseCase(whisperRepository)
    }

    val translateTextUseCase: TranslateTextUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .TranslateTextUseCase(ragRepository)
    }

    val summarizeTextUseCase: SummarizeTextUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .SummarizeTextUseCase(ragRepository)
    }

    val processMeetingUseCase: ProcessMeetingUseCase by lazy {
        val prefs = appContext.getSharedPreferences(
            "upload_sessions",
            android.content.Context.MODE_PRIVATE
        )
        com.example.whisperandroid.domain.usecase.ProcessMeetingUseCase(
            pipelineRepository,
            uploadRepository,
            mqttHelper,
            prefs,
            failedUploadStore,
            signedUploadModeStore
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

    val sendEmailByMacUseCase: SendEmailByMacUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .SendEmailByMacUseCase(emailRepository)
    }

    val getTuyaDevicesUseCase: GetTuyaDevicesUseCase by lazy {
        com.example.whisperandroid.domain.usecase
            .GetTuyaDevicesUseCase(tuyaRepository)
    }

    val signedUploadModeStore: com.example.whisperandroid.data.local.SignedUploadModeStore by lazy {
        com.example.whisperandroid.data.local.SignedUploadModeStore(appContext)
    }

    val failedUploadStore: com.example.whisperandroid.data.local.FailedUploadStore by lazy {
        com.example.whisperandroid.data.local.FailedUploadStore(appContext)
    }
}
