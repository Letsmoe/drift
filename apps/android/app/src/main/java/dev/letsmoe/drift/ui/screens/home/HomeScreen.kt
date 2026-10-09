package dev.letsmoe.drift.ui.screens.home

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.RecommendedItem
import dev.letsmoe.drift.core.model.RecommendedRow
import dev.letsmoe.drift.core.model.Shelf
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalBottomOverlayPadding
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.ErrorState
import dev.letsmoe.drift.ui.components.LoadingMoreFooter
import dev.letsmoe.drift.ui.components.MediaCard
import dev.letsmoe.drift.ui.components.OnNearEnd
import dev.letsmoe.drift.ui.components.DriftChip
import dev.letsmoe.drift.ui.components.DriftPullToRefresh
import dev.letsmoe.drift.ui.components.RootTopBar
import dev.letsmoe.drift.ui.components.SectionHeader
import dev.letsmoe.drift.ui.components.ShelfRow
import dev.letsmoe.drift.ui.components.Skeleton
import dev.letsmoe.drift.ui.components.SongGrid
import dev.letsmoe.drift.ui.components.rememberItemOpener
import dev.letsmoe.drift.ui.screens.common.LoadState
import dev.letsmoe.drift.ui.theme.DriftTheme
import java.time.LocalTime

private const val FEED_CONTEXT = "recommend:feed"
private const val HOME_CONTEXT = "home"

@Composable
fun HomeScreen(modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> HomeViewModel(container.core, container.impressionLogger) }
    val feed by viewModel.feed.collectAsStateWithLifecycle()
    val rows by viewModel.rows.collectAsStateWithLifecycle()
    val shelvesState by viewModel.shelvesState.collectAsStateWithLifecycle()
    val shelves by viewModel.shelfPager.entries.collectAsStateWithLifecycle()
    val hasMoreShelves by viewModel.shelfPager.hasMore.collectAsStateWithLifecycle()
    val isLoadingMoreShelves by viewModel.shelfPager.isLoadingMore.collectAsStateWithLifecycle()
    val isRefreshing by viewModel.isRefreshing.collectAsStateWithLifecycle()
    val filter by viewModel.filter.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    listState.OnNearEnd(isEnabled = hasMoreShelves, onNearEnd = viewModel.shelfPager::loadMore)

    Column(modifier = modifier.fillMaxSize()) {
        RootTopBar(title = greeting())
        DriftPullToRefresh(isRefreshing = isRefreshing, onRefresh = viewModel::refresh) {
            LazyColumn(
                state = listState,
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(bottom = LocalBottomOverlayPadding.current),
                verticalArrangement = Arrangement.spacedBy(24.dp),
            ) {
                item { FilterBar(selected = filter, onSelect = viewModel::selectFilter) }
                if (filter != HomeFilter.Podcasts) {
                    recommendationSections(viewModel, feed, rows)
                }
                shelfSections(viewModel, shelvesState, shelves, filter)
                if (isLoadingMoreShelves) {
                    item { LoadingMoreFooter() }
                }
            }
        }
    }
}

private fun greeting(): String {
    val hour = LocalTime.now().hour
    return when {
        hour < 5 -> "Good night"
        hour < 12 -> "Good morning"
        hour < 18 -> "Good afternoon"
        else -> "Good evening"
    }
}

@Composable
private fun FilterBar(selected: HomeFilter, onSelect: (HomeFilter) -> Unit) {
    Row(
        modifier = Modifier.padding(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        for (option in HomeFilter.entries) {
            DriftChip(label = option.label, isSelected = option == selected, onClick = { onSelect(option) })
        }
    }
}

private fun LazyListScope.recommendationSections(
    viewModel: HomeViewModel,
    feed: LoadState<List<RecommendedItem>>,
    rows: LoadState<List<RecommendedRow>>,
) {
    when (feed) {
        LoadState.Loading -> item(key = "feed-loading") { SongGridSkeleton(title = "Made for you") }
        is LoadState.Error -> item(key = "feed-error") { SectionNote("Recommendations are unavailable: ${feed.message}") }
        is LoadState.Content -> if (feed.data.isNotEmpty()) {
            item(key = "feed") { RecommendedGrid(title = "Made for you", items = feed.data, viewModel = viewModel) }
        }
    }
    when (rows) {
        LoadState.Loading -> item(key = "rows-loading") { CardRowSkeleton() }
        is LoadState.Error -> Unit
        is LoadState.Content -> for (row in rows.data) {
            item(key = "row-${row.key}") { RecommendedCardRow(row = row, viewModel = viewModel) }
        }
    }
}

@Composable
private fun RecommendedGrid(title: String, items: List<RecommendedItem>, viewModel: HomeViewModel) {
    val playerController = LocalAppContainer.current.playerController
    val tracks = items.map { item -> item.track }
    Column {
        SectionHeader(title = title)
        SongGrid(
            tracks = tracks,
            onSongClick = { index, track ->
                viewModel.click(items[index], index)
                playerController.playOrToggle(track, tracks, index, FEED_CONTEXT)
            },
            onSongVisible = { index, _ -> viewModel.impress(items[index], index) },
        )
    }
}

@Composable
private fun RecommendedCardRow(row: RecommendedRow, viewModel: HomeViewModel) {
    val playerController = LocalAppContainer.current.playerController
    val tracks = row.items.map { item -> item.track }
    val context = "recommend:${row.key}"
    ShelfRow(title = row.title) {
        itemsIndexed(row.items, key = { _, item -> item.track.videoId }) { index, item ->
            LaunchedEffect(item.track.videoId) { viewModel.impress(item, index) }
            MediaCard(
                title = item.track.title,
                subtitle = item.track.artistNames,
                thumbnails = item.track.thumbnails,
                onClick = {
                    viewModel.click(item, index)
                    playerController.playOrToggle(item.track, tracks, index, context)
                },
            )
        }
    }
}

private fun LazyListScope.shelfSections(
    viewModel: HomeViewModel,
    shelvesState: LoadState<Unit>,
    shelves: List<Shelf>,
    filter: HomeFilter,
) {
    when (shelvesState) {
        LoadState.Loading -> {
            item(key = "shelves-loading") { SongGridSkeleton(title = "") }
            item(key = "shelves-loading-cards") { CardRowSkeleton() }
        }
        is LoadState.Error -> item(key = "shelves-error") {
            Box(modifier = Modifier.fillMaxWidth().height(280.dp)) {
                ErrorState(message = shelvesState.message, onRetry = viewModel::retryShelves)
            }
        }
        is LoadState.Content -> {
            val visibleShelves = shelves.filter { shelf -> isVisible(shelf, filter) }
            itemsOf(visibleShelves) { shelf -> HomeShelf(shelf) }
            if (visibleShelves.isEmpty()) {
                item(key = "shelves-empty") { SectionNote("Nothing to show here yet.") }
            }
        }
    }
}

private fun isVisible(shelf: Shelf, filter: HomeFilter): Boolean =
    when (filter) {
        HomeFilter.All -> true
        HomeFilter.Music -> !shelf.isPodcastShelf()
        HomeFilter.Podcasts -> shelf.isPodcastShelf()
    }

private fun LazyListScope.itemsOf(shelves: List<Shelf>, content: @Composable (Shelf) -> Unit) {
    shelves.forEachIndexed { index, shelf ->
        item(key = "shelf-$index-${shelf.title}") { content(shelf) }
    }
}

// Song-only shelves ("Quick picks") render as the paged song grid, the rest as card rows.
@Composable
private fun HomeShelf(shelf: Shelf) {
    val playerController = LocalAppContainer.current.playerController
    val tracks = shelf.songTracks()
    if (tracks == null) {
        ShelfRow(shelf = shelf, onItemClick = rememberItemOpener(HOME_CONTEXT))
        return
    }
    Column {
        SectionHeader(title = shelf.title)
        SongGrid(
            tracks = tracks,
            onSongClick = { index, track -> playerController.playOrToggle(track, tracks, index, HOME_CONTEXT) },
        )
    }
}

@Composable
private fun SectionNote(message: String) {
    Text(
        text = message,
        style = MaterialTheme.typography.bodyMedium,
        color = DriftTheme.colors.textDim,
        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 16.dp),
    )
}

@Composable
private fun SongGridSkeleton(title: String) {
    Column(modifier = Modifier.padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        if (title.isNotEmpty()) {
            Text(text = title, style = MaterialTheme.typography.titleLarge)
        }
        repeat(4) {
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                Skeleton(modifier = Modifier.size(52.dp))
                Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Skeleton(modifier = Modifier.fillMaxWidth(0.6f).height(14.dp))
                    Skeleton(modifier = Modifier.fillMaxWidth(0.35f).height(12.dp))
                }
            }
        }
    }
}

@Composable
private fun CardRowSkeleton() {
    Column(modifier = Modifier.padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Skeleton(modifier = Modifier.fillMaxWidth(0.4f).height(22.dp))
        LazyRow(horizontalArrangement = Arrangement.spacedBy(12.dp), userScrollEnabled = false) {
            itemsIndexed(listOf(0, 1, 2)) { _, _ ->
                Skeleton(modifier = Modifier.size(160.dp))
            }
        }
    }
}
