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
import org.eclipse.paho.client.mqttv3.MqttCallback
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

    private val _messages = kotlinx.coroutines.flow.MutableSharedFlow<Pair<String, String>>(
        extraBufferCapacity = 64,
        onBufferOverflow = kotlinx.coroutines.channels.BufferOverflow.DROP_OLDEST
    )
    val messages = _messages.asSharedFlow()

    private fun getUsername(): String {
        return com.example.whisperandroid.util.DeviceUtils.getDeviceId(context)
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
            object : MqttCallback {
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
        val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as android.net.ConnectivityManager
        val network = connectivityManager.activeNetwork ?: return false
        val capabilities = connectivityManager.getNetworkCapabilities(network) ?: return false
        return capabilities.hasCapability(android.net.NetworkCapabilities.NET_CAPABILITY_INTERNET)
    }

    fun connect(password: String) {
        if (mqttAndroidClient.isConnected || _connectionStatus.value == MqttConnectionStatus.CONNECTED) {
            Log.d(tag, "MQTT already connected. Skipping connect.")
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

        val mqttConnectOptions = MqttConnectOptions()
        mqttConnectOptions.isAutomaticReconnect = true
        mqttConnectOptions.isCleanSession = false
        mqttConnectOptions.keepAliveInterval = 30
        mqttConnectOptions.userName = username
        mqttConnectOptions.password = password.toCharArray()

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

                        val username = getUsername()
                        val env = BuildConfig.APPLICATION_ENVIRONMENT
                        subscribe("users/$username/$env/chat/answer")
                        subscribe("users/$username/$env/whisper/answer")
                        subscribe("users/$username/$env/task")
                        subscribe("users/$username/$env/chat")
                        _connectionStatus.value = MqttConnectionStatus.CONNECTED
                    }

                    override fun onFailure(
                        asyncActionToken: IMqttToken,
                        exception: Throwable
                    ) {
                        Log.w(tag, "Failed to connect to: $serverUri")
                        exception.printStackTrace()

                        val status = if (exception is java.net.UnknownHostException || exception.cause is java.net.UnknownHostException) {
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

    private fun subscribe(topic: String) {
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

    fun publishAudio(
        payload: ByteArray,
        language: String = "id"
    ) {
        val base64Audio = android.util.Base64.encodeToString(payload, android.util.Base64.NO_WRAP)
        val terminalId = tokenManager.getTerminalId() ?: "unknown-terminal"
        val json =
            """
            {
                "audio": "$base64Audio",
                "terminal_id": "$terminalId",
                "language": "$language"
            }
            """.trimIndent()
        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        publish("users/$username/$env/whisper", json.toByteArray())
    }

    fun publishChat(
        text: String,
        language: String = "id"
    ) {
        val terminalId = tokenManager.getTerminalId() ?: "unknown-terminal"
        val json =
            """
            {
                "prompt": "$text",
                "terminal_id": "$terminalId",
                "language": "$language"
            }
            """.trimIndent()
        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        publish("users/$username/$env/chat", json.toByteArray())
    }

    fun publishTaskMessage(event: String, task: String) {
        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        val json = """{"event": "$event", "task": "$task"}"""
        publish("users/$username/$env/task", json.toByteArray())
    }

    private fun publish(
        topic: String,
        payload: ByteArray
    ) {
        val isConnected = try { mqttAndroidClient.isConnected } catch (e: Exception) { false }
        Log.d(tag, "Attempting to publish to $topic. Client connected state: $isConnected")

        if (!isConnected) {
            Log.e(tag, "Cannot publish to $topic: MQTT client is not connected")
            return
        }

        val message = MqttMessage(payload)
        message.qos = 0
        message.isRetained = false

        try {
            mqttAndroidClient.publish(topic, message)
            Log.d(tag, "Successfully published to $topic: ${payload.size} bytes")
        } catch (e: Exception) {
            Log.e(tag, "Error publishing to $topic: ${e.message}")
            e.printStackTrace()
        }
    }
}
