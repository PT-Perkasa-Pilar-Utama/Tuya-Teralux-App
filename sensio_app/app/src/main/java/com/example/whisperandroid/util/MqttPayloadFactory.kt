package com.example.whisperandroid.util

import org.json.JSONObject

object MqttPayloadFactory {
    fun buildChatPayload(
        text: String,
        terminalId: String,
        language: String,
        requestId: String?,
        tuyaUid: String?
    ): JSONObject {
        return JSONObject().apply {
            if (requestId != null) {
                put("request_id", requestId)
            }
            put("prompt", text)
            put("terminal_id", terminalId)
            put("language", language)
            if (!tuyaUid.isNullOrBlank()) {
                put("uid", tuyaUid)
            }
        }
    }

    fun buildAudioPayload(
        base64Audio: String,
        terminalId: String,
        language: String,
        requestId: String?,
        tuyaUid: String?
    ): JSONObject {
        return JSONObject().apply {
            put("audio", base64Audio)
            put("terminal_id", terminalId)
            put("language", language)
            if (requestId != null) {
                put("request_id", requestId)
            }
            if (!tuyaUid.isNullOrBlank()) {
                put("uid", tuyaUid)
            }
        }
    }
}
