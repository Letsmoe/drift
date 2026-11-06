package dev.letsmoe.drift.downloads

import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.CollectionKind
import dev.letsmoe.drift.core.model.DownloadCollection
import dev.letsmoe.drift.core.model.DownloadItem
import dev.letsmoe.drift.core.model.DownloadState
import dev.letsmoe.drift.core.model.DownloadStatus
import dev.letsmoe.drift.core.model.DownloadSummary
import dev.letsmoe.drift.core.model.DownloadUpdate
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.StorageUsage
import dev.letsmoe.drift.core.model.Track
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicInteger

private const val LIST_REFRESH_DELAY_MILLISECONDS = 200L
private const val STATE_LOOKUP_DELAY_MILLISECONDS = 20L

/**
 * Offline downloads: the state of each track, the collections kept downloaded and the lists the
 * Downloads page shows. The core owns the queue; this class mirrors what its listener reports.
 */
class DownloadsState(
    private val core: DriftCore,
    private val scope: CoroutineScope,
) {
    private val mutableStatuses = MutableStateFlow<Map<String, DownloadStatus>>(emptyMap())

    /** Status of every track looked up so far; tracks that were never queued are absent. */
    val statuses: StateFlow<Map<String, DownloadStatus>> = mutableStatuses.asStateFlow()

    private val mutableSummary = MutableStateFlow(DownloadSummary())
    val summary: StateFlow<DownloadSummary> = mutableSummary.asStateFlow()

    private val mutableCollections = MutableStateFlow<List<DownloadCollection>>(emptyList())
    val collections: StateFlow<List<DownloadCollection>> = mutableCollections.asStateFlow()

    private val mutableQueue = MutableStateFlow<List<DownloadItem>>(emptyList())

    /** Queued, running and failed downloads; filled while a screen watches the lists. */
    val queue: StateFlow<List<DownloadItem>> = mutableQueue.asStateFlow()

    private val mutableDownloadedTracks = MutableStateFlow<List<Track>>(emptyList())

    /** Finished downloads, playable offline; filled while a screen watches the lists. */
    val downloadedTracks: StateFlow<List<Track>> = mutableDownloadedTracks.asStateFlow()

    private val mutableUsage = MutableStateFlow(StorageUsage())
    val usage: StateFlow<StorageUsage> = mutableUsage.asStateFlow()

    private val mutableErrorMessage = MutableStateFlow("")

    /** The last background failure; cleared by the next successful refresh. */
    val errorMessage: StateFlow<String> = mutableErrorMessage.asStateFlow()

    private val requestedVideoIds = ConcurrentHashMap.newKeySet<String>()
    private val pendingVideoIds = ConcurrentHashMap.newKeySet<String>()
    private val stateLookupScheduled = AtomicBoolean(false)
    private val listRefreshScheduled = AtomicBoolean(false)
    private val listWatchers = AtomicInteger(0)
    private var initialLoad: Job? = null

    /** Loads collections and the summary and starts listening for updates; call once at startup. */
    fun start() {
        core.listenToDownloads { update -> applyUpdate(update) }
        initialLoad = scope.launch {
            refreshCollections()
            refreshSummary()
        }
    }

    /** A screen showing the queue or the downloaded songs calls this while it is open. */
    fun startWatchingLists() {
        listWatchers.incrementAndGet()
        scope.launch { refreshLists() }
    }

    fun stopWatchingLists() {
        listWatchers.decrementAndGet()
    }

    fun stateOf(videoId: String): DownloadState {
        val status = mutableStatuses.value[videoId]
        if (status == null) {
            return DownloadState.None
        }
        return status.state
    }

    fun collection(kind: CollectionKind, id: String): DownloadCollection? =
        mutableCollections.value.firstOrNull { candidate -> candidate.kind == kind && candidate.id == id }

    /** Looks up the states of [videoIds] not looked up yet; lookups made close together go out as one call. */
    fun ensureStates(videoIds: List<String>) {
        for (videoId in videoIds) {
            if (requestedVideoIds.add(videoId)) {
                pendingVideoIds.add(videoId)
            }
        }
        if (pendingVideoIds.isEmpty() || !stateLookupScheduled.compareAndSet(false, true)) {
            return
        }
        scope.launch {
            delay(STATE_LOOKUP_DELAY_MILLISECONDS)
            stateLookupScheduled.set(false)
            lookUpPendingStates()
        }
    }

    suspend fun download(track: Track, kind: ItemKind = ItemKind.Song) {
        core.downloadTracks(listOf(track), kind)
    }

    suspend fun cancel(videoId: String) {
        core.cancelDownload(videoId)
    }

    suspend fun retry(videoId: String) {
        core.retryDownload(videoId)
    }

    suspend fun remove(videoId: String) {
        core.deleteDownload(videoId)
    }

    suspend fun deleteAll() {
        core.deleteAllDownloads()
        refreshCollections()
    }

    suspend fun pause() {
        core.pauseDownloads()
    }

    suspend fun resume() {
        core.resumeDownloads()
    }

    /** Downloads a whole collection and keeps it in sync as it changes. */
    suspend fun keepDownloaded(kind: CollectionKind, id: String) {
        core.downloadCollection(kind, id, true)
        refreshCollections()
    }

    /** Deletes the collection's downloads that no other collection or single download needs. */
    suspend fun removeCollection(kind: CollectionKind, id: String) {
        core.removeDownloadedCollection(kind, id)
        refreshCollections()
    }

    /** Called when a collection's page opens: a kept YouTube collection checks for new tracks. */
    suspend fun refreshOnOpen(kind: CollectionKind, id: String) {
        initialLoad?.join()
        val collection = collection(kind, id)
        if (collection == null || !collection.keep) {
            return
        }
        try {
            core.refreshDownloadedCollection(kind, id)
            refreshCollections()
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (_: Exception) {
            // Offline or YouTube unreachable: the stored copy stays as it is.
        }
    }

    private suspend fun lookUpPendingStates() {
        val videoIds = pendingVideoIds.toList()
        pendingVideoIds.removeAll(videoIds.toSet())
        if (videoIds.isEmpty()) {
            return
        }
        try {
            storeLookedUpStates(videoIds, core.downloadStates(videoIds))
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            requestedVideoIds.removeAll(videoIds.toSet())
            reportError(error)
        }
    }

    // An update that arrived while the lookup was in flight is newer than the lookup.
    private fun storeLookedUpStates(videoIds: List<String>, found: Map<String, DownloadStatus>) {
        mutableStatuses.update { current ->
            val merged = current.toMutableMap()
            for (videoId in videoIds) {
                val status = found[videoId]
                if (videoId !in current && status != null) {
                    merged[videoId] = status
                }
            }
            merged
        }
    }

    private fun applyUpdate(update: DownloadUpdate) {
        mutableSummary.value = update.summary
        val videoId = update.status.videoId
        if (videoId.isEmpty()) {
            scheduleListRefresh()
            return
        }
        val previousState = stateOf(videoId)
        requestedVideoIds.add(videoId)
        storeStatus(update.status)
        if (previousState != update.status.state) {
            scheduleListRefresh()
        }
    }

    private fun storeStatus(status: DownloadStatus) {
        mutableStatuses.update { current ->
            if (status.state == DownloadState.None) {
                current - status.videoId
            } else {
                current + (status.videoId to status)
            }
        }
    }

    // State changes arrive in bursts (a collection queues dozens of tracks), so lists reload once per burst.
    private fun scheduleListRefresh() {
        if (!listRefreshScheduled.compareAndSet(false, true)) {
            return
        }
        scope.launch {
            delay(LIST_REFRESH_DELAY_MILLISECONDS)
            listRefreshScheduled.set(false)
            refreshLists()
        }
    }

    private suspend fun refreshLists() {
        refreshCollections()
        if (listWatchers.get() <= 0) {
            return
        }
        try {
            mutableQueue.value = core.downloadQueue()
            mutableDownloadedTracks.value = core.downloadedTracks()
            mutableUsage.value = core.storageUsage()
            mutableErrorMessage.value = ""
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            reportError(error)
        }
    }

    private suspend fun refreshCollections() {
        try {
            mutableCollections.value = core.downloadedCollections()
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            reportError(error)
        }
    }

    private suspend fun refreshSummary() {
        try {
            mutableSummary.value = core.downloadSummary()
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            reportError(error)
        }
    }

    private fun reportError(error: Exception) {
        mutableErrorMessage.value = error.message ?: error.javaClass.simpleName
    }
}
