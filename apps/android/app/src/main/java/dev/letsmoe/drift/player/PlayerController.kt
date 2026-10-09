package dev.letsmoe.drift.player

import android.content.ComponentName
import android.content.Context
import androidx.media3.common.C
import androidx.media3.common.Player
import androidx.media3.session.MediaController
import androidx.media3.session.SessionToken
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.QueueEntry
import dev.letsmoe.drift.core.model.QueueState
import dev.letsmoe.drift.core.model.RepeatMode
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.playback.PlaybackEvents
import dev.letsmoe.drift.playback.PlaybackService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.guava.await
import kotlinx.coroutines.launch

private const val POSITION_TICK_MS = 250L

/**
 * What the UI uses to play music. Queue commands go straight to the Go core; the playback service
 * sees the queue change and loads the new current entry into the player (see
 * QueuePlaybackDriver), so there is a single path from "queue changed" to "player changed".
 * Transport commands (play, pause, seek, next, previous) go through the MediaController so the
 * notification, headset buttons and the app all behave the same.
 */
class PlayerController(
    private val appContext: Context,
    private val core: DriftCore,
    private val playbackEvents: PlaybackEvents,
    private val scope: CoroutineScope,
) {
    private var mediaController: MediaController? = null

    val queue: StateFlow<QueueState> = core.queueState

    val currentEntry: StateFlow<QueueEntry?> =
        queue.map { state -> state.currentEntry }.stateIn(scope, SharingStarted.Eagerly, null)

    val currentTrack: StateFlow<Track?> =
        currentEntry.map { entry -> entry?.track }.stateIn(scope, SharingStarted.Eagerly, null)

    private val mutableIsPlaying = MutableStateFlow(false)

    /** True when playback is running or about to run (buffering included). */
    val isPlaying: StateFlow<Boolean> = mutableIsPlaying.asStateFlow()

    private val mutableIsBuffering = MutableStateFlow(false)
    val isBuffering: StateFlow<Boolean> = mutableIsBuffering.asStateFlow()

    private val mutablePositionMs = MutableStateFlow(0L)
    val positionMs: StateFlow<Long> = mutablePositionMs.asStateFlow()

    private val mutableDurationMs = MutableStateFlow(0L)

    /** Falls back to the track's catalog duration until the stream reports its own. */
    val durationMs: StateFlow<Long> = mutableDurationMs.asStateFlow()

    val errorMessage: StateFlow<String> = playbackEvents.errorMessage

    private val mutablePlayRequests = MutableSharedFlow<Unit>(extraBufferCapacity = 1)

    /** Emits when the user starts playback; the Activity asks for the notification permission. */
    val playRequests: SharedFlow<Unit> = mutablePlayRequests.asSharedFlow()

    private val controllerListener = object : Player.Listener {
        override fun onEvents(player: Player, events: Player.Events) {
            syncFromController(player)
        }
    }

    fun connect() {
        scope.launch {
            val token = SessionToken(appContext, ComponentName(appContext, PlaybackService::class.java))
            val controller = MediaController.Builder(appContext, token).buildAsync().await()
            controller.addListener(controllerListener)
            mediaController = controller
            syncFromController(controller)
            runPositionTicker()
        }
        scope.launch {
            currentEntry.collect { syncDurationFallback() }
        }
    }

    private suspend fun runPositionTicker() {
        while (true) {
            delay(POSITION_TICK_MS)
            val controller = mediaController
            if (controller != null && mutableIsPlaying.value) {
                mutablePositionMs.value = controller.currentPosition
            }
        }
    }

    private fun syncFromController(player: Player) {
        val wantsToPlay = player.playWhenReady
        val state = player.playbackState
        mutableIsPlaying.value = wantsToPlay && (state == Player.STATE_READY || state == Player.STATE_BUFFERING)
        mutableIsBuffering.value = wantsToPlay && state == Player.STATE_BUFFERING
        mutablePositionMs.value = player.currentPosition
        if (player.duration != C.TIME_UNSET) {
            mutableDurationMs.value = player.duration
            return
        }
        syncDurationFallback()
    }

    private fun syncDurationFallback() {
        val track = currentTrack.value
        if (track == null) {
            mutableDurationMs.value = 0
            return
        }
        val controllerDuration = mediaController?.duration ?: C.TIME_UNSET
        if (controllerDuration == C.TIME_UNSET) {
            mutableDurationMs.value = track.durationSeconds * 1000L
        }
    }

    // Queue commands

    fun playTracks(tracks: List<Track>, kind: ItemKind = ItemKind.Song, startIndex: Int = 0, context: String = "track") {
        mutablePlayRequests.tryEmit(Unit)
        runQueueCommand { core.queuePlay(tracks, kind, startIndex, context) }
    }

    fun playNext(tracks: List<Track>) = runQueueCommand { core.queuePlayNext(tracks) }

    fun addToQueue(tracks: List<Track>) = runQueueCommand { core.queueAdd(tracks) }

    fun jump(entryId: Long) = runQueueCommand { core.queueJump(entryId) }

    fun remove(entryId: Long) = runQueueCommand { core.queueRemove(entryId) }

    fun move(entryId: Long, newIndex: Int) = runQueueCommand { core.queueMove(entryId, newIndex) }

    fun clearQueue() = runQueueCommand { core.queueClear() }

    fun setShuffle(enabled: Boolean) = runQueueCommand { core.queueSetShuffle(enabled) }

    fun toggleShuffle() = setShuffle(!queue.value.shuffle)

    /** Cycles repeat off, all, one, off. */
    fun cycleRepeat() {
        val nextMode = when (queue.value.repeat) {
            RepeatMode.Off -> RepeatMode.All
            RepeatMode.All -> RepeatMode.One
            RepeatMode.One -> RepeatMode.Off
        }
        runQueueCommand { core.queueSetRepeat(nextMode) }
    }

    fun setAutoplay(enabled: Boolean) = runQueueCommand { core.queueSetAutoplay(enabled) }

    /** Plays the track when it is not current, otherwise pauses or resumes it. */
    fun playOrToggle(
        track: Track,
        tracks: List<Track> = listOf(track),
        startIndex: Int = 0,
        context: String = "track",
        kind: ItemKind = ItemKind.Song,
    ) {
        if (isCurrent(track)) {
            toggle()
            return
        }
        playTracks(tracks, kind, startIndex, context)
    }

    /** True when the queue was started from [context] (`album:<id>`, `liked`, ...) and has a current track. */
    fun isContextCurrent(context: String): Boolean = queue.value.context == context && currentTrack.value != null

    /** Starts the context from the top (or a random track with shuffle on), or toggles it when it is current. */
    fun playContext(tracks: List<Track>, context: String, kind: ItemKind = ItemKind.Song) {
        if (tracks.isEmpty()) {
            return
        }
        if (isContextCurrent(context)) {
            toggle()
            return
        }
        var startIndex = 0
        if (queue.value.shuffle) {
            startIndex = tracks.indices.random()
        }
        playTracks(tracks, kind, startIndex, context)
    }

    /** Turns shuffle on and starts the context at a random track. */
    fun shuffleContext(tracks: List<Track>, context: String, kind: ItemKind = ItemKind.Song) {
        if (tracks.isEmpty()) {
            return
        }
        mutablePlayRequests.tryEmit(Unit)
        runQueueCommand {
            core.queueSetShuffle(true)
            core.queuePlay(tracks, kind, tracks.indices.random(), context)
        }
    }

    fun isCurrent(track: Track): Boolean = currentTrack.value?.videoId == track.videoId

    private fun runQueueCommand(command: suspend () -> Unit) {
        scope.launch {
            try {
                command()
            } catch (error: Exception) {
                playbackEvents.reportError(error.message ?: error.javaClass.simpleName)
            }
        }
    }

    // Transport commands

    fun toggle() {
        val controller = mediaController
        if (controller == null || currentTrack.value == null) {
            return
        }
        if (mutableIsPlaying.value) {
            controller.pause()
            return
        }
        mutablePlayRequests.tryEmit(Unit)
        controller.play()
    }

    fun next() {
        mediaController?.seekToNext()
    }

    fun previous() {
        mediaController?.seekToPrevious()
    }

    fun seekTo(positionMs: Long) {
        mutablePositionMs.value = positionMs
        mediaController?.seekTo(positionMs)
    }
}
