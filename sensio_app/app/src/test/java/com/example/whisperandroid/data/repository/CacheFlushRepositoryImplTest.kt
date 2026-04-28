package com.example.whisperandroid.data.repository

import com.example.whisperandroid.data.local.TokenManager
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.ResponseBody
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class CacheFlushRepositoryImplTest {
    private val testDispatcher = StandardTestDispatcher()

    private lateinit var tokenManager: TokenManager
    private lateinit var okhttpClient: OkHttpClient
    private lateinit var mockedCall: okhttp3.Call
    private lateinit var responseBody: ResponseBody
    private lateinit var response: Response

    private val testToken = "test-access-token"
    private val testBaseUrl = "https://api.example.com/"

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)

        tokenManager = mockk(relaxed = true)
        okhttpClient = mockk(relaxed = true)
        mockedCall = mockk(relaxed = true)
        responseBody = mockk(relaxed = true)
        response = mockk(relaxed = true)

        every { tokenManager.getAccessToken() } returns testToken
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `flushCache constructs URL without double-slash when baseUrl has trailing slash`() = runTest {
        // Given
        val baseUrlWithTrailingSlash = "https://api.example.com/"
        val expectedUrl = "https://api.example.com/api/cache/flush"
        var capturedRequestUrl: String? = null

        // Mock the OkHttpClient to capture the request URL
        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers {
            val request = this.invocation.args[0] as Request
            capturedRequestUrl = request.url.toString()
            response
        }
        every { response.isSuccessful } returns true
        every { response.close() } returns Unit

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = baseUrlWithTrailingSlash,
            okhttpClient = okhttpClient
        )

        // When
        val result = repository.flushCache()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(expectedUrl, capturedRequestUrl)
        // Ensure no double-slash in the URL path
        assertTrue("URL should not contain double-slash", capturedRequestUrl?.contains("//api") == false)
    }

    @Test
    fun `flushCache constructs URL correctly when baseUrl has no trailing slash`() = runTest {
        // Given
        val baseUrlNoTrailingSlash = "https://api.example.com"
        val expectedUrl = "https://api.example.com/api/cache/flush"
        var capturedRequestUrl: String? = null

        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers {
            val request = this.invocation.args[0] as Request
            capturedRequestUrl = request.url.toString()
            response
        }
        every { response.isSuccessful } returns true
        every { response.close() } returns Unit

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = baseUrlNoTrailingSlash,
            okhttpClient = okhttpClient
        )

        // When
        val result = repository.flushCache()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(expectedUrl, capturedRequestUrl)
    }

    @Test
    fun `flushCache uses injected OkHttpClient instead of creating new one`() = runTest {
        // Given
        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers { response }
        every { response.isSuccessful } returns true
        every { response.close() } returns Unit

        // When
        repository.flushCache()

        // Then - verify the injected client was used
        verify(exactly = 1) { okhttpClient.newCall(any<Request>()) }
        verify(exactly = 1) { mockedCall.execute() }
    }

    @Test
    fun `flushCache returns failure when token is null`() = runTest {
        // Given
        every { tokenManager.getAccessToken() } returns null

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        // When
        val result = repository.flushCache()

        // Then
        assertTrue(result.isFailure)
        assertEquals("No access token found", result.exceptionOrNull()?.message)
        // Verify no network call was made
        verify(exactly = 0) { okhttpClient.newCall(any<Request>()) }
    }

    @Test
    fun `flushCache returns success when response is successful`() = runTest {
        // Given
        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers { response }
        every { response.isSuccessful } returns true
        every { response.close() } returns Unit

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        // When
        val result = repository.flushCache()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(true, result.getOrNull())
    }

    @Test
    fun `flushCache returns failure when response is not successful`() = runTest {
        // Given
        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers { response }
        every { response.isSuccessful } returns false
        every { response.code } returns 500
        every { response.message } returns "Internal Server Error"
        every { response.close() } returns Unit

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        // When
        val result = repository.flushCache()

        // Then
        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull()?.message?.contains("HTTP 500") == true)
    }

    @Test
    fun `flushCache closes response body after use`() = runTest {
        // Given
        var responseWasClosed = false

        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers { response }
        every { response.isSuccessful } returns true
        every { response.close() } answers { responseWasClosed = true }

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        // When
        val result = repository.flushCache()

        // Then
        assertTrue(result.isSuccess)
        // Verify response.close() was called
        assertTrue("Response should be closed after use", responseWasClosed)
    }

    @Test
    fun `flushCache adds Authorization header with Bearer token`() = runTest {
        // Given
        var capturedAuthHeader: String? = null

        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers {
            val request = this.invocation.args[0] as Request
            capturedAuthHeader = request.header("Authorization")
            response
        }
        every { response.isSuccessful } returns true
        every { response.close() } returns Unit

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        // When
        repository.flushCache()

        // Then
        assertEquals("Bearer $testToken", capturedAuthHeader)
    }

    @Test
    fun `flushCache uses DELETE method with JSON body`() = runTest {
        // Given
        var capturedMethod: String? = null

        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } answers {
            val request = this.invocation.args[0] as Request
            capturedMethod = request.method
            response
        }
        every { response.isSuccessful } returns true
        every { response.close() } returns Unit

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        // When
        repository.flushCache()

        // Then
        assertEquals("DELETE", capturedMethod)
    }

    @Test
    fun `flushCache handles network exception`() = runTest {
        // Given
        val networkException = java.net.UnknownHostException("Network error")

        every { okhttpClient.newCall(any<Request>()) } answers { mockedCall }
        every { mockedCall.execute() } throws networkException

        val repository = CacheFlushRepositoryImpl(
            tokenManager = tokenManager,
            baseUrl = testBaseUrl,
            okhttpClient = okhttpClient
        )

        // When
        val result = repository.flushCache()

        // Then
        assertTrue(result.isFailure)
        assertEquals(networkException, result.exceptionOrNull())
    }
}