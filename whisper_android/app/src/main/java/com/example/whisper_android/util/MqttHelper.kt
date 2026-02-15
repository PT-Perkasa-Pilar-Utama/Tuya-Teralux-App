package com.example.whisper_android.util

import android.content.Context
import android.util.Log
import com.example.whisper_android.BuildConfig
import org.eclipse.paho.android.service.MqttAndroidClient
import org.eclipse.paho.client.mqttv3.*
import java.util.UUID

class MqttHelper(context: Context) {
    private var mqttAndroidClient: MqttAndroidClient
    private val serverUri = BuildConfig.MQTT_BROKER_URL
    private val clientID = "WhisperAndroid_" + UUID.randomUUID().toString()
    private val username = BuildConfig.MQTT_USERNAME
    private val password = BuildConfig.MQTT_PASSWORD
    private val TAG = "MqttHelper"
    var onMessageReceived: ((topic: String, message: String) -> Unit)? = null

    init {
        val appContext = context.applicationContext
        mqttAndroidClient = MqttAndroidClient(appContext, serverUri, clientID)
        mqttAndroidClient.setCallback(object : MqttCallback {
            override fun connectionLost(cause: Throwable?) {
                Log.d(TAG, "Connection lost: ${cause?.message}")
            }

            override fun messageArrived(topic: String, message: MqttMessage) {
                Log.d(TAG, "Message arrived: $topic -> ${message.toString()}")
                onMessageReceived?.invoke(topic, message.toString())
            }

            override fun deliveryComplete(token: IMqttDeliveryToken?) {}
        })
    }

    fun connect() {
        if (mqttAndroidClient.isConnected) return

        val mqttConnectOptions = MqttConnectOptions()
        mqttConnectOptions.isAutomaticReconnect = true
        mqttConnectOptions.isCleanSession = false
        mqttConnectOptions.userName = username
        mqttConnectOptions.password = password.toCharArray()

        try {
            mqttAndroidClient.connect(mqttConnectOptions, null, object : IMqttActionListener {
                override fun onSuccess(asyncActionToken: IMqttToken) {
                    val disconnectedBufferOptions = DisconnectedBufferOptions()
                    disconnectedBufferOptions.isBufferEnabled = true
                    disconnectedBufferOptions.bufferSize = 100
                    disconnectedBufferOptions.isPersistBuffer = false
                    disconnectedBufferOptions.isDeleteOldestMessages = false
                    mqttAndroidClient.setBufferOpts(disconnectedBufferOptions)
                    Log.d(TAG, "Success Connected to $serverUri")
                    
                    // Subscribe to chat topic
                    subscribe("users/teralux/chat/answer")
                }

                override fun onFailure(asyncActionToken: IMqttToken, exception: Throwable) {
                    Log.w(TAG, "Failed to connect to: $serverUri")
                    exception.printStackTrace()
                }
            })
        } catch (ex: MqttException) {
            ex.printStackTrace()
        }
    }

    private fun subscribe(topic: String) {
        try {
            mqttAndroidClient.subscribe(topic, 0, null, object : IMqttActionListener {
                override fun onSuccess(asyncActionToken: IMqttToken) {
                    Log.d(TAG, "Subscribed to $topic")
                }

                override fun onFailure(asyncActionToken: IMqttToken, exception: Throwable) {
                    Log.e(TAG, "Failed to subscribe to $topic")
                }
            })
        } catch (e: MqttException) {
            e.printStackTrace()
        }
    }

    fun disconnect() {
        if (!mqttAndroidClient.isConnected) return
        try {
            mqttAndroidClient.disconnect()
            Log.d(TAG, "Disconnected from MQTT")
        } catch (e: MqttException) {
            e.printStackTrace()
        }
    }

    fun publishAudio(payload: ByteArray) {
        publish("users/teralux/whisper", payload)
    }

    fun publishChat(text: String) {
        publish("users/teralux/chat", text.toByteArray())
    }

    private fun publish(topic: String, payload: ByteArray) {
        if (!mqttAndroidClient.isConnected) {
            Log.w(TAG, "Client not connected, skipping publish to $topic")
            return
        }

        val message = MqttMessage(payload)
        message.qos = 0
        message.isRetained = false
        
        try {
            mqttAndroidClient.publish(topic, message)
            Log.d(TAG, "Published to $topic: ${payload.size} bytes")
        } catch (e: MqttException) {
            Log.e(TAG, "Error publishing to $topic: ${e.message}")
            e.printStackTrace()
        }
    }
}
