package com.example.whisperandroid.data.remote.api

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.HTTP
import retrofit2.http.Header
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.PUT

/**
 * Contract tests for Retrofit API interfaces.
 * * These tests ensure that:
 * 1. API endpoint paths match the expected contract
 * 2. HTTP methods are correctly annotated
 * 3. Required headers are present
 * 4. Response envelope structure is consistent
 * * Run with: ./gradlew test --tests "*ApiContractTest*"
 */
class ApiContractTest {

    // ========================================================================
    // Terminal API Contract Tests
    // ========================================================================

    @Test
    fun terminalApi_registerTerminal_hasCorrectPath() {
        val method = TerminalApi::class.java.getMethod("registerTerminal", String::class.java, com.example.whisperandroid.data.remote.dto.TerminalRequestDto::class.java)
        val post = method.getAnnotation(POST::class.java)
        assertEquals("/api/terminal", post?.value)
    }

    @Test
    fun terminalApi_registerTerminal_requiresApiKeyHeader() {
        val method = TerminalApi::class.java.getMethod("registerTerminal", String::class.java, com.example.whisperandroid.data.remote.dto.TerminalRequestDto::class.java)
        val hasHeader = method.parameterAnnotations.any { paramAnnotation ->
            paramAnnotation.filterIsInstance<Header>().any { it.value == "X-API-KEY" }
        }
        assertTrue(hasHeader)
    }

    @Test
    fun terminalApi_getTerminalByMac_hasCorrectPath() {
        val method = TerminalApi::class.java.getMethod("getTerminalByMac", String::class.java, String::class.java)
        val get = method.getAnnotation(retrofit2.http.GET::class.java)
        assertEquals("/api/terminal/mac/{mac}", get?.value)
    }

    @Test
    fun terminalApi_updateTerminal_hasCorrectPath() {
        val method = TerminalApi::class.java.getMethod("updateTerminal", String::class.java, String::class.java, com.example.whisperandroid.data.remote.dto.UpdateTerminalRequestDto::class.java)
        val put = method.getAnnotation(PUT::class.java)
        assertEquals("/api/terminal/{id}", put?.value)
    }

    @Test
    fun terminalApi_updateTerminal_requiresAuthHeader() {
        val method = TerminalApi::class.java.getMethod("updateTerminal", String::class.java, String::class.java, com.example.whisperandroid.data.remote.dto.UpdateTerminalRequestDto::class.java)
        val hasHeader = method.parameterAnnotations.any { paramAnnotation ->
            paramAnnotation.filterIsInstance<Header>().any { it.value == "Authorization" }
        }
        assertTrue(hasHeader)
    }

    @Test
    fun terminalApi_getMqttCredentials_hasCorrectPath() {
        val method = TerminalApi::class.java.getMethod("getMqttCredentials", String::class.java, String::class.java)
        val get = method.getAnnotation(GET::class.java)
        assertEquals("/api/mqtt/users/{username}", get?.value)
    }

    // ========================================================================
    // Pipeline API Contract Tests
    // ========================================================================

    @Test
    fun pipelineApi_executeJob_hasCorrectPath() {
        val method = PipelineApi::class.java.getMethod(
            "executePipeline", okhttp3.MultipartBody.Part::class.java,
            String::class.java, String::class.java, Boolean::class.java,
            Boolean::class.java, Boolean::class.java, String::class.java,
            String::class.java, String::class.java, String::class.java,
            String::class.java, String::class.java, String::class.java,
            String::class.java
        )
        val post = method.getAnnotation(POST::class.java)
        assertEquals("/api/models/pipeline/job", post?.value)
    }

    @Test
    fun pipelineApi_executeJob_isMultipart() {
        val method = PipelineApi::class.java.getMethod(
            "executePipeline", okhttp3.MultipartBody.Part::class.java,
            String::class.java, String::class.java, Boolean::class.java,
            Boolean::class.java, Boolean::class.java, String::class.java,
            String::class.java, String::class.java, String::class.java,
            String::class.java, String::class.java, String::class.java,
            String::class.java
        )
        val multipart = method.getAnnotation(Multipart::class.java)
        assertTrue(multipart != null)
    }

    @Test
    fun pipelineApi_getPipelineStatus_hasCorrectPath() {
        val method = PipelineApi::class.java.getMethod("getPipelineStatus", String::class.java, String::class.java)
        val get = method.getAnnotation(GET::class.java)
        assertEquals("/api/models/pipeline/status/{task_id}", get?.value)
    }

    @Test
    fun pipelineApi_cancelPipelineTask_hasCorrectPath() {
        val method = PipelineApi::class.java.getMethod("cancelPipelineTask", String::class.java, String::class.java)
        val delete = method.getAnnotation(DELETE::class.java)
        assertEquals("/api/models/pipeline/status/{task_id}", delete?.value)
    }

    // ========================================================================
    // Response Envelope Structure Tests
    // ========================================================================

    @Test
    fun responseEnvelope_hasRequiredFields() {
        // This test verifies that our response DTOs have the expected structure
        // All API responses should follow the StandardResponse pattern:
        // { "status": boolean, "message": string, "data": T? }

        val statusField = com.example.whisperandroid.data.remote.dto.TerminalResponseDto::class.java.getDeclaredField("status")
        val messageField = com.example.whisperandroid.data.remote.dto.TerminalResponseDto::class.java.getDeclaredField("message")
        val dataField = com.example.whisperandroid.data.remote.dto.TerminalResponseDto::class.java.getDeclaredField("data")

        assertEquals(Boolean::class.java, statusField.type)
        assertEquals(String::class.java, messageField.type)
        // data field can be nullable
    }

    // ========================================================================
    // Path Consistency Tests
    // ========================================================================

    @Test
    fun allApiPaths_useConsistentPrefixes() {
        val apiClasses = listOf(
            TerminalApi::class.java,
            PipelineApi::class.java,
            com.example.whisperandroid.data.remote.api.WhisperApi::class.java,
            com.example.whisperandroid.data.remote.api.RAGApi::class.java
        )

        val validPrefixes = setOf(
            "/api/terminal",
            "/api/models",
            "/api/mqtt",
            "/api/tuya",
            "/api/scene",
            "/api/device",
            "/api/notification",
            "/api/big",
            "/api/cache",
            "/api/health"
        )

        for (apiClass in apiClasses) {
            for (method in apiClass.declaredMethods) {
                val pathAnnotation = method.annotations
                    .firstOrNull { it is POST || it is GET || it is PUT || it is DELETE || it is HTTP }

                val path = when (pathAnnotation) {
                    is POST -> pathAnnotation.value
                    is GET -> pathAnnotation.value
                    is PUT -> pathAnnotation.value
                    is DELETE -> pathAnnotation.value
                    is HTTP -> pathAnnotation.path
                    else -> null
                }

                if (path != null) {
                    val hasValidPrefix = validPrefixes.any { prefix ->
                        path.startsWith(prefix)
                    }
                    assertTrue(
                        "Method ${apiClass.simpleName}.${method.name} has invalid path prefix: $path",
                        hasValidPrefix
                    )
                }
            }
        }
    }
}
