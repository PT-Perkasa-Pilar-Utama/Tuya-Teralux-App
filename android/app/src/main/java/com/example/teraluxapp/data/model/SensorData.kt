package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class SensorDataResponse(
    @SerializedName("temperature") val temperature: Double,
    @SerializedName("humidity") val humidity: Int,
    @SerializedName("battery_percentage") val batteryPercentage: Int,
    @SerializedName("status_text") val statusText: String,
    @SerializedName("temp_unit") val tempUnit: String
)
