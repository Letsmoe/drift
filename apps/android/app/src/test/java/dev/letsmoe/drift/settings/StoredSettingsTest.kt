package dev.letsmoe.drift.settings

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

private class MapStorage : SettingsStorage {
    val values = mutableMapOf<String, Any>()

    override fun readBoolean(key: String): Boolean? = values[key] as Boolean?
    override fun readInt(key: String): Int? = values[key] as Int?
    override fun writeBoolean(key: String, value: Boolean) {
        values[key] = value
    }

    override fun writeInt(key: String, value: Int) {
        values[key] = value
    }
}

class StoredSettingsTest {
    @Test
    fun startsWithDefaults() {
        val settings = StoredSettings(MapStorage())
        assertFalse(settings.crossfadeEnabled.value)
        assertEquals(CrossfadeSeconds.DEFAULT, settings.crossfadeSeconds.value)
        assertTrue(settings.downloadOnlyOnWifi.value)
    }

    @Test
    fun changesSurviveARestart() {
        val storage = MapStorage()
        val settings = StoredSettings(storage)
        settings.setCrossfadeEnabled(true)
        settings.setCrossfadeSeconds(9)
        settings.setDownloadOnlyOnWifi(false)

        val reloaded = StoredSettings(storage)
        assertTrue(reloaded.crossfadeEnabled.value)
        assertEquals(9, reloaded.crossfadeSeconds.value)
        assertFalse(reloaded.downloadOnlyOnWifi.value)
    }

    @Test
    fun clampsCrossfadeSeconds() {
        val settings = StoredSettings(MapStorage())
        settings.setCrossfadeSeconds(99)
        assertEquals(CrossfadeSeconds.MAX, settings.crossfadeSeconds.value)
        settings.setCrossfadeSeconds(0)
        assertEquals(CrossfadeSeconds.MIN, settings.crossfadeSeconds.value)
    }

    @Test
    fun repairsOutOfRangeStoredSeconds() {
        val storage = MapStorage()
        storage.writeInt("crossfade_seconds", 0)
        assertEquals(CrossfadeSeconds.DEFAULT, StoredSettings(storage).crossfadeSeconds.value)
        storage.writeInt("crossfade_seconds", 40)
        assertEquals(CrossfadeSeconds.MAX, StoredSettings(storage).crossfadeSeconds.value)
    }
}
