package com.example.teraluxapp.ui.teralux

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.example.teraluxapp.data.repository.FakeTeraluxRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class RegisterTeraluxViewModelTest {
    
    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()
    
    private val testDispatcher = StandardTestDispatcher()
    private lateinit var repository: FakeTeraluxRepository
    private lateinit var viewModel: RegisterTeraluxViewModel
    
    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        repository = FakeTeraluxRepository()
        viewModel = RegisterTeraluxViewModel(repository)
    }
    
    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }
    
    @Test
    fun `registerDevice updates state to Loading then Success`() = runTest {
        // Given
        val macAddress = "AA:BB:CC:DD:EE:FF"
        val deviceName = "Test Device"
        
        // When
        viewModel.registerDevice(macAddress, "101", deviceName)
        testDispatcher.scheduler.advanceUntilIdle()
        
        // Then
        assertTrue(viewModel.uiState.value is RegisterUiState.Success)
    }
    
    @Test
    fun `initial state is Idle`() {
        // Then
        assertTrue(viewModel.uiState.value is RegisterUiState.Idle)
    }
    
    @Test
    fun `registerDevice with error updates state to Error`() = runTest {
        // Given
        val macAddress = "AA:BB:CC:DD:EE:FF"
        val deviceName = "Test Device"
        val errorMessage = "Network error"
        repository.setError(errorMessage)
        
        // When
        viewModel.registerDevice(macAddress, "101", deviceName)
        testDispatcher.scheduler.advanceUntilIdle()
        
        // Then
        val state = viewModel.uiState.value
        assertTrue(state is RegisterUiState.Error)
        assertEquals(errorMessage, (state as RegisterUiState.Error).message)
    }
}
