package com.example.whisperandroid.data.local

import android.content.Context
import android.util.Log
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.failedUploadDataStore: DataStore<Preferences> by preferencesDataStore(
    name = "failed_uploads"
)

class FailedUploadStore(private val context: Context) {
    private val gson = Gson()
    private val tag = "FailedUploadStore"
    private val maxRetries = 3

    private object Keys {
        val FAILED_UPLOADS = stringPreferencesKey("failed_uploads")
    }

    fun getFailedUploadsFlow(): Flow<List<FailedUpload>> {
        return context.failedUploadDataStore.data.map { prefs ->
            val json = prefs[Keys.FAILED_UPLOADS] ?: return@map emptyList()
            val type = object : TypeToken<List<FailedUpload>>() {}.type
            try {
                gson.fromJson(json, type) ?: emptyList()
            } catch (e: Exception) {
                Log.e(tag, "Error parsing failed uploads: ${e.message}")
                emptyList()
            }
        }
    }

    suspend fun addFailedUpload(filePath: String, bookingId: String, error: String) {
        context.failedUploadDataStore.edit { prefs ->
            val current = getFailedUploadList(prefs)
            val existingIndex = current.indexOfFirst { it.localFilePath == filePath }
            val newRetryCount = if (existingIndex >= 0) {
                current[existingIndex].retryCount + 1
            } else {
                1
            }

            if (newRetryCount > maxRetries) {
                val updated = current.filter { it.localFilePath != filePath }
                prefs[Keys.FAILED_UPLOADS] = gson.toJson(updated)
                Log.w(tag, "Max retries ($maxRetries) exceeded for $filePath, removed from retry queue")
                return@edit
            }

            val newEntry = if (existingIndex >= 0) {
                current[existingIndex].withError(error).copy(retryCount = newRetryCount)
            } else {
                FailedUpload(
                    localFilePath = filePath,
                    bookingId = bookingId,
                    lastError = error,
                    retryCount = newRetryCount
                )
            }
            val updated = if (existingIndex >= 0) {
                current.toMutableList().apply { set(existingIndex, newEntry) }
            } else {
                current + newEntry
            }
            prefs[Keys.FAILED_UPLOADS] = gson.toJson(updated)
            Log.i(tag, "Added failed upload: $filePath, retry count: $newRetryCount")
        }
    }

    suspend fun removeFailedUpload(filePath: String) {
        context.failedUploadDataStore.edit { prefs ->
            val current = getFailedUploadList(prefs)
            val updated = current.filter { it.localFilePath != filePath }
            prefs[Keys.FAILED_UPLOADS] = gson.toJson(updated)
            Log.i(tag, "Removed failed upload: $filePath")
        }
    }

    suspend fun getFailedUpload(filePath: String): FailedUpload? {
        var result: FailedUpload? = null
        context.failedUploadDataStore.edit { prefs ->
            val current = getFailedUploadList(prefs)
            result = current.find { it.localFilePath == filePath }
        }
        return result
    }

    suspend fun incrementRetryCount(filePath: String) {
        context.failedUploadDataStore.edit { prefs ->
            val current = getFailedUploadList(prefs)
            val updated = current.map {
                if (it.localFilePath == filePath) it.withIncrementedRetry() else it
            }
            prefs[Keys.FAILED_UPLOADS] = gson.toJson(updated)
            Log.d(tag, "Incremented retry count for: $filePath")
        }
    }

    private fun getFailedUploadList(prefs: Preferences): List<FailedUpload> {
        val json = prefs[Keys.FAILED_UPLOADS] ?: return emptyList()
        val type = object : TypeToken<List<FailedUpload>>() {}.type
        return try {
            gson.fromJson(json, type) ?: emptyList()
        } catch (e: Exception) {
            Log.e(tag, "Error parsing failed uploads: ${e.message}")
            emptyList()
        }
    }
}
