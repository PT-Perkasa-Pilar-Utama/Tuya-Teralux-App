package com.example.whisperandroid.util

import android.content.Context
import android.util.Log
import com.example.whisperandroid.BuildConfig
import info.mqtt.android.service.MqttAndroidClient
import java.util.UUID
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import org.eclipse.paho.client.mqttv3.DisconnectedBufferOptions
import org.eclipse.paho.client.mqttv3.IMqttActionListener
import org.eclipse.paho.client.mqttv3.IMqttDeliveryToken
import org.eclipse.paho.client.mqttv3.IMqttToken
import org.eclipse.paho.client.mqttv3.MqttConnectOptions
import org.eclipse.paho.client.mqttv3.MqttException
import org.eclipse.paho.client.mqttv3.MqttMessage

class MqttHelper(
    private val context: Context
) {
    private var mqttAndroidClient: MqttAndroidClient
    private val tokenManager = com.example.whisperandroid.data.local.TokenManager(context)

    private val _connectionStatus = kotlinx.coroutines.flow.MutableStateFlow(
        MqttConnectionStatus.DISCONNECTED
    )
    val connectionStatus = _connectionStatus.asStateFlow()

    private val serverUri = BuildConfig.MQTT_BROKER_URL
    private val clientID = "WhisperAndroid_" + UUID.randomUUID().toString()

    private val tag = "MqttHelper"

    // Track external topics for re-subscription on reconnection
    private val externalTopics = mutableSetOf<String>()

    private val _messages = kotlinx.coroutines.flow.MutableSharedFlow<Pair<String, String>>(
        extraBufferCapacity = 64,
        onBufferOverflow = kotlinx.coroutines.channels.BufferOverflow.DROP_OLDEST
    )
    val messages = _messages.asSharedFlow()

    private fun getUsername(): String {
        return com.example.whisperandroid.util.DeviceUtils.getDeviceId(context)
    }

    private fun resolveTuyaUid(): String? {
        return tokenManager.getTuyaUid()?.trim()?.takeIf { it.isNotEmpty() }
    }

    fun getTaskTopic(): String? {
        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        return "users/$username/$env/task"
    }

    enum class MqttConnectionStatus {
        DISCONNECTED,
        CONNECTING,
        CONNECTED,
        FAILED,
        NO_INTERNET
    }

    init {
        val appContext = context.applicationContext
        mqttAndroidClient = MqttAndroidClient(appContext, serverUri, clientID)
        mqttAndroidClient.setCallback(
            object : org.eclipse.paho.client.mqttv3.MqttCallbackExtended {
                override fun connectComplete(reconnect: Boolean, serverURI: String?) {
                    Log.d(tag, "Connect complete: reconnect=$reconnect, uri=$serverURI")
                    _connectionStatus.value = MqttConnectionStatus.CONNECTED

                    // Re-subscribe if this was an automatic reconnection
                    if (reconnect) {
                        subscribeToAllTopics()
                    }
                }

                override fun connectionLost(cause: Throwable?) {
                    Log.d(tag, "Connection lost: ${cause?.message}")
                    _connectionStatus.value = MqttConnectionStatus.DISCONNECTED
                }

                override fun messageArrived(
                    topic: String,
                    message: MqttMessage
                ) {
                    Log.d(tag, "Message arrived: $topic -> $message")
                    _messages.tryEmit(topic to message.toString())
                }

                override fun deliveryComplete(token: IMqttDeliveryToken?) {}
            }
        )
    }

    private fun isNetworkAvailable(): Boolean {
        val connectivityManager = context
            .getSystemService(Context.CONNECTIVITY_SERVICE) as android.net.ConnectivityManager
        val network = connectivityManager.activeNetwork ?: return false
        val capabilities = connectivityManager.getNetworkCapabilities(network) ?: return false
        return capabilities.hasCapability(android.net.NetworkCapabilities.NET_CAPABILITY_INTERNET)
    }

    private fun subscribeToAllTopics() {
        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        subscribeInternal("users/$username/$env/chat/answer")
        subscribeInternal("users/$username/$env/whisper/answer")
        subscribeInternal("users/$username/$env/task")
        subscribeInternal("users/$username/$env/chat")

        // Re-subscribe to external topics (e.g., notification topic)
        externalTopics.forEach { topic ->
            subscribeInternal(topic)
        }
    }

    /**
     * Public subscribe method for external components (e.g., reminder coordinator).
     */
    fun subscribe(topic: String) {
        externalTopics.add(topic)
        subscribeInternal(topic)
    }

    private fun subscribeInternal(topic: String) {
        try {
            mqttAndroidClient.subscribe(
                topic,
                0,
                null,
                object : IMqttActionListener {
                    override fun onSuccess(asyncActionToken: IMqttToken) {
                        Log.d(tag, "Subscribed to $topic")
                    }

                    override fun onFailure(
                        asyncActionToken: IMqttToken,
                        exception: Throwable
                    ) {
                        Log.e(tag, "Failed to subscribe to $topic")
                    }
                }
            )
        } catch (e: MqttException) {
            e.printStackTrace()
        }
    }

    /**
     * Connect to MQTT broker by fetching credentials from backend first.
     * Password is never stored - only used for this connection attempt.
     */
    suspend fun connect() {
        val isClientConnected = try { mqttAndroidClient.isConnected } catch (e: Exception) { false }

        if (isClientConnected && _connectionStatus.value == MqttConnectionStatus.CONNECTED) {
            Log.d(tag, "MQTT already connected. Skipping connect.")
            return
        }

        if (isClientConnected && _connectionStatus.value != MqttConnectionStatus.CONNECTED) {
            Log.d(
                tag,
                "MQTT client connected but state is ${_connectionStatus.value}. " +
                    "Syncing state and re-subscribing."
            )
            _connectionStatus.value = MqttConnectionStatus.CONNECTED
            subscribeToAllTopics()
            return
        }

        if (_connectionStatus.value == MqttConnectionStatus.CONNECTING) {
            Log.d(tag, "MQTT connection already in progress. Skipping connect.")
            return
        }

        if (!isNetworkAvailable()) {
            Log.w(tag, "No internet connection available. Skipping MQTT connect.")
            _connectionStatus.value = MqttConnectionStatus.NO_INTERNET
            return
        }

        val username = getUsername()

        // Fetch credentials from backend (password is never stored)
        val password = fetchMqttCredentials(username)
            ?: run {
                Log.e(tag, "Failed to fetch MQTT credentials for $username")
                _connectionStatus.value = MqttConnectionStatus.FAILED
                return
            }

        val mqttConnectOptions = MqttConnectOptions()
        mqttConnectOptions.isAutomaticReconnect = true
        mqttConnectOptions.isCleanSession = false
        mqttConnectOptions.keepAliveInterval = 30
        mqttConnectOptions.connectionTimeout = 5 // 5 seconds connection timeout
        mqttConnectOptions.userName = username
        mqttConnectOptions.password = password.toCharArray()

        // Enable TLS for MQTTS connections (ssl:// or tcps://)
        // Note: Paho Android only supports ssl:// and tcps:// schemes
        if (serverUri.startsWith("ssl://") || serverUri.startsWith("tcps://")) {
            try {
                val sslContext = javax.net.ssl.SSLContext.getInstance("TLSv1.2")
                sslContext.init(null, null, java.security.SecureRandom())
                mqttConnectOptions.socketFactory = sslContext.socketFactory
                Log.d(tag, "TLS 1.2 enabled for MQTTS connection: $serverUri")
            } catch (e: Exception) {
                Log.e(tag, "Failed to initialize TLS: ${e.message}")
            }
        }

        _connectionStatus.value = MqttConnectionStatus.CONNECTING
        try {
            mqttAndroidClient.connect(
                mqttConnectOptions,
                null,
                object : IMqttActionListener {
                    override fun onSuccess(asyncActionToken: IMqttToken) {
                        val disconnectedBufferOptions = DisconnectedBufferOptions()
                        disconnectedBufferOptions.isBufferEnabled = true
                        disconnectedBufferOptions.bufferSize = 100
                        disconnectedBufferOptions.isPersistBuffer = false
                        disconnectedBufferOptions.isDeleteOldestMessages = false
                        mqttAndroidClient.setBufferOpts(disconnectedBufferOptions)
                        Log.d(tag, "Success Connected to $serverUri")

                        subscribeToAllTopics()
                        _connectionStatus.value = MqttConnectionStatus.CONNECTED
                    }

                    override fun onFailure(
                        asyncActionToken: IMqttToken,
                        exception: Throwable
                    ) {
                        Log.w(tag, "Failed to connect to: $serverUri")
                        exception.printStackTrace()

                        val isUnknownHost = exception is java.net.UnknownHostException ||
                            exception.cause is java.net.UnknownHostException
                        val status = if (isUnknownHost) {
                            MqttConnectionStatus.NO_INTERNET
                        } else {
                            MqttConnectionStatus.FAILED
                        }
                        _connectionStatus.value = status
                    }
                }
            )
        } catch (ex: MqttException) {
            ex.printStackTrace()
        }
    }

    /**
     * Fetch MQTT credentials from backend.
     * Password is returned only for this connection attempt and is never stored.
     */
    private suspend fun fetchMqttCredentials(username: String): String? {
        return try {
            val token = tokenManager.getAccessToken()
                ?: run {
                    Log.e(tag, "No access token available for fetching MQTT credentials")
                    return null
                }

            kotlinx.coroutines.withTimeout(10000L) {
                val repository = com.example.whisperandroid.data.di.NetworkModule.repository
                val result = repository.fetchMqttPassword(username)
                result.getOrNull()
            }
        } catch (e: Exception) {
            Log.e(tag, "Error fetching MQTT credentials: ${e.message}")
            null
        }
    }

    fun disconnect() {
        if (!mqttAndroidClient.isConnected) return
        try {
            mqttAndroidClient.disconnect()
            Log.d(tag, "Disconnected from MQTT")
            _connectionStatus.value = MqttConnectionStatus.DISCONNECTED
        } catch (e: MqttException) {
            e.printStackTrace()
        }
    }

    suspend fun publishAudio(
        payload: ByteArray,
        language: String = "id",
        requestId: String? = null
    ): Result<Unit> {
        val base64Audio = android.util.Base64.encodeToString(payload, android.util.Base64.NO_WRAP)
        val terminalId = tokenManager.getTerminalId() ?: "unknown-terminal"
        val tuyaUid = resolveTuyaUid()
        val jsonPayload = MqttPayloadFactory.buildAudioPayload(
            base64Audio = base64Audio,
            terminalId = terminalId,
            language = language,
            requestId = requestId,
            tuyaUid = tuyaUid
        )

        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        return publishWithTimeout("users/$username/$env/whisper", jsonPayload.toString().toByteArray())
    }

    suspend fun publishChat(
        text: String,
        language: String = "id",
        requestId: String? = null
    ): Result<Unit> {
        val terminalId = tokenManager.getTerminalId() ?: "unknown-terminal"
        val tuyaUid = resolveTuyaUid()
        val jsonPayload = MqttPayloadFactory.buildChatPayload(
            text = text,
            terminalId = terminalId,
            language = language,
            requestId = requestId,
            tuyaUid = tuyaUid
        )

        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        return publishWithTimeout("users/$username/$env/chat", jsonPayload.toString().toByteArray())
    }

    fun publishTaskMessage(event: String, task: String) {
        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        val json = """{"event": "$event", "task": "$task"}"""
        val payload = json.toByteArray()
        val isConnected = try { mqttAndroidClient.isConnected } catch (e: Exception) { false }
        if (!isConnected) return
        val message = MqttMessage(payload)
        message.qos = 0
        message.isRetained = false
        try {
            mqttAndroidClient.publish("users/$username/$env/task", message)
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    private suspend fun publishWithTimeout(
        topic: String,
        payload: ByteArray
    ): Result<Unit> {
        return try {
            kotlinx.coroutines.withTimeout(4000L) {
                kotlinx.coroutines.suspendCancellableCoroutine<Unit> { cont ->
                    val isConnected = try {
                        mqttAndroidClient.isConnected
                    } catch (e: Exception) {
                        false
                    }
                    Log.d(tag, "Publish to $topic. Connected: $isConnected")

                    if (!isConnected) {
                        Log.e(tag, "Cannot publish to $topic: MQTT client is not connected")
                        if (cont.isActive) {
                            cont.resumeWith(
                                Result.failure(
                                    IllegalStateException("MQTT client is not connected")
                                )
                            )
                        }
                        return@suspendCancellableCoroutine
                    }

                    try {
                        mqttAndroidClient.publish(
                            topic,
                            payload,
                            0,
                            false,
                            null,
                            object : IMqttActionListener {
                                override fun onSuccess(asyncActionToken: IMqttToken?) {
                                    Log.d(
                                        tag,
                                        "Successfully published to $topic: ${payload.size} bytes"
                                    )
                                    if (cont.isActive) cont.resumeWith(Result.success(Unit))
                                }

                                override fun onFailure(
                                    asyncActionToken: IMqttToken?,
                                    exception: Throwable?
                                ) {
                                    Log.e(tag, "Failed to publish to $topic: ${exception?.message}")
                                    if (cont.isActive) {
                                        cont.resumeWith(
                                            Result.failure(
                                                exception ?: RuntimeException("Unknown MQTT error")
                                            )
                                        )
                                    }
                                }
                            }
                        )
                    } catch (e: Exception) {
                        Log.e(tag, "Error publishing to $topic: ${e.message}")
                        if (cont.isActive) cont.resumeWith(Result.failure(e))
                    }
                }
            }
            Result.success(Unit)
        } catch (e: kotlinx.coroutines.TimeoutCancellationException) {
            Log.e(tag, "Publish to $topic timed out after 4s")
            Result.failure(e)
        } catch (e: Exception) {
            Log.e(tag, "Unexpected error in publishWithTimeout to $topic: ${e.message}")
            Result.failure(e)
        }
    }
}
