package dev.letsmoe.drift.playback

import dev.letsmoe.drift.core.model.RepeatMode
import kotlin.math.PI
import kotlin.math.cos
import kotlin.math.sin

/** Songs shorter than this many crossfade lengths end without one, so they don't fade most of their length. */
const val MINIMUM_LENGTH_IN_FADES = 3

/** What the planner needs to know about playback at one moment. */
data class CrossfadeSituation(
    val enabled: Boolean,
    val fadeSeconds: Int,
    val positionMs: Long,
    /** Null while the stream has not reported its length. */
    val durationMs: Long?,
    val isPlaying: Boolean,
    val hasNextEntry: Boolean,
)

/** The crossfade timing decisions, with no player types so they can be unit tested. */
object CrossfadePlanner {
    /** True when the ending song should hand over to the next one now. */
    fun shouldStart(situation: CrossfadeSituation): Boolean {
        val durationMs = situation.durationMs
        if (!situation.enabled || !situation.isPlaying || !situation.hasNextEntry || durationMs == null) {
            return false
        }
        val fadeMs = situation.fadeSeconds * 1000L
        if (durationMs < fadeMs * MINIMUM_LENGTH_IN_FADES) {
            return false
        }
        val remainingMs = durationMs - situation.positionMs
        return remainingMs <= fadeMs
    }

    /**
     * Whether advancing the queue lands on a different entry. Repeat one replays the song, and a
     * single-entry queue on repeat all would load the same entry again, so neither crossfades.
     */
    fun hasNextEntry(entryCount: Int, index: Int, repeat: RepeatMode): Boolean {
        if (repeat == RepeatMode.One || entryCount < 2) {
            return false
        }
        return index < entryCount - 1 || repeat == RepeatMode.All
    }

    /** How far the fade has come, from 0 to 1. */
    fun progress(elapsedMs: Long, fadeMs: Long): Float {
        if (fadeMs <= 0) {
            return 1f
        }
        return (elapsedMs.toFloat() / fadeMs).coerceIn(0f, 1f)
    }

    // Equal-power curves keep the loudness even while both songs play.
    fun incomingVolume(progress: Float): Float = sin(progress.coerceIn(0f, 1f) * PI.toFloat() / 2)

    fun outgoingVolume(progress: Float): Float = cos(progress.coerceIn(0f, 1f) * PI.toFloat() / 2)
}
