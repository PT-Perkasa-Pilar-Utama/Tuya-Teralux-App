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
            errorResponse.message
        } else {
            "An error occurred. Please try again"
        }
    } catch (e: Exception) {
        "An error occurred. Please try again"
    }
}
