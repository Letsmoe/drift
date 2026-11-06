package dev.letsmoe.drift.ui.screens.podcasts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.Episode
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.Shelf
import dev.letsmoe.drift.library.LibraryState
import dev.letsmoe.drift.ui.describeError
import dev.letsmoe.drift.ui.screens.common.ContinuationPage
import dev.letsmoe.drift.ui.screens.common.ContinuationPager
import dev.letsmoe.drift.ui.screens.common.LoadState
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

// Opening every saved show's page is a request each; the newest saved ones are enough.
private const val LATEST_EPISODE_SHOW_LIMIT = 8

/**
 * The Podcasts tab: the user's saved shows, the newest episode of each, then YouTube's podcast
 * shows by topic with infinite continuation.
 */
class PodcastsViewModel(
    private val core: DriftCore,
    libraryState: LibraryState,
) : ViewModel() {
    val savedShows: StateFlow<List<Item>> = libraryState.savedItems
        .map { savedItems -> savedItems.map { saved -> saved.item }.filter { item -> item.kind == ItemKind.Podcast } }
        .distinctUntilChanged()
        .stateIn(viewModelScope, SharingStarted.Eagerly, emptyList())

    private val mutableLatestEpisodes = MutableStateFlow<List<Episode>>(emptyList())
    val latestEpisodes: StateFlow<List<Episode>> = mutableLatestEpisodes.asStateFlow()

    private val mutableShelvesState = MutableStateFlow<LoadState<Unit>>(LoadState.Loading)
    val shelvesState: StateFlow<LoadState<Unit>> = mutableShelvesState.asStateFlow()

    /** Topic shelves of shows; later pages are fetched as the user scrolls. */
    val shelfPager = ContinuationPager<Shelf>(viewModelScope) { token -> fetchMoreShelves(token) }

    init {
        viewModelScope.launch { loadShelves() }
        viewModelScope.launch { savedShows.collectLatest { shows -> loadLatestEpisodes(shows) } }
    }

    fun retryShelves() {
        mutableShelvesState.value = LoadState.Loading
        viewModelScope.launch { loadShelves() }
    }

    private suspend fun loadShelves() {
        try {
            val page = core.podcastShows()
            shelfPager.reset(page.shelves, page.continuation)
            mutableShelvesState.value = LoadState.Content(Unit)
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            mutableShelvesState.value = LoadState.Error(describeError(error))
        }
    }

    private suspend fun fetchMoreShelves(token: String): ContinuationPage<Shelf> {
        val next = core.podcastShowsContinuation(token)
        return ContinuationPage(next.shelves, next.continuation)
    }

    // Show pages list their newest episode first; a show that fails to load is left out.
    private suspend fun loadLatestEpisodes(shows: List<Item>) {
        val episodes = coroutineScope {
            shows.take(LATEST_EPISODE_SHOW_LIMIT).map { show -> async { newestEpisodeOf(show.id) } }.awaitAll()
        }
        mutableLatestEpisodes.value = episodes.filterNotNull()
    }

    private suspend fun newestEpisodeOf(showId: String): Episode? {
        try {
            return core.podcast(showId).episodes.firstOrNull()
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            return null
        }
    }
}
