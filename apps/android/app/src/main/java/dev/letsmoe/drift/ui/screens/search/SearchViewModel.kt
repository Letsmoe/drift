package dev.letsmoe.drift.ui.screens.search

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.Shelf
import dev.letsmoe.drift.ui.describeError
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed interface SearchState {
    data object Idle : SearchState

    data object Loading : SearchState

    data class Failed(val message: String) : SearchState

    data class Done(val shelves: List<Shelf>) : SearchState
}

private const val SUGGESTION_DEBOUNCE_MS = 250L
private const val RECENT_SEARCH_LIMIT = 10
private const val TOP_RESULT_TITLE = "Top result"
private const val SONGS_TITLE = "Songs"

/**
 * Search: the field text, debounced suggestions, recent searches and the results of the last
 * submitted query. Lives with the navigation entry, so it all survives tab switches.
 */
class SearchViewModel(private val core: DriftCore) : ViewModel() {
    private val mutableQuery = MutableStateFlow("")
    val query: StateFlow<String> = mutableQuery.asStateFlow()

    /** The query the current [state] belongs to; differs from [query] while the user edits. */
    private val mutableSubmittedQuery = MutableStateFlow("")
    val submittedQuery: StateFlow<String> = mutableSubmittedQuery.asStateFlow()

    private val mutableState = MutableStateFlow<SearchState>(SearchState.Idle)
    val state: StateFlow<SearchState> = mutableState.asStateFlow()

    private val mutableSuggestions = MutableStateFlow<List<String>>(emptyList())
    val suggestions: StateFlow<List<String>> = mutableSuggestions.asStateFlow()

    private val mutableRecentQueries = MutableStateFlow<List<String>>(emptyList())
    val recentQueries: StateFlow<List<String>> = mutableRecentQueries.asStateFlow()

    /** Title of the shelf the results are narrowed to; empty shows everything. */
    private val mutableFilterTitle = MutableStateFlow("")
    val filterTitle: StateFlow<String> = mutableFilterTitle.asStateFlow()

    private var suggestionJob: Job? = null
    private var searchJob: Job? = null

    init {
        loadRecentQueries()
    }

    fun onQueryChange(text: String) {
        mutableQuery.value = text
        suggestionJob?.cancel()
        if (text.isBlank()) {
            clear()
            return
        }
        suggestionJob = viewModelScope.launch {
            delay(SUGGESTION_DEBOUNCE_MS)
            loadSuggestions(text.trim())
        }
    }

    /** Puts a suggestion into the field without searching, so it can be extended. */
    fun fillIntoField(text: String) {
        onQueryChange(text)
    }

    fun submit(text: String) {
        val trimmed = text.trim()
        if (trimmed.isEmpty()) {
            return
        }
        suggestionJob?.cancel()
        mutableSuggestions.value = emptyList()
        mutableQuery.value = trimmed
        mutableSubmittedQuery.value = trimmed
        mutableFilterTitle.value = ""
        mutableState.value = SearchState.Loading
        searchJob?.cancel()
        searchJob = viewModelScope.launch { runSearch(trimmed) }
    }

    fun retry() {
        submit(mutableSubmittedQuery.value)
    }

    /** Back to the recent searches. */
    fun clear() {
        suggestionJob?.cancel()
        searchJob?.cancel()
        mutableQuery.value = ""
        mutableSubmittedQuery.value = ""
        mutableSuggestions.value = emptyList()
        mutableState.value = SearchState.Idle
        mutableFilterTitle.value = ""
        loadRecentQueries()
    }

    fun selectFilter(title: String) {
        mutableFilterTitle.value = title
    }

    fun clearRecentQueries() {
        viewModelScope.launch {
            try {
                core.clearSearches()
                mutableRecentQueries.value = emptyList()
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (error: Exception) {
                mutableRecentQueries.value = emptyList()
            }
        }
    }

    // Suggestions are a convenience; the field keeps working without them.
    private suspend fun loadSuggestions(text: String) {
        try {
            mutableSuggestions.value = core.searchSuggestions(text)
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            mutableSuggestions.value = emptyList()
        }
    }

    private suspend fun runSearch(text: String) {
        try {
            mutableState.value = SearchState.Done(core.search(text).shelves)
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            mutableState.value = SearchState.Failed(describeError(error))
        }
    }

    private fun loadRecentQueries() {
        viewModelScope.launch {
            try {
                mutableRecentQueries.value = core.recentSearches(RECENT_SEARCH_LIMIT)
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (error: Exception) {
                mutableRecentQueries.value = emptyList()
            }
        }
    }
}

/** The first top result, or the first item of the first shelf. */
fun List<Shelf>.topResult(): Item? {
    val topShelf = firstOrNull { shelf -> shelf.title == TOP_RESULT_TITLE }
    if (topShelf != null) {
        return topShelf.items.firstOrNull()
    }
    return firstOrNull()?.items?.firstOrNull()
}

/** Shelf titles offered as filter chips (everything except the top result). */
fun List<Shelf>.filterTitles(): List<String> =
    map { shelf -> shelf.title }.filter { title -> title != TOP_RESULT_TITLE && title.isNotEmpty() }

/** The shelves shown as lists below the top result. */
fun List<Shelf>.listShelves(): List<Shelf> = filter { shelf -> shelf.title != TOP_RESULT_TITLE && shelf.items.isNotEmpty() }

fun isSongsShelf(shelf: Shelf): Boolean = shelf.title == SONGS_TITLE
