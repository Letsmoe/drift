package dev.letsmoe.drift.playback

import dev.letsmoe.drift.core.model.RepeatMode
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class CrossfadePlannerTest {
    private val songMs = 200_000L

    private fun situation(
        enabled: Boolean = true,
        fadeSeconds: Int = 6,
        positionMs: Long = 195_000,
        durationMs: Long? = songMs,
        isPlaying: Boolean = true,
        hasNextEntry: Boolean = true,
    ) = CrossfadeSituation(enabled, fadeSeconds, positionMs, durationMs, isPlaying, hasNextEntry)

    @Test
    fun startsWithinTheFadeLengthOfTheEnd() {
        assertTrue(CrossfadePlanner.shouldStart(situation(positionMs = 194_000)))
        assertTrue(CrossfadePlanner.shouldStart(situation(positionMs = 199_000)))
    }

    @Test
    fun waitsUntilTheFadeLengthFromTheEnd() {
        assertFalse(CrossfadePlanner.shouldStart(situation(positionMs = 193_900)))
        assertFalse(CrossfadePlanner.shouldStart(situation(positionMs = 0)))
    }

    @Test
    fun doesNothingWhenDisabled() {
        assertFalse(CrossfadePlanner.shouldStart(situation(enabled = false)))
    }

    @Test
    fun doesNothingWhilePausedOrWithoutNextSongOrDuration() {
        assertFalse(CrossfadePlanner.shouldStart(situation(isPlaying = false)))
        assertFalse(CrossfadePlanner.shouldStart(situation(hasNextEntry = false)))
        assertFalse(CrossfadePlanner.shouldStart(situation(durationMs = null)))
    }

    @Test
    fun skipsSongsShorterThanThreeFadeLengths() {
        val shortSituation = situation(fadeSeconds = 6, durationMs = 17_999, positionMs = 15_000)
        assertFalse(CrossfadePlanner.shouldStart(shortSituation))
        val longEnough = situation(fadeSeconds = 6, durationMs = 18_000, positionMs = 15_000)
        assertTrue(CrossfadePlanner.shouldStart(longEnough))
    }

    @Test
    fun nextEntryFollowsQueuePositionAndRepeat() {
        assertTrue(CrossfadePlanner.hasNextEntry(3, 0, RepeatMode.Off))
        assertFalse(CrossfadePlanner.hasNextEntry(3, 2, RepeatMode.Off))
        assertTrue(CrossfadePlanner.hasNextEntry(3, 2, RepeatMode.All))
        assertFalse(CrossfadePlanner.hasNextEntry(3, 0, RepeatMode.One))
    }

    @Test
    fun aSingleEntryQueueNeverHasANextEntry() {
        assertFalse(CrossfadePlanner.hasNextEntry(1, 0, RepeatMode.All))
        assertFalse(CrossfadePlanner.hasNextEntry(0, -1, RepeatMode.Off))
    }

    @Test
    fun progressIsClamped() {
        assertEquals(0f, CrossfadePlanner.progress(-10, 6000), 0f)
        assertEquals(0.5f, CrossfadePlanner.progress(3000, 6000), 0.0001f)
        assertEquals(1f, CrossfadePlanner.progress(9000, 6000), 0f)
        assertEquals(1f, CrossfadePlanner.progress(0, 0), 0f)
    }

    @Test
    fun volumesCrossOverAtEqualPower() {
        assertEquals(0f, CrossfadePlanner.incomingVolume(0f), 0.0001f)
        assertEquals(1f, CrossfadePlanner.incomingVolume(1f), 0.0001f)
        assertEquals(1f, CrossfadePlanner.outgoingVolume(0f), 0.0001f)
        assertEquals(0f, CrossfadePlanner.outgoingVolume(1f), 0.0001f)
        for (step in 0..10) {
            val progress = step / 10f
            val incoming = CrossfadePlanner.incomingVolume(progress)
            val outgoing = CrossfadePlanner.outgoingVolume(progress)
            assertEquals(1f, incoming * incoming + outgoing * outgoing, 0.0001f)
        }
    }
}
