package com.example.teraluxapp.utils

import com.example.teraluxapp.data.model.BaseResponse
import com.google.gson.Gson
import retrofit2.Response

/**
 * Extension function to extract error message from Retrofit Response
 * Parses the error body to get the backend's error message
 */
fun <T> Response<T>.getErrorMessage(): String {
    return try {
        val errorBody = this.errorBody()?.string()
        if (errorBody != null) {
            val gson = Gson()
            val errorResponse = gson.fromJson(errorBody, BaseResponse::class.java)
            
            var message = errorResponse.message
            
            // Check for validation details
            if (errorResponse.details != null) {
                if (errorResponse.details is Map<*, *>) {
                    val detailsMap = errorResponse.details as Map<*, *>
                    val detailsBuilder = StringBuilder()
                    detailsBuilder.append("\n")
                    
                    detailsMap.forEach { (key, value) ->
                        if (key is String) {
                             val errorList = if (value is List<*>) {
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
        } else {
            "An error occurred. Please try again"
        }
    } catch (e: Exception) {
        "An error occurred. Please try again"
    }
}
