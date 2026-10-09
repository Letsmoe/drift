package dev.letsmoe.drift.playback

import android.util.Log
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.QueueEntry
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Deferred
import kotlinx.coroutines.async
import kotlinx.coroutines.launch

private const val LOG_TAG = "PlayReporter"

/**
 * Reports one play at a time to the library (startPlay once, then updatePlay with progress), the
 * way player.svelte.ts does. Failures are logged and never interrupt playback.
 */
class PlayReporter(
    private val core: DriftCore,
    private val scope: CoroutineScope,
) {
    private var playId: Deferred<Long?>? = null

    fun start(entry: QueueEntry, playContext: String) {
        playId = scope.async {
            try {
                core.startPlay(entry.track, entry.kind, playContext)
            } catch (error: Exception) {
                Log.w(LOG_TAG, "startPlay failed", error)
                null
            }
        }
    }

    /** Reports progress of the current play; [completed] also closes it. */
    fun report(progressSeconds: Int, completed: Boolean) {
        val pendingPlayId = playId
        if (pendingPlayId == null) {
            return
        }
        if (completed) {
            playId = null
        }
        scope.launch {
            sendProgress(pendingPlayId.await(), progressSeconds, completed)
        }
    }

    private suspend fun sendProgress(resolvedPlayId: Long?, progressSeconds: Int, completed: Boolean) {
        if (resolvedPlayId == null) {
            return
        }
        try {
            core.updatePlay(resolvedPlayId, progressSeconds, completed)
        } catch (error: Exception) {
            Log.w(LOG_TAG, "updatePlay failed", error)
        }
    }
}
