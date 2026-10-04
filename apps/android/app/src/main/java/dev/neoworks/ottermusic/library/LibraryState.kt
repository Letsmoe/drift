package dev.neoworks.ottermusic.library

import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.LibraryPlaylist
import dev.neoworks.ottermusic.core.model.PlaylistSummary
import dev.neoworks.ottermusic.core.model.SavedItem
import dev.neoworks.ottermusic.core.model.Track
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

private const val LIKED = 1
private const val NOT_RATED = 0

/** The user's local library: ratings, playlists and saved items, kept in sync with the core. */
class LibraryState(
    private val core: OtterCore,
    private val scope: CoroutineScope,
) {
    private val mutableRatings = MutableStateFlow<Map<String, Int>>(emptyMap())
    private val requestedRatingIds = mutableSetOf<String>()

    /** Ratings looked up so far: 1 liked, -1 disliked, 0 none. */
    val ratings: StateFlow<Map<String, Int>> = mutableRatings.asStateFlow()

    val likedVideoIds: StateFlow<Set<String>> =
        ratings
            .map { ratings -> ratings.filterValues { rating -> rating == LIKED }.keys }
            .stateIn(scope, SharingStarted.Eagerly, emptySet())

    private val mutablePlaylists = MutableStateFlow<List<PlaylistSummary>>(emptyList())
    val playlists: StateFlow<List<PlaylistSummary>> = mutablePlaylists.asStateFlow()

    private val mutableSavedItems = MutableStateFlow<List<SavedItem>>(emptyList())
    val savedItems: StateFlow<List<SavedItem>> = mutableSavedItems.asStateFlow()

    private val mutableErrorMessage = MutableStateFlow("")
    val errorMessage: StateFlow<String> = mutableErrorMessage.asStateFlow()

    /** Loads playlists and saved items; call once at startup. */
    fun load() {
        refreshPlaylists()
        refreshSavedItems()
    }

    /** Loads the ratings of [videoIds] that have not been looked up yet. */
    fun ensureRatings(videoIds: List<String>) {
        val missingIds = videoIds.filter { videoId -> videoId !in requestedRatingIds }.distinct()
        if (missingIds.isEmpty()) {
            return
        }
        requestedRatingIds.addAll(missingIds)
        scope.launch { lookUpRatings(missingIds) }
    }

    private suspend fun lookUpRatings(videoIds: List<String>) {
        try {
            val found = core.ratings(videoIds)
            storeLookedUpRatings(videoIds, found)
        } catch (error: Exception) {
            requestedRatingIds.removeAll(videoIds.toSet())
            reportError(error)
        }
    }

    // A rating set while the lookup was in flight is newer than the lookup.
    private fun storeLookedUpRatings(videoIds: List<String>, found: Map<String, Int>) {
        val current = mutableRatings.value
        val merged = current.toMutableMap()
        for (videoId in videoIds) {
            if (videoId !in current) {
                merged[videoId] = found[videoId] ?: NOT_RATED
            }
        }
        mutableRatings.value = merged
    }

    fun rating(videoId: String): Int = mutableRatings.value[videoId] ?: NOT_RATED

    fun isLiked(videoId: String): Boolean = rating(videoId) == LIKED

    fun toggleLike(track: Track, kind: ItemKind = ItemKind.Song) {
        if (isLiked(track.videoId)) {
            rate(track, kind, NOT_RATED)
            return
        }
        rate(track, kind, LIKED)
    }

    // Optimistic: the heart reacts at once and rolls back if the core rejects the rating.
    private fun rate(track: Track, kind: ItemKind, newRating: Int) {
        val previousRating = rating(track.videoId)
        mutableRatings.value = mutableRatings.value + (track.videoId to newRating)
        scope.launch {
            try {
                core.setRating(track, kind, newRating)
            } catch (error: Exception) {
                mutableRatings.value = mutableRatings.value + (track.videoId to previousRating)
                reportError(error)
            }
        }
    }

    fun refreshPlaylists() {
        scope.launch {
            try {
                mutablePlaylists.value = core.playlists()
            } catch (error: Exception) {
                reportError(error)
            }
        }
    }

    suspend fun createPlaylist(name: String): LibraryPlaylist {
        val created = core.createPlaylist(name)
        refreshPlaylists()
        return created
    }

    suspend fun addToPlaylist(playlistId: String, tracks: List<Track>) {
        core.addToPlaylist(playlistId, tracks)
        refreshPlaylists()
    }

    suspend fun removeFromPlaylist(playlistId: String, entryId: Long) {
        core.removeFromPlaylist(playlistId, entryId)
        refreshPlaylists()
    }

    /** Renames a playlist, keeping its description unless a new one is given. */
    suspend fun renamePlaylist(playlistId: String, name: String, description: String? = null) {
        val newDescription = description ?: descriptionOf(playlistId)
        core.renamePlaylist(playlistId, name, newDescription)
        refreshPlaylists()
    }

    suspend fun deletePlaylist(playlistId: String) {
        core.deletePlaylist(playlistId)
        refreshPlaylists()
    }

    private fun descriptionOf(playlistId: String): String =
        mutablePlaylists.value.firstOrNull { summary -> summary.id == playlistId }?.description ?: ""

    fun refreshSavedItems() {
        scope.launch {
            try {
                mutableSavedItems.value = core.savedItems()
            } catch (error: Exception) {
                reportError(error)
            }
        }
    }

    fun isSaved(kind: ItemKind, id: String): Boolean =
        mutableSavedItems.value.any { saved -> saved.item.kind == kind && saved.item.id == id }

    suspend fun toggleSaved(item: Item) {
        if (isSaved(item.kind, item.id)) {
            core.unsave(item.kind, item.id)
        } else {
            core.save(item)
        }
        refreshSavedItems()
    }

    private fun reportError(error: Exception) {
        mutableErrorMessage.value = error.message ?: error.javaClass.simpleName
    }
}
