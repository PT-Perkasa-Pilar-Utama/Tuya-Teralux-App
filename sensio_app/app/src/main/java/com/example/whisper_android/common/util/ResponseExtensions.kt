package com.example.whisper_android.common.util

import com.example.whisper_android.data.remote.dto.StandardResponseDto
import com.google.gson.Gson
import retrofit2.HttpException
import retrofit2.Response

/**
 * Extension to extract error message from Retrofit Response
 */
fun <T> Response<T>.getErrorMessage(): String =
    try {
        val errorBody = this.errorBody()?.string()
        if (errorBody != null) {
            parseErrorBody(errorBody)
        } else {
            "An error occurred. Please try again"
        }
    } catch (e: Exception) {
        "An error occurred. Please try again"
    }

/**
 * Extension to extract error message from HttpException
 */
fun HttpException.getErrorMessage(): String =
    try {
        val errorBody = response()?.errorBody()?.string()
        if (errorBody != null) {
            parseErrorBody(errorBody)
        } else {
            message() ?: "An error occurred"
        }
    } catch (e: Exception) {
        message() ?: "An error occurred"
    }

private fun parseErrorBody(errorBody: String): String =
    try {
        val gson = Gson()
        // We use Any for T because we don't care about data type when parsing error
        val errorResponse = gson.fromJson(errorBody, StandardResponseDto::class.java)

        var message = errorResponse.message

        // Check for validation details
        if (errorResponse.details != null) {
            if (errorResponse.details is Map<*, *>) {
                val detailsMap = errorResponse.details as Map<*, *>
                val detailsBuilder = StringBuilder()
                detailsBuilder.append("\n")

                detailsMap.forEach { (key, value) ->
                    if (key is String) {
                        val errorList =
                            if (value is List<*>) {
                                value.joinToString(", ")
                            } else {
                                value.toString()
                            }
                        detailsBuilder.append("- $key: $errorList\n")
                    }
                }
                if (detailsBuilder.length > 1) {
                    message += detailsBuilder.toString()
                }
            } else if (errorResponse.details is String) {
                message += "\n${errorResponse.details}"
            }
        }
        message
    } catch (e: Exception) {
        "An error occurred. Please try again"
    }
