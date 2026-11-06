package dev.letsmoe.drift.playback

import android.content.Intent
import androidx.media3.common.AudioAttributes
import androidx.media3.common.C
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.session.MediaSession
import androidx.media3.session.MediaSessionService
import dev.letsmoe.drift.DriftApplication
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel

/** Background playback: owns the ExoPlayer, the media session and the media notification. */
class PlaybackService : MediaSessionService() {
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    private lateinit var exoPlayer: ExoPlayer
    private lateinit var driver: QueuePlaybackDriver
    private lateinit var crossfader: Crossfader
    private var mediaSession: MediaSession? = null

    override fun onCreate() {
        super.onCreate()
        val container = (application as DriftApplication).container
        exoPlayer = buildExoPlayer()
        driver = QueuePlaybackDriver(exoPlayer, container.core, container.playbackEvents, serviceScope)
        crossfader = Crossfader(
            mainPlayer = exoPlayer,
            createOutgoingPlayer = ::buildOutgoingPlayer,
            settings = container.settings,
            core = container.core,
            scope = serviceScope,
            advanceQueue = { driver.advanceForCrossfade() },
        )
        driver.start()
        crossfader.start()
        mediaSession = MediaSession.Builder(this, QueuePlayer(exoPlayer, driver, crossfader)).build()
    }

    private fun buildExoPlayer(): ExoPlayer {
        return ExoPlayer.Builder(this)
            .setAudioAttributes(musicAudioAttributes(), true)
            .setHandleAudioBecomingNoisy(true)
            .setWakeMode(C.WAKE_MODE_NETWORK)
            .build()
    }

    // The ending song during a crossfade; the main player keeps the audio focus and the session.
    private fun buildOutgoingPlayer(): ExoPlayer {
        return ExoPlayer.Builder(this)
            .setAudioAttributes(musicAudioAttributes(), false)
            .setWakeMode(C.WAKE_MODE_NETWORK)
            .build()
    }

    private fun musicAudioAttributes(): AudioAttributes {
        return AudioAttributes.Builder()
            .setUsage(C.USAGE_MEDIA)
            .setContentType(C.AUDIO_CONTENT_TYPE_MUSIC)
            .build()
    }

    override fun onGetSession(controllerInfo: MediaSession.ControllerInfo): MediaSession? = mediaSession

    override fun onTaskRemoved(rootIntent: Intent?) {
        if (!exoPlayer.playWhenReady) {
            stopSelf()
        }
    }

    override fun onDestroy() {
        crossfader.release()
        driver.release()
        mediaSession?.release()
        mediaSession = null
        exoPlayer.release()
        serviceScope.cancel()
        super.onDestroy()
    }
}
