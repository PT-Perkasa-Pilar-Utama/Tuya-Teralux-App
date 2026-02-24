package com.example.teraluxapp.ui.login

import android.content.Context
import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.example.teraluxapp.data.model.AuthResponse
import com.example.teraluxapp.data.model.Teralux
import com.example.teraluxapp.data.repository.AuthRepository
import com.example.teraluxapp.data.repository.TeraluxRepository
import com.example.teraluxapp.utils.DeviceInfoUtils
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@ExperimentalCoroutinesApi
class LoginViewModelTest {

    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()

    private val testDispatcher = StandardTestDispatcher()

    private lateinit var authRepository: AuthRepository
    private lateinit var teraluxRepository: TeraluxRepository
    private lateinit var context: Context
    private lateinit var viewModel: LoginViewModel

    @Before
    fun setUp() {
        Dispatchers.setMain(testDispatcher)
        authRepository = mockk()
        teraluxRepository = mockk()
        context = mockk(relaxed = true)

        mockkObject(DeviceInfoUtils)
        every { DeviceInfoUtils.getMacAddress(any()) } returns "00:00:00:00:00:00"

        viewModel = LoginViewModel(authRepository, teraluxRepository, context)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial state is Idle`() {
        assertTrue(viewModel.uiState.value is LoginUiState.Idle)
    }

    @Test
    fun `login success updates state to Success with token and uid`() = runTest {
        val authResponse = AuthResponse("test-token", 3600, "refresh-token", "test-uid")
        val teralux = Teralux("1", "test-mac", "test-room", "test-name", "2024-01-01", "2024-01-01")

        coEvery { authRepository.authenticate() } returns Result.success(authResponse)
        coEvery { teraluxRepository.checkDeviceRegistration(any()) } returns Result.success(teralux)

        viewModel.login()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertTrue(state is LoginUiState.Success)
        assertEquals(authResponse.accessToken, (state as LoginUiState.Success).token)
        assertEquals(authResponse.uid, state.uid)
    }

    @Test
    fun `login failure updates state to Error`() = runTest {
        val errorMessage = "Authentication failed"
        coEvery { authRepository.authenticate() } returns Result.failure(Exception(errorMessage))

        viewModel.login()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertTrue(state is LoginUiState.Error)
        assertEquals(errorMessage, (state as LoginUiState.Error).message)
    }
}
