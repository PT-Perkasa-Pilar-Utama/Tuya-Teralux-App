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
import io.mockk.coVerify
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
    fun `resetAiProvider_whenFlagFalseAndProviderPresent_callsUpdateWithEmptyString`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow

        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>()
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC:DD:EE:FF"
        io.mockk.every { tokenManager.getTerminalId() } returns "terminal-123"

        val mockRegistration = mockk<com.example.whisperandroid.domain.model.TerminalRegistration>()
        io.mockk.every { mockRegistration.aiProvider } returns "gemini"

        coEvery { terminalRepository.getTerminalByMac("AA:BB:CC:DD:EE:FF") } returns Result.success(mockRegistration)
        coEvery { terminalRepository.updateTerminal("terminal-123", "") } returns Result.success(Unit)

        val tuyaSyncReadyFlow = MutableStateFlow(false)

        val vm = DashboardViewModel(
            authUC,
            devicesUC,
            modeStore,
            terminalRepository,
            tokenManager,
            tuyaSyncReadyFlow,
            isAiEngineSelectorVisible = false
        )
        advanceUntilIdle()

        coVerify(exactly = 1) { terminalRepository.updateTerminal("terminal-123", "") }
        assertEquals("", vm.uiState.value.aiProvider)
    }

    @Test
    fun `resetAiProvider_whenFlagFalseAndProviderEmpty_doesNotCallUpdate`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow

        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>()
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC:DD:EE:FF"
        io.mockk.every { tokenManager.getTerminalId() } returns "terminal-123"

        val mockRegistration = mockk<com.example.whisperandroid.domain.model.TerminalRegistration>()
        io.mockk.every { mockRegistration.aiProvider } returns null

        coEvery { terminalRepository.getTerminalByMac("AA:BB:CC:DD:EE:FF") } returns Result.success(mockRegistration)

        val tuyaSyncReadyFlow = MutableStateFlow(false)

        val vm = DashboardViewModel(
            authUC,
            devicesUC,
            modeStore,
            terminalRepository,
            tokenManager,
            tuyaSyncReadyFlow,
            isAiEngineSelectorVisible = false
        )
        advanceUntilIdle()

        coVerify(exactly = 0) { terminalRepository.updateTerminal(any(), any()) }
    }

    @Test
    fun `resetAiProvider_whenFlagTrue_doesNotCallUpdate`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow

        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>()
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC:DD:EE:FF"
        io.mockk.every { tokenManager.getTerminalId() } returns "terminal-123"

        val mockRegistration = mockk<com.example.whisperandroid.domain.model.TerminalRegistration>()
        io.mockk.every { mockRegistration.aiProvider } returns "openai"

        coEvery { terminalRepository.getTerminalByMac("AA:BB:CC:DD:EE:FF") } returns Result.success(mockRegistration)

        val tuyaSyncReadyFlow = MutableStateFlow(false)

        val vm = DashboardViewModel(
            authUC,
            devicesUC,
            modeStore,
            terminalRepository,
            tokenManager,
            tuyaSyncReadyFlow,
            isAiEngineSelectorVisible = true
        )
        advanceUntilIdle()

        coVerify(exactly = 0) { terminalRepository.updateTerminal(any(), any()) }
        assertEquals("openai", vm.uiState.value.aiProvider)
    }

    @Test
    fun `resetAiProvider_onFailure_doesNotCrashOrShowError`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow

        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>()
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC:DD:EE:FF"
        io.mockk.every { tokenManager.getTerminalId() } returns "terminal-123"

        val mockRegistration = mockk<com.example.whisperandroid.domain.model.TerminalRegistration>()
        io.mockk.every { mockRegistration.aiProvider } returns "gemini"

        coEvery { terminalRepository.getTerminalByMac("AA:BB:CC:DD:EE:FF") } returns Result.success(mockRegistration)
        coEvery { terminalRepository.updateTerminal("terminal-123", "") } returns Result.failure(Exception("Network error"))

        val tuyaSyncReadyFlow = MutableStateFlow(false)

        val vm = DashboardViewModel(
            authUC,
            devicesUC,
            modeStore,
            terminalRepository,
            tokenManager,
            tuyaSyncReadyFlow,
            isAiEngineSelectorVisible = false
        )
        advanceUntilIdle()

        // Verify no error in UI state
        assertEquals(null, vm.uiState.value.error)
        // Provider remains unchanged
        assertEquals("gemini", vm.uiState.value.aiProvider)
    }

    @Test
    fun `resetAiProvider_calledTwice_onlyExecutesOnce`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow

        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>()
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC:DD:EE:FF"
        io.mockk.every { tokenManager.getTerminalId() } returns "terminal-123"

        val mockRegistration = mockk<com.example.whisperandroid.domain.model.TerminalRegistration>()
        io.mockk.every { mockRegistration.aiProvider } returns "gemini"

        coEvery { terminalRepository.getTerminalByMac("AA:BB:CC:DD:EE:FF") } returns Result.success(mockRegistration)
        coEvery { terminalRepository.updateTerminal("terminal-123", "") } returns Result.success(Unit)

        val tuyaSyncReadyFlow = MutableStateFlow(false)

        val vm = DashboardViewModel(
            authUC,
            devicesUC,
            modeStore,
            terminalRepository,
            tokenManager,
            tuyaSyncReadyFlow,
            isAiEngineSelectorVisible = false
        )
        advanceUntilIdle()

        // Manually trigger load again to simulate calling twice
        vm.loadCurrentAiProvider()
        advanceUntilIdle()

        // Verify updateTerminal was called exactly once
        coVerify(exactly = 1) { terminalRepository.updateTerminal("terminal-123", "") }
    }

    @Test
    fun `resetAiProvider_afterSuccess_uiStateReflectsClearedProvider`() = runTest {
        val authUC = mockk<AuthenticateUseCase>(relaxed = true)
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>()
        val modeFlow = MutableStateFlow(false)
        io.mockk.every { modeStore.isEnabled } returns modeFlow

        val terminalRepository = mockk<TerminalRepository>(relaxed = true)
        val tokenManager = mockk<TokenManager>()
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC:DD:EE:FF"
        io.mockk.every { tokenManager.getTerminalId() } returns "terminal-123"

        val mockRegistration = mockk<com.example.whisperandroid.domain.model.TerminalRegistration>()
        io.mockk.every { mockRegistration.aiProvider } returns "openai"

        coEvery { terminalRepository.getTerminalByMac("AA:BB:CC:DD:EE:FF") } returns Result.success(mockRegistration)
        coEvery { terminalRepository.updateTerminal("terminal-123", "") } returns Result.success(Unit)

        val tuyaSyncReadyFlow = MutableStateFlow(false)

        val vm = DashboardViewModel(
            authUC,
            devicesUC,
            modeStore,
            terminalRepository,
            tokenManager,
            tuyaSyncReadyFlow,
            isAiEngineSelectorVisible = false
        )
        advanceUntilIdle()

        assertEquals("", vm.uiState.value.aiProvider)
    }
}
