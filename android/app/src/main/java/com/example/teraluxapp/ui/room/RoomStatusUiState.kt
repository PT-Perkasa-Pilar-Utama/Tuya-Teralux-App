package com.example.teraluxapp.ui.room

sealed class RoomStatusUiState {
    data object Loading : RoomStatusUiState()
    data class Success(
        val roomName: String,
        val status: RoomStatus,
        val bookingInfo: BookingInfo? = null,
        val date: String
    ) : RoomStatusUiState()
    data class Error(val message: String) : RoomStatusUiState()
}

enum class RoomStatus {
    VACANT,
    BOOKED,
    OCCUPIED
}

data class BookingInfo(
    val guestName: String,
    val company: String,
    val timeRange: String
)
