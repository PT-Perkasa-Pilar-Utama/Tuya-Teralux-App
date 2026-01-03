package com.example.teraluxapp.ui.room

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.teraluxapp.data.network.RetrofitClient
import com.example.teraluxapp.utils.PreferencesManager
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*
import javax.inject.Inject

@HiltViewModel
class RoomStatusViewModel @Inject constructor(
    @ApplicationContext private val context: Context
) : ViewModel() {
    
    private val _uiState = MutableStateFlow<RoomStatusUiState>(RoomStatusUiState.Loading)
    val uiState: StateFlow<RoomStatusUiState> = _uiState.asStateFlow()
    
    private val _showPasswordDialog = MutableStateFlow(false)
    val showPasswordDialog: StateFlow<Boolean> = _showPasswordDialog.asStateFlow()
    
    private val _passwordError = MutableStateFlow<String?>(null)
    val passwordError: StateFlow<String?> = _passwordError.asStateFlow()
    
    private var currentToken: String = ""
    private var currentUid: String = ""
    private var roomId: String? = null
    
    init {
        loadRoomStatus()
    }
    
    fun setCredentials(token: String, uid: String) {
        currentToken = token
        currentUid = uid
        loadRoomStatus()
    }
    
    private fun loadRoomStatus() {
        viewModelScope.launch {
            try {
                // Fetch Teralux data to get name and room ID
                val teraluxId = PreferencesManager.getTeraluxId(context)
                var teraluxName: String? = null
                if (teraluxId != null && currentToken.isNotEmpty()) {
                    val response = RetrofitClient.instance.getTeraluxById("Bearer $currentToken", teraluxId)
                    if (response.isSuccessful && response.body() != null) {
                        val teraluxData = response.body()!!.data
                        teraluxName = teraluxData?.name
                        roomId = teraluxData?.roomId
                    }
                }
                
                // Simulated room status - replace with actual API call
                val dateFormat = SimpleDateFormat("EEEE, dd MMMM yyyy", Locale.getDefault())
                val currentDate = dateFormat.format(Date())
                
                _uiState.value = RoomStatusUiState.Success(
                    roomName = teraluxName ?: "Conference Room",
                    roomId = roomId,
                    status = RoomStatus.OCCUPIED,
                    date = currentDate,
                    bookingInfo = BookingInfo(
                        guestName = "Mr. Abdul Azis",
                        company = "PT Berlian Utama Jaya",
                        timeRange = "08:00 AM - 10:00 AM"
                    )
                )
            } catch (e: Exception) {
                _uiState.value = RoomStatusUiState.Error("Failed to load room status: ${e.message}")
            }
        }
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
