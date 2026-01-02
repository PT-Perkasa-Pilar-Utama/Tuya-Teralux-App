package com.example.teraluxapp.data.network

import com.example.teraluxapp.config.ApiConfig
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

object RetrofitClient {
    private val logging = HttpLoggingInterceptor().apply {
        level = HttpLoggingInterceptor.Level.BODY
    }

    private val client = OkHttpClient.Builder()
        .addInterceptor(FailoverInterceptor())
        .addInterceptor(AuthInterceptor())
        .addInterceptor { chain ->
            val original = chain.request()
            val builder = original.newBuilder()
            
            val path = original.url.encodedPath
            if (path.contains("api/tuya/auth") || path.contains("api/teralux")) {
                builder.header("X-API-KEY", com.example.teraluxapp.BuildConfig.API_KEY)
            }
            
            chain.proceed(builder.build())
        }
        .addInterceptor(logging)
        .build()

    val instance: ApiService by lazy {
        Retrofit.Builder()
            .baseUrl(ApiConfig.BASE_URLS[0])
            .addConverterFactory(GsonConverterFactory.create())
            .client(client)
            .build()
            .create(ApiService::class.java)
    }
}
