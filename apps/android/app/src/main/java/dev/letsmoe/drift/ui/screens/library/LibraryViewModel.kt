package dev.letsmoe.drift.ui.screens.library

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.library.LibraryState
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class LibraryViewModel(
    private val core: DriftCore,
    val libraryState: LibraryState,
) : ViewModel() {
    private val mutableFilter = MutableStateFlow<LibraryFilter?>(null)
    val filter: StateFlow<LibraryFilter?> = mutableFilter.asStateFlow()

    private val mutableSort = MutableStateFlow(LibrarySort.Recents)
    val sort: StateFlow<LibrarySort> = mutableSort.asStateFlow()

    private val mutableLikedCount = MutableStateFlow(0)
    val likedCount: StateFlow<Int> = mutableLikedCount.asStateFlow()

    val entries: StateFlow<List<LibraryEntry>> =
        combine(libraryState.playlists, libraryState.savedItems, filter, sort) { playlists, savedItems, filter, sort ->
            val allEntries = playlists.map { summary -> LibraryEntry.LocalPlaylist(summary) } +
                savedItems.map { saved -> LibraryEntry.Saved(saved) }
            arrange(allEntries, filter, sort)
        }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(STOP_TIMEOUT_MS), emptyList())

    init {
        libraryState.refreshPlaylists()
        libraryState.refreshSavedItems()
        viewModelScope.launch {
            // The like count follows every like and unlike made anywhere in the app.
            libraryState.likedVideoIds.collect { refreshLikedCount() }
        }
    }

    /** Selecting the active chip clears the filter. */
    fun toggleFilter(selected: LibraryFilter) {
        if (mutableFilter.value == selected) {
            mutableFilter.value = null
            return
        }
        mutableFilter.value = selected
    }

    fun toggleSort() {
        if (mutableSort.value == LibrarySort.Recents) {
            mutableSort.value = LibrarySort.Alphabetical
            return
        }
        mutableSort.value = LibrarySort.Recents
    }

    private suspend fun refreshLikedCount() {
        try {
            mutableLikedCount.value = core.likedTracks().size
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            mutableLikedCount.value = 0
        }
    }

    private fun arrange(entries: List<LibraryEntry>, filter: LibraryFilter?, sort: LibrarySort): List<LibraryEntry> {
        var visible = entries
        if (filter != null) {
            visible = entries.filter { entry -> entry.filter == filter }
        }
        if (sort == LibrarySort.Alphabetical) {
            return visible.sortedBy { entry -> entry.title.lowercase() }
        }
        return visible.sortedByDescending { entry -> entry.recencyTimestamp }
    }

    private companion object {
        const val STOP_TIMEOUT_MS = 5_000L
    }
}
