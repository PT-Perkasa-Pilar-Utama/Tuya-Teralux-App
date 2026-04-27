package com.example.whisperandroid

import android.content.Context
import android.provider.Settings
import com.example.whisperandroid.data.auth.AuthStateManager
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.remote.api.LoginData
import com.example.whisperandroid.data.remote.api.LoginRequest
import com.example.whisperandroid.data.remote.api.LoginResponse
import io.mockk.coEvery
import io.mockk.every
import io.mockk.mockk
import io.mockk.mockkStatic
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Response

class AuthFlowIntegrationTest {

    private lateinit var commonApi: CommonApi
    private lateinit var context: Context
    private lateinit var tokenManager: TokenManager

    private val testAndroidId = "1234567890abcdef"
    private val testTerminalId = "12345678-90ab-cdef-0000-000000000000"

    @Before
    fun setup() {
        commonApi = mockk(relaxed = true)
        context = mockk()
        tokenManager = mockk(relaxed = true)

        // Mock Settings.Secure.getString for android ID
        mockkStatic("android.provider.Settings")
        every { Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID) } returns testAndroidId
        AuthStateManager.init(commonApi, context)
    }

    @Test
    fun `validateAuthWithBackend returns Authenticated on 200 success`() = kotlinx.coroutines.runBlocking {
        val loginResponse = LoginResponse(
            status = true,
            message = "Login successful",
            data = LoginData(
                terminal_id = testTerminalId,
                access_token = "valid_token",
                message = null
            )
        )
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.success(loginResponse)

        val result = AuthStateManager.validateAuthWithBackend()

        assertTrue(result.isSuccess)
        assertEquals(AuthStateManager.AuthState.Authenticated, result.getOrNull())
    }

    @Test
    fun `validateAuthWithBackend returns NotRegistered on 404`() = kotlinx.coroutines.runBlocking {
        val errorBody = "Not found".toResponseBody("text/plain".toMediaType())
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.error(404, errorBody)

        val result = AuthStateManager.validateAuthWithBackend()

        assertTrue(result.isSuccess)
        assertEquals(AuthStateManager.AuthState.NotRegistered, result.getOrNull())
    }

    @Test
    fun `validateAuthWithBackend returns Unauthorized on 401`() = kotlinx.coroutines.runBlocking {
        val errorBody = "Unauthorized".toResponseBody("text/plain".toMediaType())
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.error(401, errorBody)

        val result = AuthStateManager.validateAuthWithBackend()

        assertTrue(result.isSuccess)
        assertEquals(AuthStateManager.AuthState.Unauthorized, result.getOrNull())
    }

    @Test
    fun `validateAuthWithBackend returns Error on 5xx server error`() = kotlinx.coroutines.runBlocking {
        val errorBody = "Internal server error".toResponseBody("text/plain".toMediaType())
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.error(503, errorBody)

        val result = AuthStateManager.validateAuthWithBackend()

        assertTrue(result.isSuccess)
        assertEquals(AuthStateManager.AuthState.Error, result.getOrNull())
    }

    @Test
    fun `validateAuthWithBackend returns Error on network exception`() = kotlinx.coroutines.runBlocking {
        coEvery { commonApi.login(any<LoginRequest>()) } throws java.net.UnknownHostException("No internet connection")

        val result = AuthStateManager.validateAuthWithBackend()

        assertTrue(result.isFailure)
    }

    @Test
    fun `after registration, validateAuthWithBackend returns Authenticated`() = kotlinx.coroutines.runBlocking {
        val loginResponse = LoginResponse(
            status = true,
            message = "Login successful",
            data = LoginData(
                terminal_id = testTerminalId,
                access_token = "post_reg_token",
                message = null
            )
        )
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.success(loginResponse)

        val result = AuthStateManager.validateAuthWithBackend()

        assertTrue(result.isSuccess)
        assertEquals(AuthStateManager.AuthState.Authenticated, result.getOrNull())
    }
}