package com.example.whisperandroid.presentation.dashboard

import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.local.BackgroundAssistantModeStore
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.data.remote.dto.TuyaDeviceDto
import com.example.whisperandroid.data.remote.dto.TuyaDevicesResponseDto
import com.example.whisperandroid.domain.repository.TerminalRepository
import com.example.whisperandroid.domain.usecase.AuthenticateUseCase
import com.example.whisperandroid.domain.usecase.GetTuyaDevicesUseCase
import io.mockk.coEvery
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class DashboardViewModelTest {
    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        NetworkModule.setTuyaSyncReady(false)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `fetchDevices stores synced device list on success`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>()
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)

        val expectedDevices = listOf(
            TuyaDeviceDto(
                id = "dev-1",
                name = "Lampu Kamar",
                category = "dj",
                productName = "Lamp",
                online = true,
                icon = "",
                status = emptyList()
            )
        )

        coEvery { devicesUC.invoke() } returns Result.success(
            TuyaDevicesResponseDto(
                devices = expectedDevices,
                totalDevices = 1,
                currentPageCount = 1,
                page = 1,
                perPage = 20,
                total = 1
            )
        )
        io.mockk.every { modeStore.isEnabled } returns modeFlow
        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>(relaxed = true)
        val tuyaSyncReadyFlow = MutableStateFlow(false)

        val vm = DashboardViewModel(authUC, devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)
        vm.fetchDevices(force = true)
        tuyaSyncReadyFlow.value = true // Simulate the change we expect
        advanceUntilIdle()

        val state = vm.uiState.value
        assertTrue(state.isTuyaSyncReady)
        assertEquals(null, state.error)
    }

    @Test
    fun `fetchDevices keeps previous list on failure`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>()
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow

        val firstDevices = listOf(
            TuyaDeviceDto(
                id = "dev-1",
                name = "AC Kamar",
                category = "ac",
                productName = "AC",
                online = true,
                icon = "",
                status = emptyList()
            )
        )

        coEvery { devicesUC.invoke() } returnsMany listOf(
            Result.success(
                TuyaDevicesResponseDto(
                    devices = firstDevices,
                    totalDevices = 1,
                    currentPageCount = 1,
                    page = 1,
                    perPage = 20,
                    total = 1
                )
            ),
            Result.failure(Exception("sync failed"))
        )

        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>(relaxed = true)
        val tuyaSyncReadyFlow = MutableStateFlow(true)
        val vm = DashboardViewModel(authUC, devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)
        vm.fetchDevices(force = true)
        advanceUntilIdle()
        vm.fetchDevices(force = true)
        advanceUntilIdle()

        val state = vm.uiState.value
        assertEquals("sync failed", state.error)
    }

    @Test
    fun `updateAiProvider sets terminalNotFound error with redirect flag on 404`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>()
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow
        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>(relaxed = true)
        coEvery { tokenManager.getTerminalId() } returns "terminal-123"

        // Simulate 404 terminal not found error
        coEvery { terminalRepository.updateTerminal("terminal-123", "gemini") } returns Result.failure(
            Exception("Terminal not found")
        )

        val tuyaSyncReadyFlow = MutableStateFlow(true)
        val vm = DashboardViewModel(authUC, devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)

        vm.updateAiProvider("gemini")
        advanceUntilIdle()

        val state = vm.uiState.value
        assertEquals("Terminal not found. Please register your device.", state.error)
        assertEquals(true, state.shouldRedirectToRegister)
        assertEquals(false, state.isSavingAiProvider)
    }

    @Test
    fun `updateAiProvider does not re-trigger error on repeated calls when terminal not found`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>()
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow
        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>(relaxed = true)
        coEvery { tokenManager.getTerminalId() } returns "terminal-123"

        // First call fails with 404
        coEvery { terminalRepository.updateTerminal("terminal-123", "gemini") } returns Result.failure(
            Exception("Terminal not found")
        )

        val tuyaSyncReadyFlow = MutableStateFlow(true)
        val vm = DashboardViewModel(authUC, devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)

        // First call
        vm.updateAiProvider("gemini")
        advanceUntilIdle()

        val firstState = vm.uiState.value
        assertEquals("Terminal not found. Please register your device.", firstState.error)
        assertEquals(true, firstState.shouldRedirectToRegister)

        // Second call should not change the redirect flag (Toast guard prevents re-trigger)
        vm.updateAiProvider("gemini")
        advanceUntilIdle()

        val secondState = vm.uiState.value
        // The error should still be the terminal-not-found message
        assertEquals("Terminal not found. Please register your device.", secondState.error)
        // shouldRedirectToRegister should remain true (not re-triggered)
        assertEquals(true, secondState.shouldRedirectToRegister)
    }
}
