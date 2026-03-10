package com.example.whisperandroid.util

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class MqttPayloadFactoryTest {
    @Test
    fun `buildChatPayload uses tuya uid and never requires mqtt username`() {
        val payload = MqttPayloadFactory.buildChatPayload(
            text = "nyalakan lampu",
            terminalId = "term-1",
            language = "id",
            requestId = "req-1",
            tuyaUid = "sg1765086176746IkwBD"
        )

        assertEquals("sg1765086176746IkwBD", payload.getString("uid"))
        assertEquals("term-1", payload.getString("terminal_id"))
        assertEquals("nyalakan lampu", payload.getString("prompt"))
    }

    @Test
    fun `buildChatPayload omits uid when tuya uid is unavailable`() {
        val payload = MqttPayloadFactory.buildChatPayload(
            text = "nyalakan lampu",
            terminalId = "term-1",
            language = "id",
            requestId = null,
            tuyaUid = null
        )

        assertFalse(payload.has("uid"))
    }

    @Test
    fun `buildAudioPayload preserves tuya uid when provided`() {
        val payload = MqttPayloadFactory.buildAudioPayload(
            base64Audio = "YWJj",
            terminalId = "term-1",
            language = "id",
            requestId = "req-2",
            tuyaUid = "sg1765086176746IkwBD"
        )

        assertTrue(payload.has("uid"))
        assertEquals("sg1765086176746IkwBD", payload.getString("uid"))
        assertEquals("YWJj", payload.getString("audio"))
    }
}
