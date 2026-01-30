package com.example.teraluxapp.utils

import android.content.Context
import android.util.Log

import org.eclipse.paho.android.service.MqttAndroidClient
import org.eclipse.paho.client.mqttv3.*
import java.util.UUID

class MqttHelper(context: Context) {
    private var mqttAndroidClient: MqttAndroidClient
    private val serverUri = "wss://ws.farismnrr.com:443/mqtt" // Fixed MQTT Broker URL
    private val clientID = "Android_" + UUID.randomUUID().toString()
    private val username = "teralux"
    private val password = "REDACTED_SECRET"
    private val TAG = "MqttHelper"

    init {
        mqttAndroidClient = MqttAndroidClient(context, serverUri, clientID)
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
                    subscribeToTopic()
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
    
    // Callback for incoming messages
    var onMessageArrived: ((topic: String, message: String) -> Unit)? = null

    private fun subscribeToTopic() {
        val topic = "users/teralux/whisper"
        try {
            mqttAndroidClient.subscribe(topic, 0, null, object : IMqttActionListener {
                override fun onSuccess(asyncActionToken: IMqttToken) {
                    Log.d(TAG, "Subscribed to $topic")
                }

                override fun onFailure(asyncActionToken: IMqttToken, exception: Throwable) {
                    Log.w(TAG, "Failed to subscribe to $topic")
                }
            })

            mqttAndroidClient.setCallback(object : MqttCallbackExtended {
                override fun connectComplete(reconnect: Boolean, serverURI: String) {
                   if (reconnect) {
                       subscribeToTopic() // Re-subscribe on reconnect
                   }
                }

                override fun connectionLost(cause: Throwable) {
                    Log.d(TAG, "The Connection was lost.")
                }

                override fun messageArrived(topic: String, message: MqttMessage) {
                    Log.d(TAG, "Incoming message: " + String(message.payload))
                    onMessageArrived?.invoke(topic, String(message.payload))
                }

                override fun deliveryComplete(token: IMqttDeliveryToken) {}
            })
        } catch (ex: MqttException) {
            System.err.println("Exception creating subscribe")
            ex.printStackTrace()
        }
    }
    
    fun publishAudio(payload: ByteArray) {
        val topic = "users/teralux/whisper"
        val message = MqttMessage(payload)
        message.qos = 0 // Fire and forget for audio
        message.isRetained = false
        
        try {
            mqttAndroidClient.publish(topic, message)
            Log.d(TAG, "Published audio chunk: ${payload.size} bytes")
        } catch (e: MqttException) {
            Log.e(TAG, "Error publishing audio: ${e.message}")
            e.printStackTrace()
        }
    }
}
