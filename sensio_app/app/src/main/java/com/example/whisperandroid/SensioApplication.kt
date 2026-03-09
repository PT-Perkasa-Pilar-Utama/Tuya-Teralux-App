package com.example.whisperandroid

import android.app.Application
import com.example.whisperandroid.data.di.NetworkModule

class SensioApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        NetworkModule.init(this)
    }
}
