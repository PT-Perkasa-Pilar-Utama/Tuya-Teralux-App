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
                    setupCallback()
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
    
    private fun setupCallback() {
        mqttAndroidClient.setCallback(object : MqttCallbackExtended {
            override fun connectComplete(reconnect: Boolean, serverURI: String) {
                Log.d(TAG, "Connection complete. Reconnect: $reconnect")
            }

            override fun connectionLost(cause: Throwable) {
                Log.d(TAG, "The Connection was lost.")
            }

            override fun messageArrived(topic: String, message: MqttMessage) {
                // Not used in send-only mode
            }

            override fun deliveryComplete(token: IMqttDeliveryToken) {}
        })
    }
    
    fun publishAudio(
        payload: ByteArray,
        teraluxId: String,
        uid: String,
        diarize: Boolean = false,
        language: String = "id"
    ) {
        val topic = "users/teralux/whisper"
        
        // Convert to JSON as expected by Backend SpeechTranscribeController
        val base64Audio = android.util.Base64.encodeToString(payload, android.util.Base64.NO_WRAP)
        val jsonPayload = """
            {
                "audio": "$base64Audio",
                "teralux_id": "$teraluxId",
                "uid": "$uid",
                "diarize": $diarize,
                "language": "$language"
            }
        """.trimIndent()

        val message = MqttMessage(jsonPayload.toByteArray())
        message.qos = 0 
        message.isRetained = false
        
        try {
            mqttAndroidClient.publish(topic, message)
            Log.d(TAG, "Published audio JSON: ${jsonPayload.length} chars (Audio: ${payload.size} bytes)")
        } catch (e: MqttException) {
            Log.e(TAG, "Error publishing audio: ${e.message}")
            e.printStackTrace()
        }
    }
}
