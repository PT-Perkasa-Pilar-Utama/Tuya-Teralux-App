package com.example.teraluxapp.data.network

import com.example.teraluxapp.data.model.AuthResponse
import com.example.teraluxapp.data.model.BaseResponse
import com.example.teraluxapp.data.model.DeviceResponse
import com.example.teraluxapp.data.model.DeviceListResponse
import com.example.teraluxapp.data.model.SingleDeviceResponse
import com.example.teraluxapp.data.model.SensorDataResponse
import com.example.teraluxapp.data.model.TeraluxListResponse
import com.example.teraluxapp.data.model.TeraluxResponseDTO
import com.example.teraluxapp.data.model.CreateTeraluxRequest
import com.example.teraluxapp.data.model.CreateTeraluxResponse
import com.example.teraluxapp.data.model.UpdateTeraluxRequest
import com.example.teraluxapp.data.model.CreateDeviceRequest
import com.example.teraluxapp.data.model.CreateDeviceResponse
import com.example.teraluxapp.data.model.TuyaSyncDeviceDTO
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Body
import retrofit2.Response
import retrofit2.http.PUT
import retrofit2.http.DELETE
import retrofit2.http.Query

interface ApiService {
    @GET("api/tuya/auth")
    suspend fun authenticate(): Response<BaseResponse<AuthResponse>>

    // Teralux Device Check & Registration (Public with API Key)
    @GET("api/teralux/mac/{mac}")
    suspend fun getTeraluxByMAC(
        @Path("mac") macAddress: String
    ): BaseResponse<TeraluxResponseDTO>

    @POST("api/teralux")
    suspend fun registerTeralux(
        @Body request: CreateTeraluxRequest
    ): Response<BaseResponse<CreateTeraluxResponse>>

    @GET("api/tuya/devices")
    suspend fun getDevices(
        @Header("Authorization") token: String,
        @retrofit2.http.Query("page") page: Int? = null,
        @retrofit2.http.Query("limit") limit: Int? = null,
        @retrofit2.http.Query("category") category: String? = null
    ): Response<BaseResponse<DeviceResponse>>

    @POST("api/tuya/devices/{id}/commands/switch")
    suspend fun sendDeviceCommand(
        @Header("Authorization") token: String,
        @Path("id") deviceId: String,
        @Body request: Command
    ): Response<BaseResponse<CommandResponse>>

    @GET("api/tuya/devices/{id}")
    suspend fun getDeviceById(
        @Header("Authorization") token: String,
        @Path("id") deviceId: String
    ): BaseResponse<SingleDeviceResponse>

    @PUT("api/devices/{id}/status")
    suspend fun sendIRACCommand(
        @Header("Authorization") token: String,
        @Path("id") deviceId: String,
        @Body request: IRACCommandRequest
    ): Response<BaseResponse<CommandResponse>>

    @GET("api/tuya/devices/{id}/sensor")
    suspend fun getSensorData(
        @Header("Authorization") token: String,
        @Path("id") deviceId: String
    ): Response<BaseResponse<SensorDataResponse>>

    @retrofit2.http.DELETE("api/cache/flush")
    suspend fun flushCache(
        @Header("Authorization") token: String
    ): Response<BaseResponse<Any?>>

    @GET("api/teralux/{id}")
    suspend fun getTeraluxById(
        @Header("Authorization") token: String,
        @Path("id") teraluxId: String
    ): Response<BaseResponse<TeraluxResponseDTO>>

    @PUT("api/teralux/{id}")
    suspend fun updateTeralux(
        @Header("Authorization") token: String,
        @Path("id") teraluxId: String,
        @Body request: UpdateTeraluxRequest
    ): Response<BaseResponse<Any?>>

    @POST("api/devices")
    suspend fun createDevice(
        @Header("Authorization") token: String,
        @Body request: CreateDeviceRequest
    ): Response<BaseResponse<CreateDeviceResponse>>

    @DELETE("api/devices/{id}")
    suspend fun deleteDevice(
        @Header("Authorization") token: String,
        @Path("id") deviceId: String
    ): Response<BaseResponse<Any?>>

    @GET("api/devices/teralux/{teraluxId}")
    suspend fun getDevicesByTeraluxId(
        @Header("Authorization") token: String,
        @Path("teraluxId") teraluxId: String,
        @retrofit2.http.Query("page") page: Int? = null,
        @retrofit2.http.Query("limit") limit: Int? = null
    ): Response<BaseResponse<DeviceListResponse>>


    @GET("api/tuya/devices/sync")
    suspend fun syncDevices(
        @Header("Authorization") token: String
    ): Response<BaseResponse<List<TuyaSyncDeviceDTO>>>
}

// Type-safe command value
sealed class CommandValue {
    data class BoolValue(val value: Boolean) : CommandValue()
    data class IntValue(val value: Int) : CommandValue()
    data class StringValue(val value: String) : CommandValue()
    
    fun toJsonValue(): kotlin.Any = when (this) {
        is BoolValue -> value
        is IntValue -> value
        is StringValue -> value
    }
}

@com.google.gson.annotations.JsonAdapter(CommandValueAdapter::class)
data class Command(val code: String, val value: CommandValue)
data class CommandResponse(val success: Boolean)

// IR AC Command (for Smart IR Hub)
data class IRACCommandRequest(
    val remote_id: String,
    val code: String,
    val value: Int
)

// Device State
data class SaveDeviceStateRequest(val commands: List<StateCommand>)
@com.google.gson.annotations.JsonAdapter(StateCommandValueAdapter::class)
data class StateCommand(val code: String, val value: CommandValue)
data class DeviceStateResponse(
    @com.google.gson.annotations.SerializedName("device_id") val deviceId: String,
    @com.google.gson.annotations.SerializedName("last_commands") val lastCommands: List<StateCommand>,
    @com.google.gson.annotations.SerializedName("updated_at") val updatedAt: Long
)

// Gson adapters for CommandValue serialization
class CommandValueAdapter : com.google.gson.TypeAdapter<CommandValue>() {
    override fun write(out: com.google.gson.stream.JsonWriter, value: CommandValue?) {
        if (value == null) {
            out.nullValue()
            return
        }
        when (value) {
            is CommandValue.BoolValue -> out.value(value.value)
            is CommandValue.IntValue -> out.value(value.value)
            is CommandValue.StringValue -> out.value(value.value)
        }
    }

    override fun read(`in`: com.google.gson.stream.JsonReader): CommandValue {
        return when (`in`.peek()) {
            com.google.gson.stream.JsonToken.BOOLEAN -> CommandValue.BoolValue(`in`.nextBoolean())
            com.google.gson.stream.JsonToken.NUMBER -> {
                val str = `in`.nextString()
                CommandValue.IntValue(str.toDoubleOrNull()?.toInt() ?: str.toInt())
            }
            com.google.gson.stream.JsonToken.STRING -> CommandValue.StringValue(`in`.nextString())
            else -> throw com.google.gson.JsonParseException("Unexpected token for CommandValue")
        }
    }
}

class StateCommandValueAdapter : com.google.gson.TypeAdapter<CommandValue>() {
    private val delegate = CommandValueAdapter()
    override fun write(out: com.google.gson.stream.JsonWriter, value: CommandValue?) = delegate.write(out, value)
    override fun read(`in`: com.google.gson.stream.JsonReader): CommandValue = delegate.read(`in`)
}

