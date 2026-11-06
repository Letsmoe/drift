package dev.letsmoe.drift.playback

import android.os.SystemClock
import androidx.media3.common.C
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.settings.SettingsStore
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withTimeoutOrNull
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

private const val CHECK_INTERVAL_MS = 200L
private const val FADE_STEP_MS = 50L

/** Longest wait for the second player to take over the ending song before giving up on the crossfade. */
private const val HAND_OVER_TIMEOUT_MS = 3_000L

/**
 * Crossfades into the next song. The main player always plays the current song; near its end a
 * second player takes over the ending song at the same position and fades it out, while the main
 * player moves on to the next song (the queue advances through [advanceQueue]) and fades it in.
 * The second player is not part of the media session. Mirrors crossfade.svelte.ts.
 */
class Crossfader(
    private val mainPlayer: ExoPlayer,
    private val createOutgoingPlayer: () -> ExoPlayer,
    private val settings: SettingsStore,
    private val core: DriftCore,
    private val scope: CoroutineScope,
    /** Completes the ending play and advances the queue; returns the new entry id, or null when it did not move. */
    private val advanceQueue: suspend () -> Long?,
) {
    private var watchJob: Job? = null
    private var fadeJob: Job? = null
    private var outgoingPlayer: ExoPlayer? = null
    private var incomingEntryId: Long? = null
    private var failedEntryId: Long? = null

    private val mainListener = object : Player.Listener {
        override fun onPlayWhenReadyChanged(playWhenReady: Boolean, reason: Int) {
            if (!playWhenReady) {
                finish()
            }
        }

        override fun onPlayerError(error: PlaybackException) {
            finish()
        }
    }

    fun start() {
        mainPlayer.addListener(mainListener)
        watchJob = scope.launch {
            while (true) {
                delay(CHECK_INTERVAL_MS)
                startWhenDue()
            }
        }
    }

    fun release() {
        watchJob?.cancel()
        mainPlayer.removeListener(mainListener)
        finish()
    }

    /** Ends a running crossfade at once: the ending song stops, the next one plays at full volume. */
    fun finish() {
        fadeJob?.cancel()
        fadeJob = null
        incomingEntryId = null
        mainPlayer.volume = 1f
        releaseOutgoingPlayer()
    }

    private fun startWhenDue() {
        if (fadeJob != null) {
            return
        }
        val entryId = core.queueState.value.currentEntry?.id
        if (entryId == null || entryId == failedEntryId) {
            return
        }
        if (!CrossfadePlanner.shouldStart(describeSituation())) {
            return
        }
        fadeJob = scope.launch { crossfade(entryId) }
    }

    private fun describeSituation(): CrossfadeSituation {
        val queue = core.queueState.value
        var durationMs: Long? = null
        if (mainPlayer.duration != C.TIME_UNSET) {
            durationMs = mainPlayer.duration
        }
        return CrossfadeSituation(
            enabled = settings.crossfadeEnabled.value,
            fadeSeconds = settings.crossfadeSeconds.value,
            positionMs = mainPlayer.currentPosition,
            durationMs = durationMs,
            isPlaying = mainPlayer.isPlaying,
            hasNextEntry = CrossfadePlanner.hasNextEntry(queue.entries.size, queue.index, queue.repeat),
        )
    }

    private suspend fun crossfade(endingEntryId: Long) {
        val outgoing = handOver()
        if (outgoing == null) {
            failedEntryId = endingEntryId
            finish()
            return
        }
        mainPlayer.volume = 0f
        incomingEntryId = advanceQueue()
        if (incomingEntryId == null) {
            failedEntryId = endingEntryId
            finish()
            return
        }
        ramp(outgoing)
        finish()
    }

    // The second player plays the ending song silently until it runs in step with the main one.
    private suspend fun handOver(): ExoPlayer? {
        val mediaItem = mainPlayer.currentMediaItem
        if (mediaItem == null) {
            return null
        }
        val outgoing = createOutgoingPlayer()
        outgoingPlayer = outgoing
        outgoing.volume = 0f
        outgoing.setMediaItem(mediaItem, mainPlayer.currentPosition)
        outgoing.prepare()
        outgoing.play()
        if (!awaitReady(outgoing)) {
            return null
        }
        outgoing.seekTo(mainPlayer.currentPosition)
        if (!awaitReady(outgoing) || !mainPlayer.isPlaying) {
            return null
        }
        outgoing.volume = 1f
        return outgoing
    }

    private suspend fun awaitReady(player: ExoPlayer): Boolean {
        val ready = withTimeoutOrNull(HAND_OVER_TIMEOUT_MS) {
            try {
                waitForReadyState(player)
                true
            } catch (error: PlaybackException) {
                false
            }
        }
        return ready == true
    }

    private suspend fun waitForReadyState(player: ExoPlayer) {
        suspendCancellableCoroutine { continuation ->
            if (player.playbackState == Player.STATE_READY) {
                continuation.resume(Unit)
                return@suspendCancellableCoroutine
            }
            val listener = object : Player.Listener {
                override fun onPlaybackStateChanged(playbackState: Int) {
                    if (playbackState == Player.STATE_READY) {
                        player.removeListener(this)
                        continuation.resume(Unit)
                    }
                }

                override fun onPlayerError(error: PlaybackException) {
                    player.removeListener(this)
                    continuation.resumeWithException(error)
                }
            }
            player.addListener(listener)
            continuation.invokeOnCancellation { player.removeListener(listener) }
        }
    }

    private suspend fun ramp(outgoing: ExoPlayer) {
        val fadeMs = settings.crossfadeSeconds.value * 1000L
        val startedAt = SystemClock.elapsedRealtime()
        while (true) {
            if (wasInterrupted()) {
                return
            }
            val progress = CrossfadePlanner.progress(SystemClock.elapsedRealtime() - startedAt, fadeMs)
            mainPlayer.volume = CrossfadePlanner.incomingVolume(progress)
            outgoing.volume = CrossfadePlanner.outgoingVolume(progress)
            if (progress >= 1f) {
                return
            }
            delay(FADE_STEP_MS)
        }
    }

    // Picking another song during the fade ends it.
    private fun wasInterrupted(): Boolean = core.queueState.value.currentEntry?.id != incomingEntryId

    private fun releaseOutgoingPlayer() {
        val outgoing = outgoingPlayer
        outgoingPlayer = null
        if (outgoing == null) {
            return
        }
        outgoing.stop()
        outgoing.release()
    }
}
