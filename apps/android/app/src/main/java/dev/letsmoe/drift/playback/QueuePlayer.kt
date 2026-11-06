package dev.letsmoe.drift.playback

import androidx.media3.common.ForwardingPlayer
import androidx.media3.common.Player
import androidx.media3.common.util.UnstableApi

/**
 * The player the MediaSession exposes. ExoPlayer only holds the current entry, so next/previous
 * would normally be unavailable; here they are always advertised and routed to the Go queue.
 */
@androidx.annotation.OptIn(UnstableApi::class)
class QueuePlayer(
    player: Player,
    private val driver: QueuePlaybackDriver,
    private val crossfader: Crossfader,
) : ForwardingPlayer(player) {

    override fun getAvailableCommands(): Player.Commands =
        super.getAvailableCommands()
            .buildUpon()
            .addAll(
                COMMAND_SEEK_TO_NEXT,
                COMMAND_SEEK_TO_NEXT_MEDIA_ITEM,
                COMMAND_SEEK_TO_PREVIOUS,
                COMMAND_SEEK_TO_PREVIOUS_MEDIA_ITEM,
            )
            .build()

    override fun isCommandAvailable(command: Int): Boolean = availableCommands.contains(command)

    override fun hasNextMediaItem(): Boolean = true

    override fun hasPreviousMediaItem(): Boolean = true

    override fun play() {
        driver.play()
    }

    // Pausing, seeking and skipping end a running crossfade at once.
    override fun pause() {
        crossfader.finish()
        super.pause()
    }

    override fun seekTo(positionMs: Long) {
        crossfader.finish()
        super.seekTo(positionMs)
    }

    override fun seekTo(mediaItemIndex: Int, positionMs: Long) {
        crossfader.finish()
        super.seekTo(mediaItemIndex, positionMs)
    }

    override fun seekBack() {
        crossfader.finish()
        super.seekBack()
    }

    override fun seekForward() {
        crossfader.finish()
        super.seekForward()
    }

    override fun seekToNext() {
        crossfader.finish()
        driver.skipToNext()
    }

    override fun seekToNextMediaItem() {
        crossfader.finish()
        driver.skipToNext()
    }

    override fun seekToPrevious() {
        crossfader.finish()
        driver.skipToPrevious()
    }

    override fun seekToPreviousMediaItem() {
        crossfader.finish()
        driver.skipToPrevious()
    }
}
