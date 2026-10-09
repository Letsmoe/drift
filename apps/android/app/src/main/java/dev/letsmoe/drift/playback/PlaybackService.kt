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
    private var mediaSession: MediaSession? = null

    override fun onCreate() {
        super.onCreate()
        val container = (application as DriftApplication).container
        exoPlayer = buildExoPlayer()
        driver = QueuePlaybackDriver(exoPlayer, container.core, container.playbackEvents, serviceScope)
        driver.start()
        mediaSession = MediaSession.Builder(this, QueuePlayer(exoPlayer, driver)).build()
    }

    private fun buildExoPlayer(): ExoPlayer {
        val musicAttributes = AudioAttributes.Builder()
            .setUsage(C.USAGE_MEDIA)
            .setContentType(C.AUDIO_CONTENT_TYPE_MUSIC)
            .build()
        return ExoPlayer.Builder(this)
            .setAudioAttributes(musicAttributes, true)
            .setHandleAudioBecomingNoisy(true)
            .setWakeMode(C.WAKE_MODE_NETWORK)
            .build()
    }

    override fun onGetSession(controllerInfo: MediaSession.ControllerInfo): MediaSession? = mediaSession

    override fun onTaskRemoved(rootIntent: Intent?) {
        if (!exoPlayer.playWhenReady) {
            stopSelf()
        }
    }

    override fun onDestroy() {
        driver.release()
        mediaSession?.release()
        mediaSession = null
        exoPlayer.release()
        serviceScope.cancel()
        super.onDestroy()
    }
}
