package dev.letsmoe.drift.settings

import android.content.Context
import android.content.SharedPreferences

private const val PREFERENCES_NAME = "drift_settings"

/** [SettingsStorage] on Android's SharedPreferences. */
class SharedPreferencesStorage(context: Context) : SettingsStorage {
    private val preferences: SharedPreferences =
        context.getSharedPreferences(PREFERENCES_NAME, Context.MODE_PRIVATE)

    override fun readBoolean(key: String): Boolean? {
        if (!preferences.contains(key)) {
            return null
        }
        return preferences.getBoolean(key, false)
    }

    override fun readInt(key: String): Int? {
        if (!preferences.contains(key)) {
            return null
        }
        return preferences.getInt(key, 0)
    }

    override fun writeBoolean(key: String, value: Boolean) {
        preferences.edit().putBoolean(key, value).apply()
    }

    override fun writeInt(key: String, value: Int) {
        preferences.edit().putInt(key, value).apply()
    }
}
