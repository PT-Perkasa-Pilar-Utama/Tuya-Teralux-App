package com.example.teraluxapp.ui.room

import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject

@HiltViewModel
class RoomStatusViewModel @Inject constructor() : ViewModel() {
    
    private val _uiState = MutableStateFlow<RoomStatusUiState>(RoomStatusUiState.Loading)
    val uiState: StateFlow<RoomStatusUiState> = _uiState.asStateFlow()
    
    private val _showPasswordDialog = MutableStateFlow(false)
    val showPasswordDialog: StateFlow<Boolean> = _showPasswordDialog.asStateFlow()
    
    private val _passwordError = MutableStateFlow<String?>(null)
    val passwordError: StateFlow<String?> = _passwordError.asStateFlow()
    
    init {
        loadRoomStatus()
    }
    
    private fun loadRoomStatus() {
        // For now, hardcoded data - will integrate with API later
        _uiState.value = RoomStatusUiState.Success(
            roomName = "ALAMANDA ROOM",
            status = RoomStatus.OCCUPIED,
            bookingInfo = BookingInfo(
                guestName = "Mr. Abdul Azis",
                company = "PT Berlian Utama Jaya",
                timeRange = "08:00 AM - 10:00 AM"
            ),
            date = "Thursday, Dec 18th, 2025"
        )
    }
    
    fun onDashboardClick() {
        _showPasswordDialog.value = true
        _passwordError.value = null
    }
    
    fun onPasswordDialogDismiss() {
        _showPasswordDialog.value = false
        _passwordError.value = null
    }
    
    fun verifyPassword(password: String): Boolean {
        return if (password == "admin") {
            _showPasswordDialog.value = false
            _passwordError.value = null
            true
        } else {
            _passwordError.value = "Wrong password"
            false
        }
    }
}
