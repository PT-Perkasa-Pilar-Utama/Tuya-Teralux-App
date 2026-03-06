package com.example.whisperandroid.presentation.meeting

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import com.example.whisperandroid.domain.usecase.ProcessMeetingUseCase

class MeetingViewModelFactory(
    private val processMeetingUseCase: ProcessMeetingUseCase
) : ViewModelProvider.Factory {
    @Suppress("UNCHECKED_CAST")
    override fun <T : ViewModel> create(modelClass: Class<T>): T {
        if (modelClass.isAssignableFrom(MeetingViewModel::class.java)) {
            return MeetingViewModel(processMeetingUseCase) as T
        }
        throw IllegalArgumentException("Unknown ViewModel class")
    }
}
