package dev.neoworks.ottermusic.ui.screens.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.RecommendedItem
import dev.neoworks.ottermusic.core.model.RecommendedRow
import dev.neoworks.ottermusic.core.model.Shelf
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.library.ImpressionLogger
import dev.neoworks.ottermusic.ui.describeError
import dev.neoworks.ottermusic.ui.screens.common.ContinuationPage
import dev.neoworks.ottermusic.ui.screens.common.ContinuationPager
import dev.neoworks.ottermusic.ui.screens.common.LoadState
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.joinAll
import kotlinx.coroutines.launch

enum class HomeFilter(val label: String) {
    All("All"),
    Music("Music"),
    Podcasts("Podcasts"),
}

private const val RECOMMENDATION_ROW_LIMIT = 6
private const val ITEMS_PER_RECOMMENDATION_ROW = 12
private const val FEED_SIZE = 20
const val HOME_SURFACE = "home"

/**
 * Home: the "Made for you" feed and reason rows from the recommendation engine, plus YouTube's
 * home shelves with infinite continuation. Each part loads and fails on its own.
 */
class HomeViewModel(
    private val core: OtterCore,
    private val impressionLogger: ImpressionLogger,
) : ViewModel() {
    private val mutableFeed = MutableStateFlow<LoadState<List<RecommendedItem>>>(LoadState.Loading)
    val feed: StateFlow<LoadState<List<RecommendedItem>>> = mutableFeed.asStateFlow()

    private val mutableRows = MutableStateFlow<LoadState<List<RecommendedRow>>>(LoadState.Loading)
    val rows: StateFlow<LoadState<List<RecommendedRow>>> = mutableRows.asStateFlow()

    private val mutableShelvesState = MutableStateFlow<LoadState<Unit>>(LoadState.Loading)
    val shelvesState: StateFlow<LoadState<Unit>> = mutableShelvesState.asStateFlow()

    private val mutableIsRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = mutableIsRefreshing.asStateFlow()

    private val mutableFilter = MutableStateFlow(HomeFilter.All)
    val filter: StateFlow<HomeFilter> = mutableFilter.asStateFlow()

    /** Home shelves; later pages are fetched as the user scrolls. */
    val shelfPager = ContinuationPager<Shelf>(viewModelScope) { token -> fetchMoreShelves(token) }

    private val impressedVideoIds = mutableSetOf<String>()

    init {
        loadEverything(refreshFeed = false)
    }

    fun selectFilter(selected: HomeFilter) {
        mutableFilter.value = selected
    }

    /** Pull to refresh: re-ranks the feed and reloads rows and shelves, keeping the old content until the new arrives. */
    fun refresh() {
        if (mutableIsRefreshing.value) {
            return
        }
        mutableIsRefreshing.value = true
        impressedVideoIds.clear()
        viewModelScope.launch {
            loadEverything(refreshFeed = true).joinAll()
            mutableIsRefreshing.value = false
        }
    }

    fun retryShelves() {
        mutableShelvesState.value = LoadState.Loading
        viewModelScope.launch { loadShelves() }
    }

    fun retryRecommendations() {
        mutableFeed.value = LoadState.Loading
        mutableRows.value = LoadState.Loading
        viewModelScope.launch { loadFeed(refresh = false) }
        viewModelScope.launch { loadRows() }
    }

    /** Logs that a recommended item was shown, once per video until the next refresh. */
    fun impress(item: RecommendedItem, position: Int) {
        if (!impressedVideoIds.add(item.track.videoId)) {
            return
        }
        impressionLogger.logImpression(item.track.videoId, item.source, HOME_SURFACE, position)
    }

    fun click(item: RecommendedItem, position: Int) {
        impressionLogger.logClick(item.track.videoId, item.source, HOME_SURFACE, position)
    }

    private fun loadEverything(refreshFeed: Boolean): List<Job> =
        listOf(
            viewModelScope.launch { loadFeed(refreshFeed) },
            viewModelScope.launch { loadRows() },
            viewModelScope.launch { loadShelves() },
        )

    private suspend fun loadFeed(refresh: Boolean) {
        mutableFeed.value = attempt { core.recommendedFeed(FEED_SIZE, 0, refresh) }
    }

    private suspend fun loadRows() {
        mutableRows.value = attempt { core.recommendedRows(RECOMMENDATION_ROW_LIMIT, ITEMS_PER_RECOMMENDATION_ROW) }
    }

    private suspend fun loadShelves() {
        val result = attempt { core.home() }
        if (result is LoadState.Error) {
            // A failed refresh keeps the shelves already on screen.
            if (shelfPager.entries.value.isEmpty()) {
                mutableShelvesState.value = result
            }
            return
        }
        if (result is LoadState.Content) {
            shelfPager.reset(result.data.shelves, result.data.continuation)
            mutableShelvesState.value = LoadState.Content(Unit)
        }
    }

    private suspend fun fetchMoreShelves(token: String): ContinuationPage<Shelf> {
        val next = core.homeContinuation(token)
        return ContinuationPage(next.shelves, next.continuation)
    }

    private suspend fun <Data> attempt(block: suspend () -> Data): LoadState<Data> {
        try {
            return LoadState.Content(block())
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            return LoadState.Error(describeError(error))
        }
    }
}

/** True for shelves made only of podcasts and episodes. */
fun Shelf.isPodcastShelf(): Boolean = items.isNotEmpty() && items.all { item -> item.isPodcastItem() }

private fun Item.isPodcastItem(): Boolean = kind == ItemKind.Podcast || kind == ItemKind.Episode

/** The tracks of a shelf that holds only songs, videos or episodes, or null for any other shelf. */
fun Shelf.songTracks(): List<Track>? {
    val tracks = items.mapNotNull { item -> item.track }
    if (tracks.size != items.size || tracks.isEmpty()) {
        return null
    }
    return tracks
}
