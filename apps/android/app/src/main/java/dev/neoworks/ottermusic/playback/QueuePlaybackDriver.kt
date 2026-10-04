package dev.neoworks.ottermusic.playback

import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.PlaybackException
import androidx.core.net.toUri
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.QueueEntry
import dev.neoworks.ottermusic.core.model.QueueState
import dev.neoworks.ottermusic.core.model.bestFor
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/** Seconds into a track after which Previous restarts it instead of going back. */
private const val RESTART_THRESHOLD_MS = 3_000L
private const val PROGRESS_REPORT_INTERVAL_MS = 15_000L
private const val ARTWORK_TARGET_PIXELS = 544

/**
 * Keeps the ExoPlayer in step with the Go queue, which is the single source of truth.
 *
 * This is the only place that loads queue entries into the player: whenever the queue's current
 * entry id changes (UI commands, notification next/previous, autoplay) the new entry is loaded
 * here. Everything else only talks to the core. Mirrors player.svelte.ts.
 */
class QueuePlaybackDriver(
    private val player: ExoPlayer,
    private val core: OtterCore,
    private val events: PlaybackEvents,
    private val scope: CoroutineScope,
) {
    private val reporter = PlayReporter(core, scope)
    private var loadedEntryId: Long? = null
    private var loadJob: Job? = null
    private var hasInitialState = false

    private val playerListener = object : Player.Listener {
        override fun onPlaybackStateChanged(playbackState: Int) {
            if (playbackState == Player.STATE_ENDED) {
                scope.launch { trackEnded() }
            }
        }

        override fun onIsPlayingChanged(isPlaying: Boolean) {
            if (!isPlaying) {
                reportProgress(completed = false)
            }
        }

        override fun onPlayerError(error: PlaybackException) {
            fail(error.message ?: "Playback failed")
        }
    }

    fun start() {
        player.addListener(playerListener)
        scope.launch { core.queueState.collect { state -> applyState(state) } }
        scope.launch { reportWhilePlaying() }
    }

    fun release() {
        reportProgress(completed = false)
        player.removeListener(playerListener)
    }

    /** Play from the notification, session or UI; loads the entry first when nothing is loaded yet. */
    fun play() {
        val entry = core.queueState.value.currentEntry
        if (entry != null && needsLoad(entry)) {
            loadEntry(entry)
            return
        }
        player.play()
    }

    /** A user skip: the queue moves on, or playback stops at the end of the queue. */
    fun skipToNext() {
        scope.launch {
            try {
                val step = core.queueNext()
                if (!step.moved) {
                    stopAtEnd()
                }
            } catch (error: Exception) {
                fail(describe(error))
            }
        }
    }

    /** Restarts the track past 3 s, otherwise goes back one entry (restarting when there is none). */
    fun skipToPrevious() {
        if (player.currentPosition > RESTART_THRESHOLD_MS) {
            player.seekTo(0)
            return
        }
        scope.launch { goToPreviousEntry() }
    }

    private suspend fun goToPreviousEntry() {
        val entryIdBefore = core.queueState.value.currentEntry?.id
        try {
            val state = core.queuePrevious()
            if (state.currentEntry?.id == entryIdBefore) {
                player.seekTo(0)
            }
        } catch (error: Exception) {
            fail(describe(error))
        }
    }

    private fun needsLoad(entry: QueueEntry): Boolean = entry.id != loadedEntryId || events.errorMessage.value.isNotEmpty()

    private fun applyState(state: QueueState) {
        val entry = state.currentEntry
        if (!hasInitialState) {
            // A queue that survived an Activity restart is shown, not started.
            hasInitialState = true
            if (entry != null) {
                return
            }
        }
        if (entry == null) {
            unload()
            return
        }
        if (entry.id != loadedEntryId) {
            loadEntry(entry)
        }
    }

    private fun loadEntry(entry: QueueEntry) {
        reportProgress(completed = false)
        loadedEntryId = entry.id
        events.clearError()
        reporter.start(entry, core.queueState.value.context)
        loadJob?.cancel()
        loadJob = scope.launch {
            val resumeMs = lookUpResumeMs(entry)
            player.setMediaItem(buildMediaItem(entry), resumeMs)
            player.prepare()
            player.play()
        }
    }

    private suspend fun lookUpResumeMs(entry: QueueEntry): Long {
        if (entry.kind != ItemKind.Episode) {
            return 0
        }
        try {
            val position = core.resumePosition(entry.track.videoId)
            if (position.found) {
                return position.seconds * 1000L
            }
        } catch (error: Exception) {
            // Resuming is a convenience; start from the beginning instead.
        }
        return 0
    }

    private fun buildMediaItem(entry: QueueEntry): MediaItem {
        val track = entry.track
        val artworkUrl = track.thumbnails.bestFor(ARTWORK_TARGET_PIXELS)?.url
        val metadata = MediaMetadata.Builder()
            .setTitle(track.title)
            .setArtist(track.artistNames)
            .setAlbumTitle(track.album?.name)
            .setDurationMs(track.durationSeconds * 1000L)
            .setIsPlayable(true)
        if (artworkUrl != null) {
            metadata.setArtworkUri(artworkUrl.toUri())
        }
        return MediaItem.Builder()
            .setMediaId(track.videoId)
            .setUri(core.streamUrl(track.videoId))
            .setMediaMetadata(metadata.build())
            .build()
    }

    private fun unload() {
        reportProgress(completed = false)
        loadJob?.cancel()
        loadedEntryId = null
        events.clearError()
        player.stop()
        player.clearMediaItems()
    }

    private fun stopAtEnd() {
        player.pause()
        player.seekTo(0)
    }

    private suspend fun trackEnded() {
        reportProgress(completed = true)
        val endedEntryId = core.queueState.value.currentEntry?.id
        try {
            val step = core.queueAdvance()
            if (!step.moved) {
                stopAtEnd()
                return
            }
            if (step.state.currentEntry?.id == endedEntryId) {
                restartCurrent(step.state)
            }
        } catch (error: Exception) {
            fail(describe(error))
        }
    }

    // Repeat one: the same entry plays again, as a new play.
    private fun restartCurrent(state: QueueState) {
        val entry = state.currentEntry
        if (entry == null) {
            return
        }
        reporter.start(entry, state.context)
        player.seekTo(0)
        player.play()
    }

    private fun fail(message: String) {
        reportProgress(completed = false)
        events.reportError(message)
        player.pause()
    }

    private fun reportProgress(completed: Boolean) {
        val progressSeconds = (player.currentPosition / 1000).toInt()
        reporter.report(progressSeconds, completed)
    }

    private suspend fun reportWhilePlaying() {
        while (true) {
            delay(PROGRESS_REPORT_INTERVAL_MS)
            if (player.isPlaying) {
                reportProgress(completed = false)
            }
        }
    }

    private fun describe(error: Exception): String = error.message ?: error.javaClass.simpleName
}
