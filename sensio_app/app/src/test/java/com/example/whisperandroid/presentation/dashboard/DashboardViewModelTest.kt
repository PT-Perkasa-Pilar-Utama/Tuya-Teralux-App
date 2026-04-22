package com.example.whisperandroid.presentation.dashboard

import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.local.BackgroundAssistantModeStore
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.data.remote.dto.TuyaDeviceDto
import com.example.whisperandroid.data.remote.dto.TuyaDevicesResponseDto
import com.example.whisperandroid.domain.repository.AiEngineProfileState
import com.example.whisperandroid.domain.repository.TerminalRepository
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

        val vm = DashboardViewModel(devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)
        vm.fetchDevices(force = true)
        tuyaSyncReadyFlow.value = true // Simulate the change we expect
        advanceUntilIdle()

        val state = vm.uiState.value
        assertTrue(state.isTuyaSyncReady)
        assertEquals(null, state.error)
    }

    @Test
    fun `fetchDevices keeps previous list on failure`() = runTest {
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
        val vm = DashboardViewModel(devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)
        vm.fetchDevices(force = true)
        advanceUntilIdle()
        vm.fetchDevices(force = true)
        advanceUntilIdle()

        val state = vm.uiState.value
        assertEquals("sync failed", state.error)
    }

    @Test
    fun `loadCurrentAiEngineProfile sets profile when source is engine_profile`() = runTest {
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>(relaxed = true)
        val terminalRepository = mockk<TerminalRepository>()
        val tokenManager = mockk<TokenManager>(relaxed = true)
        val tuyaSyncReadyFlow = MutableStateFlow(true)
        
        io.mockk.every { modeStore.isEnabled } returns MutableStateFlow(false)
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC"
        
        coEvery { terminalRepository.getAiEngineProfileByMac("AA:BB:CC") } returns Result.success(
            AiEngineProfileState(
                profile = "fast",
                source = "engine_profile",
                effectiveProvider = null,
                effectiveMode = "fast"
            )
        )

        val vm = DashboardViewModel(devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)
        vm.loadCurrentAiEngineProfile()
        advanceUntilIdle()

        val state = vm.uiState.value
        assertEquals("fast", state.aiEngineProfile)
        assertEquals(null, state.legacyMigrationWarning)
    }

    @Test
    fun `loadCurrentAiEngineProfile sets warning when source is legacy_provider`() = runTest {
        val devicesUC = mockk<GetTuyaDevicesUseCase>(relaxed = true)
        val modeStore = mockk<BackgroundAssistantModeStore>(relaxed = true)
        val terminalRepository = mockk<TerminalRepository>()
        val tokenManager = mockk<TokenManager>(relaxed = true)
        val tuyaSyncReadyFlow = MutableStateFlow(true)
        
        io.mockk.every { modeStore.isEnabled } returns MutableStateFlow(false)
        io.mockk.every { tokenManager.getMacAddress() } returns "AA:BB:CC"
        
        coEvery { terminalRepository.getAiEngineProfileByMac("AA:BB:CC") } returns Result.success(
            AiEngineProfileState(
                profile = null,
                source = "legacy_provider",
                effectiveProvider = "openai",
                effectiveMode = "legacy"
            )
        )

        val vm = DashboardViewModel(devicesUC, modeStore, terminalRepository, tokenManager, tuyaSyncReadyFlow)
        vm.loadCurrentAiEngineProfile()
        advanceUntilIdle()

        val state = vm.uiState.value
        assertEquals(null, state.aiEngineProfile)
        assertTrue(state.legacyMigrationWarning != null)
        assertTrue(state.legacyMigrationWarning!!.contains("legacy AI provider"))
    }
}
