package com.example.whisperandroid.data.repository

import android.content.Context
import android.provider.Settings
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.remote.api.LoginData
import com.example.whisperandroid.data.remote.api.LoginRequest
import com.example.whisperandroid.data.remote.api.LoginResponse
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import io.mockk.mockkStatic
import kotlinx.coroutines.runBlocking
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Response

/**
 * Unit tests for LoginRepositoryImpl.
 */
class LoginRepositoryImplTest {

    private lateinit var commonApi: CommonApi
    private lateinit var context: Context
    private lateinit var tokenManager: TokenManager
    private lateinit var repository: LoginRepositoryImpl

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

        repository = LoginRepositoryImpl(commonApi, context, tokenManager)
    }

    @Test
    fun `login returns Authenticated on 200 success with new access_token`() = runBlocking {
        // Given
        val loginData = LoginData(
            terminal_id = testTerminalId,
            access_token = "new_access_token_123",
            message = "Login successful"
        )
        val loginResponse = LoginResponse(
            status = true,
            message = "Login successful",
            data = loginData
        )
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.success(loginResponse)

        val result = repository.login()

        assertTrue(result.isSuccess)
        assertEquals(LoginRepositoryImpl.AuthState.Authenticated, result.getOrNull())
        coVerify { tokenManager.saveAccessToken("new_access_token_123") }
        coVerify { tokenManager.saveTerminalId(testTerminalId) }
    }

    @Test
    fun `login returns Authenticated on 200 success with null access_token`() = runBlocking {
        // Given
        val loginResponse = LoginResponse(
            status = true,
            message = "Token still valid",
            data = LoginData(
                terminal_id = testTerminalId,
                access_token = null,
                message = "Token still valid"
            )
        )
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.success(loginResponse)

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(LoginRepositoryImpl.AuthState.Authenticated, result.getOrNull())
        // Should save terminal ID but not overwrite existing token
        coVerify { tokenManager.saveTerminalId(testTerminalId) }
    }

    @Test
    fun `login returns Authenticated on 200 success with empty access_token`() = runBlocking {
        // Given
        val loginResponse = LoginResponse(
            status = true,
            message = "Token still valid",
            data = LoginData(
                terminal_id = testTerminalId,
                access_token = "",
                message = "Token still valid"
            )
        )
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.success(loginResponse)

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(LoginRepositoryImpl.AuthState.Authenticated, result.getOrNull())
        // Should save terminal ID but not overwrite existing token
        coVerify { tokenManager.saveTerminalId(testTerminalId) }
    }

    @Test
    fun `login returns NotRegistered on 404 response`() = runBlocking {
        // Given
        val errorBody = "Not found".toResponseBody("text/plain".toMediaType())
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.error(404, errorBody)

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(LoginRepositoryImpl.AuthState.NotRegistered, result.getOrNull())
    }

    @Test
    fun `login returns Unauthorized on 401 response`() = runBlocking {
        // Given
        val errorBody = "Unauthorized".toResponseBody("text/plain".toMediaType())
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.error(401, errorBody)

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(LoginRepositoryImpl.AuthState.Unauthorized, result.getOrNull())
    }

    @Test
    fun `login returns Error on 503 service unavailable`() = runBlocking {
        // Given
        val errorBody = "Service unavailable".toResponseBody("text/plain".toMediaType())
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.error(503, errorBody)

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        val authState = result.getOrNull()
        assertTrue(authState is LoginRepositoryImpl.AuthState.Error)
        assertEquals("Authentication service unavailable", (authState as LoginRepositoryImpl.AuthState.Error).message)
    }

    @Test
    fun `login returns Error on 500 server error`() = runBlocking {
        // Given
        val errorBody = "Internal server error".toResponseBody("text/plain".toMediaType())
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.error(500, errorBody)

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        val authState = result.getOrNull()
        assertTrue(authState is LoginRepositoryImpl.AuthState.Error)
        assertEquals("Server error: 500", (authState as LoginRepositoryImpl.AuthState.Error).message)
    }

    @Test
    fun `login returns Error on network exception`() = runBlocking {
        // Given
        coEvery { commonApi.login(any<LoginRequest>()) } throws java.net.UnknownHostException("No internet connection")

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        val authState = result.getOrNull()
        assertTrue(authState is LoginRepositoryImpl.AuthState.Error)
        assertTrue((authState as LoginRepositoryImpl.AuthState.Error).message.startsWith("Network error:"))
    }

    @Test
    fun `login returns Error on timeout exception`() = runBlocking {
        // Given
        coEvery { commonApi.login(any<LoginRequest>()) } throws java.net.SocketTimeoutException("Connection timed out")

        // When
        val result = repository.login()

        // Then
        assertTrue(result.isSuccess)
        val authState = result.getOrNull()
        assertTrue(authState is LoginRepositoryImpl.AuthState.Error)
        assertTrue((authState as LoginRepositoryImpl.AuthState.Error).message.startsWith("Network error:"))
    }

    @Test
    fun `autoLogin delegates to login`() = runBlocking {
        // Given
        val loginResponse = LoginResponse(
            status = true,
            message = "Auto login successful",
            data = LoginData(
                terminal_id = testTerminalId,
                access_token = "auto_login_token",
                message = null
            )
        )
        coEvery { commonApi.login(any<LoginRequest>()) } returns Response.success(loginResponse)

        // When
        val result = repository.autoLogin()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(LoginRepositoryImpl.AuthState.Authenticated, result.getOrNull())
    }
}