package com.example.whisperandroid.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import io.mockk.every
import io.mockk.mockk
import io.mockk.mockkStatic
import io.mockk.verify
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

class TokenManagerTest {

    private lateinit var mockPrefs: SharedPreferences
    private lateinit var mockEditor: SharedPreferences.Editor
    private lateinit var tokenManager: TokenManager

    @Before
    fun setup() {
        mockPrefs = mockk()
        mockEditor = mockk(relaxed = true)

        mockkStatic("androidx.core.content.editKt")
        every { mockPrefs.edit(any<Boolean>(), any<SharedPreferences.Editor.() -> Unit>()) } answers {
            val action = lastArg<SharedPreferences.Editor.() -> Unit>()
            action(mockEditor)
        }

        val mockContext = mockk<Context>(relaxed = true)
        every { mockContext.getSharedPreferences("sensio_prefs", Context.MODE_PRIVATE) } returns mockPrefs

        every { mockPrefs.getString("access_token", null) } returns null
        every { mockPrefs.getLong("token_expires_at", 0L) } returns 0L
        every { mockPrefs.getString("terminal_id", null) } returns null

        tokenManager = TokenManager(mockContext)
    }

    @Test
    fun `saveAccessToken stores token in SharedPreferences`() {
        val token = "test_token_123"

        tokenManager.saveAccessToken(token)

        verify { mockEditor.putString("access_token", token) }
    }

    @Test
    fun `getAccessToken returns stored token`() {
        every { mockPrefs.getString("access_token", null) } returns "stored_token"

        val result = tokenManager.getAccessToken()

        assertEquals("stored_token", result)
    }

    @Test
    fun `getAccessToken returns null when no token stored`() {
        every { mockPrefs.getString("access_token", null) } returns null

        val result = tokenManager.getAccessToken()

        assertNull(result)
    }

    @Test
    fun `isTokenExpired returns true when no token exists`() {
        every { mockPrefs.getString("access_token", null) } returns null

        val result = tokenManager.isTokenExpired()

        assertTrue(result)
    }

    @Test
    fun `isTokenExpired returns true when token is expired`() {
        val expiredTime = System.currentTimeMillis() - 1000
        every { mockPrefs.getString("access_token", null) } returns "some_token"
        every { mockPrefs.getLong("token_expires_at", 0L) } returns expiredTime

        val result = tokenManager.isTokenExpired()

        assertTrue(result)
    }

    @Test
    fun `isTokenExpired returns false when token is still valid`() {
        val futureTime = System.currentTimeMillis() + 3600000
        every { mockPrefs.getString("access_token", null) } returns "valid_token"
        every { mockPrefs.getLong("token_expires_at", 0L) } returns futureTime

        val result = tokenManager.isTokenExpired()

        assertFalse(result)
    }

    @Test
    fun `clearToken removes all sensitive data`() {
        tokenManager.clearToken()

        verify { mockEditor.remove("access_token") }
        verify { mockEditor.remove("tuya_uid") }
        verify { mockEditor.remove("terminal_id") }
        verify { mockEditor.remove("mac_address") }
    }

    @Test
    fun `saveTerminalId stores terminal ID in SharedPreferences`() {
        val terminalId = "term_abc_123"

        tokenManager.saveTerminalId(terminalId)

        verify { mockEditor.putString("terminal_id", terminalId) }
    }

    @Test
    fun `getTerminalId returns stored terminal ID`() {
        every { mockPrefs.getString("terminal_id", null) } returns "stored_terminal"

        val result = tokenManager.getTerminalId()

        assertEquals("stored_terminal", result)
    }

    @Test
    fun `getTerminalId returns null when no terminal ID stored`() {
        every { mockPrefs.getString("terminal_id", null) } returns null

        val result = tokenManager.getTerminalId()

        assertNull(result)
    }
}