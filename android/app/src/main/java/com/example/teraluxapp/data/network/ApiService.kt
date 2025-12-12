package com.example.teraluxapp.data.network

import com.example.teraluxapp.data.model.AuthResponse
import com.example.teraluxapp.data.model.DeviceResponse
import com.example.teraluxapp.data.model.SingleDeviceResponse
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Body
import retrofit2.Response

interface ApiService {
    @POST("api/tuya/auth")
    suspend fun authenticate(): AuthResponse

    @GET("api/tuya/devices")
    suspend fun getDevices(
        @Header("access_token") token: String,
        @Header("tuya_uid") uid: String
    ): Response<DeviceResponse>

    @POST("api/tuya/devices/{id}/commands")
    suspend fun sendDeviceCommand(
        @Header("access_token") token: String,
        @Path("id") deviceId: String,
        @Body request: CommandRequest
    ): Response<CommandResponse>

    @GET("api/tuya/devices/{id}")
    suspend fun getDeviceById(
        @Header("access_token") token: String,
        @Path("id") deviceId: String
    ): SingleDeviceResponse

    @POST("api/tuya/ir-ac/command")
    suspend fun sendIRACCommand(
        @Header("access_token") token: String,
        @Body request: IRACCommandRequest
    ): Response<CommandResponse>
}

data class CommandRequest(val commands: List<Command>)
data class Command(val code: String, val value: Any)
data class CommandResponse(val success: Boolean)

// IR AC Command (for Smart IR Hub)
data class IRACCommandRequest(
    val infrared_id: String,
    val remote_id: String,
    val code: String,
    val value: Int
)
