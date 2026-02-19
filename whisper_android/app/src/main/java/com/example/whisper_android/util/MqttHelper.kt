package com.example.whisper_android.util

import android.content.Context
import android.util.Log
import com.example.whisper_android.BuildConfig
import info.mqtt.android.service.MqttAndroidClient
import org.eclipse.paho.client.mqttv3.DisconnectedBufferOptions
import org.eclipse.paho.client.mqttv3.IMqttActionListener
import org.eclipse.paho.client.mqttv3.IMqttDeliveryToken
import org.eclipse.paho.client.mqttv3.IMqttToken
import org.eclipse.paho.client.mqttv3.MqttCallback
import org.eclipse.paho.client.mqttv3.MqttConnectOptions
import org.eclipse.paho.client.mqttv3.MqttException
import org.eclipse.paho.client.mqttv3.MqttMessage
import java.util.UUID

class MqttHelper(
    context: Context,
) {
    private var mqttAndroidClient: MqttAndroidClient
    private val serverUri = BuildConfig.MQTT_BROKER_URL
    private val clientID = "WhisperAndroid_" + UUID.randomUUID().toString()
    private val username = BuildConfig.MQTT_USERNAME
    private val password = BuildConfig.MQTT_PASSWORD
    private val tag = "MqttHelper"
    var onMessageReceived: ((topic: String, message: String) -> Unit)? = null
    var onConnectionStatusChanged: ((status: MqttConnectionStatus) -> Unit)? = null

    enum class MqttConnectionStatus {
        DISCONNECTED,
        CONNECTING,
        CONNECTED,
        FAILED,
    }

    init {
        val appContext = context.applicationContext
        mqttAndroidClient = MqttAndroidClient(appContext, serverUri, clientID)
        mqttAndroidClient.setCallback(
            object : MqttCallback {
                override fun connectionLost(cause: Throwable?) {
                    Log.d(tag, "Connection lost: ${cause?.message}")
                    onConnectionStatusChanged?.invoke(MqttConnectionStatus.DISCONNECTED)
                }

                override fun messageArrived(
                    topic: String,
                    message: MqttMessage,
                ) {
                    Log.d(tag, "Message arrived: $topic -> $message")
                    onMessageReceived?.invoke(topic, message.toString())
                }

                override fun deliveryComplete(token: IMqttDeliveryToken?) {}
            },
        )
    }

    fun connect() {
        if (mqttAndroidClient.isConnected) return

        val mqttConnectOptions = MqttConnectOptions()
        mqttConnectOptions.isAutomaticReconnect = true
        mqttConnectOptions.isCleanSession = false
        mqttConnectOptions.keepAliveInterval = 30
        mqttConnectOptions.userName = username
        mqttConnectOptions.password = password.toCharArray()

        onConnectionStatusChanged?.invoke(MqttConnectionStatus.CONNECTING)
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

                        // Subscribe to chat and answer topics
                        subscribe("users/teralux/chat/answer")
                        subscribe("users/teralux/whisper/answer")
                        subscribe("users/teralux/chat")
                        onConnectionStatusChanged?.invoke(MqttConnectionStatus.CONNECTED)
                    }

                    override fun onFailure(
                        asyncActionToken: IMqttToken,
                        exception: Throwable,
                    ) {
                        Log.w(tag, "Failed to connect to: $serverUri")
                        exception.printStackTrace()
                        onConnectionStatusChanged?.invoke(MqttConnectionStatus.FAILED)
                    }
                },
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
                        exception: Throwable,
                    ) {
                        Log.e(tag, "Failed to subscribe to $topic")
                    }
                },
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
            onConnectionStatusChanged?.invoke(MqttConnectionStatus.DISCONNECTED)
        } catch (e: MqttException) {
            e.printStackTrace()
        }
    }

    fun publishAudio(
        payload: ByteArray,
        language: String = "id",
    ) {
        val base64Audio = android.util.Base64.encodeToString(payload, android.util.Base64.NO_WRAP)
        val json =
            """
            {
                "audio": "$base64Audio",
                "teralux_id": "tx-1",
                "language": "$language"
            }
            """.trimIndent()
        publish("users/teralux/whisper", json.toByteArray())
    }

    fun publishChat(
        text: String,
        language: String = "id",
    ) {
        val json =
            """
            {
                "prompt": "$text",
                "teralux_id": "tx-1",
                "language": "$language"
            }
            """.trimIndent()
        publish("users/teralux/chat", json.toByteArray())
    }

    private fun publish(
        topic: String,
        payload: ByteArray,
    ) {
        val isConnected = mqttAndroidClient.isConnected
        Log.d(tag, "Attempting to publish to $topic. Client connected state: $isConnected")

        val message = MqttMessage(payload)
        message.qos = 0
        message.isRetained = false

        try {
            mqttAndroidClient.publish(topic, message)
            Log.d(tag, "Successfully published to $topic: ${payload.size} bytes")
        } catch (e: MqttException) {
            Log.e(tag, "Error publishing to $topic: ${e.message}")
            e.printStackTrace()
        }
    }
}
