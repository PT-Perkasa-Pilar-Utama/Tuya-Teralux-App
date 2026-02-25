package com.example.teraluxapp.utils

import android.content.Context
import android.content.SharedPreferences

object PreferencesManager {
    private const val PREF_NAME = "teralux_prefs"
    private const val KEY_TERALUX_ID = "teralux_id"
    private const val KEY_UID = "uid"
    
    private fun getPreferences(context: Context): SharedPreferences {
        return context.getSharedPreferences(PREF_NAME, Context.MODE_PRIVATE)
    }
    
    fun saveTeraluxId(context: Context, teraluxId: String) {
        getPreferences(context).edit().putString(KEY_TERALUX_ID, teraluxId).apply()
    }
    
    fun getTeraluxId(context: Context): String? {
        return getPreferences(context).getString(KEY_TERALUX_ID, null)
    }
    
    fun saveUid(context: Context, uid: String) {
        getPreferences(context).edit().putString(KEY_UID, uid).apply()
    }
    
    fun getUid(context: Context): String? {
        return getPreferences(context).getString(KEY_UID, null)
    }
    
    fun clearUid(context: Context) {
        getPreferences(context).edit().remove(KEY_UID).apply()
    }
}
