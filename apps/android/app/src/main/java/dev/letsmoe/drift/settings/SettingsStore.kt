package dev.letsmoe.drift.settings

import kotlinx.coroutines.flow.StateFlow

/** Allowed crossfade lengths; the same limits and default as the desktop app. */
object CrossfadeSeconds {
    const val MIN = 1
    const val MAX = 12
    const val DEFAULT = 6
}

/**
 * User preferences, kept on the device. Each setting is a [StateFlow] holding its current value
 * plus a setter that updates the flow and persists the value. Features add their own settings here.
 */
interface SettingsStore {
    val crossfadeEnabled: StateFlow<Boolean>
    val crossfadeSeconds: StateFlow<Int>
    val downloadOnlyOnWifi: StateFlow<Boolean>

    fun setCrossfadeEnabled(enabled: Boolean)

    /** Values outside [CrossfadeSeconds.MIN]..[CrossfadeSeconds.MAX] are clamped. */
    fun setCrossfadeSeconds(seconds: Int)

    fun setDownloadOnlyOnWifi(enabled: Boolean)
}
