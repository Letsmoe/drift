package dev.letsmoe.drift.downloads

import dev.letsmoe.drift.core.DriftJson
import dev.letsmoe.drift.core.model.CollectionKind
import dev.letsmoe.drift.core.model.DownloadCollection
import dev.letsmoe.drift.core.model.DownloadState
import dev.letsmoe.drift.core.model.DownloadUpdate
import dev.letsmoe.drift.ui.formatBytes
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class DownloadsTest {
    @Test
    fun wifiOnlyWaitsForAnUnmeteredNetwork() {
        assertTrue(downloadsAllowed(isUnmetered = true, downloadOnlyOnWifi = true))
        assertFalse(downloadsAllowed(isUnmetered = false, downloadOnlyOnWifi = true))
    }

    @Test
    fun anyNetworkWorksWhenWifiOnlyIsOff() {
        assertTrue(downloadsAllowed(isUnmetered = false, downloadOnlyOnWifi = false))
        assertTrue(downloadsAllowed(isUnmetered = true, downloadOnlyOnWifi = false))
    }

    @Test
    fun decodesTheCoreUpdate() {
        val json = """
            {"status":{"videoId":"abc","state":"downloading","bytesDone":50,"bytesTotal":200,"error":""},
             "summary":{"queued":1,"downloading":1,"failed":0,"done":3,"paused":false,"blocked":true}}
        """.trimIndent()

        val update = DriftJson.decodeFromString<DownloadUpdate>(json)

        assertEquals(DownloadState.Downloading, update.status.state)
        assertEquals(0.25f, update.status.fraction, 0.001f)
        assertTrue(update.summary.blocked)
        assertEquals(3, update.summary.done)
    }

    @Test
    fun aVanishedDownloadHasNoState() {
        val json = """{"status":{"videoId":"abc","state":"","bytesDone":0,"bytesTotal":0,"error":""},"summary":{}}"""

        val update = DriftJson.decodeFromString<DownloadUpdate>(json)

        assertEquals(DownloadState.None, update.status.state)
    }

    @Test
    fun formatsStorageSizes() {
        assertEquals("512 B", formatBytes(512))
        assertEquals("340 KB", formatBytes(340L * 1024))
        assertTrue(formatBytes(5L * 1024 * 1024).endsWith(" MB"))
        assertTrue(formatBytes(3L * 1024 * 1024 * 1024).endsWith(" GB"))
    }

    @Test
    fun aCollectionIsCompleteWhenEveryTrackIsDone() {
        val partial = DownloadCollection(kind = CollectionKind.Album, id = "a", trackCount = 10, doneCount = 4)
        val complete = partial.copy(doneCount = 10)
        val empty = partial.copy(trackCount = 0, doneCount = 0)

        assertFalse(partial.isComplete)
        assertEquals(0.4f, partial.fraction, 0.001f)
        assertTrue(complete.isComplete)
        assertFalse(empty.isComplete)
    }
}
