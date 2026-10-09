package dev.letsmoe.drift.library

import android.util.Log
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.Impression
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

private const val FLUSH_INTERVAL_MS = 4_000L
private const val MAX_BATCH_SIZE = 50

/** Collects recommendation impressions and clicks and writes them to the library in batches. */
class ImpressionLogger(
    private val core: DriftCore,
    private val scope: CoroutineScope,
) {
    private var pending = mutableListOf<Impression>()
    private var flushTimer: Job? = null

    /** An item was shown to the user. */
    fun logImpression(videoId: String, source: String, surface: String, position: Int) {
        enqueue(Impression(videoId, source, surface, position, clicked = false))
    }

    /** The user clicked an item that was shown. */
    fun logClick(videoId: String, source: String, surface: String, position: Int) {
        enqueue(Impression(videoId, source, surface, position, clicked = true))
    }

    fun flush() {
        flushTimer?.cancel()
        flushTimer = null
        if (pending.isEmpty()) {
            return
        }
        val batch = pending
        pending = mutableListOf()
        scope.launch { send(batch) }
    }

    private fun enqueue(impression: Impression) {
        pending.add(impression)
        if (pending.size >= MAX_BATCH_SIZE) {
            flush()
            return
        }
        if (flushTimer == null) {
            flushTimer = scope.launch {
                delay(FLUSH_INTERVAL_MS)
                flush()
            }
        }
    }

    private suspend fun send(batch: List<Impression>) {
        try {
            core.recordImpressions(batch)
        } catch (error: Exception) {
            Log.w("ImpressionLogger", "recordImpressions failed", error)
        }
    }
}
