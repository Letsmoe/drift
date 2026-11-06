package dev.letsmoe.drift.core

import dev.letsmoe.drift.core.model.AlbumPage
import dev.letsmoe.drift.core.model.ArtistPage
import dev.letsmoe.drift.core.model.CoverColors
import dev.letsmoe.drift.core.model.HomePage
import dev.letsmoe.drift.core.model.ImportOptions
import dev.letsmoe.drift.core.model.ImportProgress
import dev.letsmoe.drift.core.model.ImportResult
import dev.letsmoe.drift.core.model.ImportSourceKind
import dev.letsmoe.drift.core.model.ImportSummary
import dev.letsmoe.drift.core.model.Impression
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.LibraryPlaylist
import dev.letsmoe.drift.core.model.PlaylistPage
import dev.letsmoe.drift.core.model.PlaylistSummary
import dev.letsmoe.drift.core.model.PlaylistTracks
import dev.letsmoe.drift.core.model.PodcastEpisodes
import dev.letsmoe.drift.core.model.PodcastPage
import dev.letsmoe.drift.core.model.QueueState
import dev.letsmoe.drift.core.model.QueueStep
import dev.letsmoe.drift.core.model.RecommendedItem
import dev.letsmoe.drift.core.model.RecommendedRow
import dev.letsmoe.drift.core.model.RepeatMode
import dev.letsmoe.drift.core.model.ResumePosition
import dev.letsmoe.drift.core.model.SavedItem
import dev.letsmoe.drift.core.model.SearchResults
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.core.model.YouTubeAccountStatus
import dev.letsmoe.drift.mobile.ImportListener
import dev.letsmoe.drift.mobile.Core
import dev.letsmoe.drift.mobile.Listener
import dev.letsmoe.drift.mobile.Mobile
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import kotlinx.serialization.encodeToString

/**
 * Typed bridge to the Go core. Every call runs on [Dispatchers.IO] and throws when the core
 * reports an error. The core owns all logic; this class only encodes, decodes and forwards.
 */
class DriftCore(dataDirectory: String) {
    private val mutableQueueState = MutableStateFlow(QueueState.Empty)

    /** Latest queue snapshot, fed by the core's change listener and by every queue command. */
    val queueState: StateFlow<QueueState> = mutableQueueState.asStateFlow()

    private val core: Core = Mobile.open(dataDirectory, Listener { stateJson -> publishQueueState(stateJson) })

    private val dispatcher: CoroutineDispatcher = Dispatchers.IO

    private fun publishQueueState(stateJson: String) {
        mutableQueueState.value = DriftJson.decodeFromString<QueueState>(stateJson)
    }

    private suspend fun <T> call(block: () -> T): T = withContext(dispatcher) { block() }

    private suspend fun callQueue(block: () -> String): QueueState =
        call {
            val state = DriftJson.decodeFromString<QueueState>(block())
            mutableQueueState.value = state
            state
        }

    private suspend fun callStep(block: () -> String): QueueStep =
        call {
            val step = DriftJson.decodeFromString<QueueStep>(block())
            mutableQueueState.value = step.state
            step
        }

    fun close() {
        core.close()
    }

    /** Synchronous: only builds a loopback URL string. */
    fun streamUrl(videoId: String): String = core.streamURL(videoId)

    // Catalog

    suspend fun home(): HomePage = call { DriftJson.decodeFromString(core.home()) }

    suspend fun homeContinuation(token: String): HomePage = call { DriftJson.decodeFromString(core.homeContinuation(token)) }

    suspend fun search(query: String): SearchResults = call { DriftJson.decodeFromString(core.search(query)) }

    suspend fun searchPodcasts(query: String): SearchResults =
        call { DriftJson.decodeFromString(core.searchPodcasts(query)) }

    suspend fun searchSuggestions(query: String): List<String> =
        call { DriftJson.decodeFromString(core.searchSuggestions(query)) }

    suspend fun album(browseId: String): AlbumPage = call { DriftJson.decodeFromString(core.album(browseId)) }

    suspend fun playlist(browseId: String): PlaylistPage = call { DriftJson.decodeFromString(core.playlist(browseId)) }

    suspend fun playlistContinuation(token: String): PlaylistTracks =
        call { DriftJson.decodeFromString(core.playlistContinuation(token)) }

    suspend fun artist(browseId: String): ArtistPage = call { DriftJson.decodeFromString(core.artist(browseId)) }

    suspend fun radio(videoId: String): List<Track> = call { DriftJson.decodeFromString(core.radio(videoId)) }

    suspend fun podcast(browseId: String): PodcastPage = call { DriftJson.decodeFromString(core.podcast(browseId)) }

    suspend fun podcastShows(): HomePage = call { DriftJson.decodeFromString(core.podcastShows()) }

    suspend fun podcastShowsContinuation(token: String): HomePage =
        call { DriftJson.decodeFromString(core.podcastShowsContinuation(token)) }

    suspend fun podcastContinuation(token: String): PodcastEpisodes =
        call { DriftJson.decodeFromString(core.podcastContinuation(token)) }

    // Library

    /** [rating] is 1 (like), -1 (dislike) or 0 (clear). */
    suspend fun setRating(track: Track, kind: ItemKind, rating: Int) {
        call { core.setRating(DriftJson.encodeToString(track), kind.wireName, rating.toLong()) }
    }

    suspend fun ratings(videoIds: List<String>): Map<String, Int> =
        call { DriftJson.decodeFromString(core.ratings(DriftJson.encodeToString(videoIds))) }

    suspend fun likedTracks(): List<Track> = call { DriftJson.decodeFromString(core.likedTracks()) }

    suspend fun playlists(): List<PlaylistSummary> = call { DriftJson.decodeFromString(core.playlists()) }

    suspend fun libraryPlaylist(id: String): LibraryPlaylist =
        call { DriftJson.decodeFromString(core.libraryPlaylist(id)) }

    suspend fun createPlaylist(name: String): LibraryPlaylist =
        call { DriftJson.decodeFromString(core.createPlaylist(name)) }

    suspend fun renamePlaylist(id: String, name: String, description: String) {
        call { core.renamePlaylist(id, name, description) }
    }

    suspend fun deletePlaylist(id: String) {
        call { core.deletePlaylist(id) }
    }

    suspend fun addToPlaylist(id: String, tracks: List<Track>) {
        call { core.addToPlaylist(id, DriftJson.encodeToString(tracks)) }
    }

    suspend fun removeFromPlaylist(id: String, entryId: Long) {
        call { core.removeFromPlaylist(id, entryId) }
    }

    suspend fun movePlaylistEntry(id: String, entryId: Long, newIndex: Int) {
        call { core.movePlaylistEntry(id, entryId, newIndex.toLong()) }
    }

    suspend fun save(item: Item) {
        call { core.save(DriftJson.encodeToString(item)) }
    }

    suspend fun unsave(kind: ItemKind, id: String) {
        call { core.unsave(kind.wireName, id) }
    }

    suspend fun isSaved(kind: ItemKind, id: String): Boolean = call { core.isSaved(kind.wireName, id) }

    suspend fun savedItems(): List<SavedItem> = call { DriftJson.decodeFromString(core.savedItems()) }

    /** Returns the play id to pass to [updatePlay]. */
    suspend fun startPlay(track: Track, kind: ItemKind, playContext: String): Long =
        call { core.startPlay(DriftJson.encodeToString(track), kind.wireName, playContext) }

    suspend fun updatePlay(playId: Long, progressSeconds: Int, completed: Boolean) {
        call { core.updatePlay(playId, progressSeconds.toLong(), completed) }
    }

    suspend fun resumePosition(videoId: String): ResumePosition =
        call { DriftJson.decodeFromString(core.resumePosition(videoId)) }

    suspend fun recentTracks(limit: Int): List<Track> = call { DriftJson.decodeFromString(core.recentTracks(limit.toLong())) }

    suspend fun recentSearches(limit: Int): List<String> =
        call { DriftJson.decodeFromString(core.recentSearches(limit.toLong())) }

    suspend fun clearSearches() {
        call { core.clearSearches() }
    }

    suspend fun recordImpressions(impressions: List<Impression>) {
        call { core.recordImpressions(DriftJson.encodeToString(impressions)) }
    }

    // Queue

    suspend fun queueState(): QueueState = callQueue { core.queueState() }

    suspend fun queuePlay(tracks: List<Track>, kind: ItemKind, startIndex: Int, playContext: String): QueueState =
        callQueue { core.queuePlay(DriftJson.encodeToString(tracks), kind.wireName, startIndex.toLong(), playContext) }

    suspend fun queuePlayNext(tracks: List<Track>): QueueState =
        callQueue { core.queuePlayNext(DriftJson.encodeToString(tracks)) }

    suspend fun queueAdd(tracks: List<Track>): QueueState = callQueue { core.queueAdd(DriftJson.encodeToString(tracks)) }

    suspend fun queueRemove(entryId: Long): QueueState = callQueue { core.queueRemove(entryId) }

    suspend fun queueMove(entryId: Long, newIndex: Int): QueueState = callQueue { core.queueMove(entryId, newIndex.toLong()) }

    suspend fun queueJump(entryId: Long): QueueState = callQueue { core.queueJump(entryId) }

    suspend fun queueClear(): QueueState = callQueue { core.queueClear() }

    suspend fun queueNext(): QueueStep = callStep { core.queueNext() }

    suspend fun queueAdvance(): QueueStep = callStep { core.queueAdvance() }

    suspend fun queuePrevious(): QueueState = callQueue { core.queuePrevious() }

    suspend fun queueSetShuffle(enabled: Boolean): QueueState = callQueue { core.queueSetShuffle(enabled) }

    suspend fun queueSetRepeat(mode: RepeatMode): QueueState = callQueue { core.queueSetRepeat(mode.wireName) }

    suspend fun queueSetAutoplay(enabled: Boolean): QueueState = callQueue { core.queueSetAutoplay(enabled) }

    // Recommendations

    suspend fun recommendedRows(rowLimit: Int, perRow: Int): List<RecommendedRow> =
        call { DriftJson.decodeFromString(core.recommendedRows(rowLimit.toLong(), perRow.toLong())) }

    suspend fun recommendedFeed(limit: Int, offset: Int, refresh: Boolean): List<RecommendedItem> =
        call { DriftJson.decodeFromString(core.recommendedFeed(limit.toLong(), offset.toLong(), refresh)) }

    // Import

    /** Counts what a Spotify data export (zip, folder or JSON) or an Exportify file or folder holds. */
    suspend fun spotifyPreviewExport(path: String): ImportSummary =
        call { DriftJson.decodeFromString(core.spotifyPreviewExport(path)) }

    /** Counts what a public open.spotify.com playlist or album link holds. */
    suspend fun spotifyPreviewUrl(link: String): ImportSummary =
        call { DriftJson.decodeFromString(core.spotifyPreviewURL(link)) }

    /**
     * Starts a Spotify import in the background; [listener] hears its progress and its end. Throws
     * when the import cannot start, e.g. because another one is running.
     */
    suspend fun spotifyImport(
        sourceKind: ImportSourceKind,
        pathOrUrl: String,
        options: ImportOptions,
        listener: ImportEventListener,
    ) {
        call {
            core.spotifyImport(sourceKind.wireName, pathOrUrl, DriftJson.encodeToString(options), importBridge(listener))
        }
    }

    /** Stops the running import; without one it does nothing. */
    suspend fun cancelImport() {
        call { core.spotifyCancelImport() }
    }

    /** Validates a music.youtube.com Cookie header and stores it for library import. */
    suspend fun youTubeSignIn(cookieHeader: String): YouTubeAccountStatus =
        call { DriftJson.decodeFromString(core.youTubeSignIn(cookieHeader)) }

    /** The stored sign-in; makes no network request. */
    suspend fun youTubeAccount(): YouTubeAccountStatus = call { DriftJson.decodeFromString(core.youTubeAccount()) }

    suspend fun youTubeSignOut() {
        call { core.youTubeSignOut() }
    }

    /** Counts the signed-in account's library. */
    suspend fun youTubePreview(): ImportSummary = call { DriftJson.decodeFromString(core.youTubePreview()) }

    /** Imports the account's library in the background; run it again to re-sync. */
    suspend fun youTubeImport(options: ImportOptions, listener: ImportEventListener) {
        call { core.youTubeImport(DriftJson.encodeToString(options), importBridge(listener)) }
    }

    private fun importBridge(listener: ImportEventListener): ImportListener =
        object : ImportListener {
            override fun onImportProgress(progressJson: String) {
                listener.onProgress(DriftJson.decodeFromString<ImportProgress>(progressJson))
            }

            override fun onImportDone(resultJson: String, errorMessage: String) {
                if (errorMessage.isNotEmpty()) {
                    listener.onFailed(errorMessage)
                    return
                }
                listener.onFinished(DriftJson.decodeFromString<ImportResult>(resultJson))
            }
        }

    // Artwork

    suspend fun coverColors(url: String): CoverColors = call { DriftJson.decodeFromString(core.coverColors(url)) }
}

/** Receives the events of a running import, on a core thread. */
interface ImportEventListener {
    fun onProgress(progress: ImportProgress)

    fun onFinished(result: ImportResult)

    /** [message] is [IMPORT_CANCELLED_MESSAGE] when the user cancelled. */
    fun onFailed(message: String)
}

/** What the core reports when an import was cancelled. */
const val IMPORT_CANCELLED_MESSAGE = "cancelled"
