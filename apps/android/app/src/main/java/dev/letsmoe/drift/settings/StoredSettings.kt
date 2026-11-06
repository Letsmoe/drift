package dev.letsmoe.drift.settings

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

/** Platform key-value persistence behind [StoredSettings]; a missing key reads as null. */
interface SettingsStorage {
    fun readBoolean(key: String): Boolean?
    fun readInt(key: String): Int?
    fun writeBoolean(key: String, value: Boolean)
    fun writeInt(key: String, value: Int)
}

private const val CROSSFADE_ENABLED_KEY = "crossfade_enabled"
private const val CROSSFADE_SECONDS_KEY = "crossfade_seconds"
private const val DOWNLOAD_ONLY_ON_WIFI_KEY = "download_only_on_wifi"

/** [SettingsStore] that loads its values once from a [SettingsStorage] and writes every change back. */
class StoredSettings(private val storage: SettingsStorage) : SettingsStore {
    private val crossfadeEnabledState = MutableStateFlow(storage.readBoolean(CROSSFADE_ENABLED_KEY) == true)
    private val crossfadeSecondsState = MutableStateFlow(loadCrossfadeSeconds())
    private val downloadOnlyOnWifiState = MutableStateFlow(storage.readBoolean(DOWNLOAD_ONLY_ON_WIFI_KEY) != false)

    override val crossfadeEnabled: StateFlow<Boolean> = crossfadeEnabledState
    override val crossfadeSeconds: StateFlow<Int> = crossfadeSecondsState
    override val downloadOnlyOnWifi: StateFlow<Boolean> = downloadOnlyOnWifiState

    override fun setCrossfadeEnabled(enabled: Boolean) {
        crossfadeEnabledState.value = enabled
        storage.writeBoolean(CROSSFADE_ENABLED_KEY, enabled)
    }

    override fun setCrossfadeSeconds(seconds: Int) {
        val clamped = seconds.coerceIn(CrossfadeSeconds.MIN, CrossfadeSeconds.MAX)
        crossfadeSecondsState.value = clamped
        storage.writeInt(CROSSFADE_SECONDS_KEY, clamped)
    }

    override fun setDownloadOnlyOnWifi(enabled: Boolean) {
        downloadOnlyOnWifiState.value = enabled
        storage.writeBoolean(DOWNLOAD_ONLY_ON_WIFI_KEY, enabled)
    }

    private fun loadCrossfadeSeconds(): Int {
        val stored = storage.readInt(CROSSFADE_SECONDS_KEY)
        if (stored == null || stored < CrossfadeSeconds.MIN) {
            return CrossfadeSeconds.DEFAULT
        }
        return stored.coerceAtMost(CrossfadeSeconds.MAX)
    }
}
