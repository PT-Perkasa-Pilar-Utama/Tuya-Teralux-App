package com.example.whisperandroid.presentation.meeting.components

import org.junit.Assert.assertEquals
import org.junit.Test

class MeetingErrorMapperTest {

    @Test
    fun testUploadInterrupted() {
        val uiModel = mapMeetingErrorToUiModel("Upload failed: network transport error")
        assertEquals("Upload interrupted", uiModel.title)
        assertEquals(true, uiModel.isRetrySuggested)
    }

    @Test
    fun testTimeoutOrion() {
        val uiModel = mapMeetingErrorToUiModel("Error 504: Gateway Time-out from Orion")
        assertEquals("The summary took too long to process", uiModel.title)

        val uiModel2 = mapMeetingErrorToUiModel("deadline exceeded during transcription")
        assertEquals("The summary took too long to process", uiModel2.title)
    }

    @Test
    fun testPipelineInitiationFailed() {
        val uiModel = mapMeetingErrorToUiModel("Pipeline initiation failed")
        assertEquals("Couldn't start the summary", uiModel.title)
    }

    @Test
    fun testSessionExpired() {
        val uiModel = mapMeetingErrorToUiModel("Unauthorized token")
        assertEquals("Session expired", uiModel.title)
        assertEquals(false, uiModel.isRetrySuggested)
    }

    @Test
    fun testNetworkIssue() {
        val uiModel = mapMeetingErrorToUiModel("network unreachable")
        assertEquals("Connection issue", uiModel.title)
    }

    @Test
    fun testGenericError() {
        val uiModel = mapMeetingErrorToUiModel("Pipeline execution failed mysteriously")
        assertEquals("Summary unavailable", uiModel.title)
    }

    @Test
    fun testNullMessage() {
        val uiModel = mapMeetingErrorToUiModel(null)
        assertEquals("Summary unavailable", uiModel.title)
    }
}
