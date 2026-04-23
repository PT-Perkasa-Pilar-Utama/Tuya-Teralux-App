package com.example.whisperandroid.utils

import android.util.Log
import com.example.whisperandroid.BuildConfig

/**
 * Centralized logging utility for Sensio app.
 * Automatically prefixes tags with "Sensio" and suppresses debug logs in release builds.
 */
object AppLog {
    private const val PREFIX = "Sensio"

    fun d(tag: String, msg: String) {
        if (BuildConfig.DEBUG) {
            Log.d(formatTag(tag), msg)
        }
    }

    fun i(tag: String, msg: String) {
        if (BuildConfig.DEBUG) {
            Log.i(formatTag(tag), msg)
        }
    }

    fun w(tag: String, msg: String) {
        Log.w(formatTag(tag), msg)
    }

    fun e(tag: String, msg: String, throwable: Throwable? = null) {
        if (throwable != null) {
            Log.e(formatTag(tag), msg, throwable)
        } else {
            Log.e(formatTag(tag), msg)
        }
    }

    private fun formatTag(tag: String): String {
        return if (tag.startsWith(PREFIX)) tag else "$PREFIX$tag"
    }
}
