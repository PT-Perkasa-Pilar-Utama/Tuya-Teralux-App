package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.CreateTeraluxRequest
import com.example.teraluxapp.data.model.CreateTeraluxResponse
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class TeraluxRepositoryTest {
    
    private lateinit var repository: FakeTeraluxRepository
    
    @Before
    fun setup() {
        repository = FakeTeraluxRepository()
    }
    
    @Test
    fun `checkDeviceRegistration returns device when registered`() = runTest {
        // Given
        val macAddress = "AA:BB:CC:DD:EE:FF"
        repository.setRegisteredDevice(macAddress)
        
        // When
        val result = repository.checkDeviceRegistration(macAddress)
        
        // Then
        assertTrue(result.isSuccess)
        assertNotNull(result.getOrNull())
        assertEquals(macAddress, result.getOrNull()?.macAddress)
    }
    
    @Test
    fun `checkDeviceRegistration returns null when not registered`() = runTest {
        // Given
        val macAddress = "AA:BB:CC:DD:EE:FF"
        
        // When
        val result = repository.checkDeviceRegistration(macAddress)
        
        // Then
        assertTrue(result.isSuccess)
        assertNull(result.getOrNull())
    }
    
    @Test
    fun `registerDevice returns success when registration succeeds`() = runTest {
        // Given
        val macAddress = "AA:BB:CC:DD:EE:FF"
        val deviceName = "Test Device"
        
        // When
        val result = repository.registerDevice(macAddress, "101", deviceName)
        
        // Then
        assertTrue(result.isSuccess)
        assertNotNull(result.getOrNull())
    }
    
    @Test
    fun `registerDevice returns failure when error occurs`() = runTest {
        // Given
        val macAddress = "AA:BB:CC:DD:EE:FF"
        val deviceName = "Test Device"
        repository.setError("Network error")
        
        // When
        val result = repository.registerDevice(macAddress, "101", deviceName)
        
        // Then
        assertTrue(result.isFailure)
        assertEquals("Network error", result.exceptionOrNull()?.message)
    }
    
    @Test
    fun `checkDeviceRegistration handles multiple MAC addresses`() = runTest {
        // Given
        val mac1 = "AA:BB:CC:DD:EE:FF"
        val mac2 = "11:22:33:44:55:66"
        repository.setRegisteredDevice(mac1)
        
        // When
        val result1 = repository.checkDeviceRegistration(mac1)
        val result2 = repository.checkDeviceRegistration(mac2)
        
        // Then
        assertNotNull(result1.getOrNull())
        assertNull(result2.getOrNull())
    }
}
