package dev.letsmoe.drift.ui.screens.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.RecommendedItem
import dev.letsmoe.drift.core.model.RecommendedRow
import dev.letsmoe.drift.core.model.Shelf
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalBottomOverlayPadding
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.navigation.openSearch
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.components.DriftPullToRefresh
import dev.letsmoe.drift.ui.components.ErrorState
import dev.letsmoe.drift.ui.components.LoadingMoreFooter
import dev.letsmoe.drift.ui.components.MediaCard
import dev.letsmoe.drift.ui.components.MediaCardSize
import dev.letsmoe.drift.ui.components.OnNearEnd
import dev.letsmoe.drift.ui.components.SearchAction
import dev.letsmoe.drift.ui.components.SectionHeader
import dev.letsmoe.drift.ui.components.ShelfRow
import dev.letsmoe.drift.ui.components.Skeleton
import dev.letsmoe.drift.ui.components.SongGrid
import dev.letsmoe.drift.ui.components.rememberItemOpener
import dev.letsmoe.drift.ui.screens.common.LoadState
import dev.letsmoe.drift.ui.theme.DriftTheme

private const val FEED_CONTEXT = "recommend:feed"
private const val HOME_CONTEXT = "home"
private const val RECENT_CONTEXT = "home:recent"
private const val RECENT_COLUMNS = 2
private val FeaturedCardSize = 220.dp
private val RecentTileHeight = 56.dp
private val RecentTileShape = RoundedCornerShape(8.dp)

/**
 * Home, top to bottom: a grid of recently played songs, "Made for you" as quick-pick columns,
 * the first reason row as large cards and the others as regular ones, then YouTube's shelves
 * without the songs already shown. No greeting: the page starts with music.
 */
@Composable
fun HomeScreen(modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> HomeViewModel(container.core, container.impressionLogger) }
    val feed by viewModel.feed.collectAsStateWithLifecycle()
    val rows by viewModel.rows.collectAsStateWithLifecycle()
    val recentTracks by viewModel.recentTracks.collectAsStateWithLifecycle()
    val shelvesState by viewModel.shelvesState.collectAsStateWithLifecycle()
    val shelves by viewModel.shelfPager.entries.collectAsStateWithLifecycle()
    val hasMoreShelves by viewModel.shelfPager.hasMore.collectAsStateWithLifecycle()
    val isLoadingMoreShelves by viewModel.shelfPager.isLoadingMore.collectAsStateWithLifecycle()
    val isRefreshing by viewModel.isRefreshing.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    val navController = LocalNavController.current
    listState.OnNearEnd(isEnabled = hasMoreShelves, onNearEnd = viewModel.shelfPager::loadMore)
    val shownVideoIds = remember(recentTracks, feed) { shownVideoIdsOf(recentTracks, feed) }
    val homeShelves = remember(shelves, shownVideoIds) { shelves.forHome(shownVideoIds) }

    Column(modifier = modifier.fillMaxSize().statusBarsPadding()) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 4.dp),
            horizontalArrangement = Arrangement.End,
        ) {
            SearchAction(onClick = { navController.openSearch() })
        }
        DriftPullToRefresh(isRefreshing = isRefreshing, onRefresh = viewModel::refresh) {
            LazyColumn(
                state = listState,
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(top = 4.dp, bottom = LocalBottomOverlayPadding.current),
                verticalArrangement = Arrangement.spacedBy(28.dp),
            ) {
                if (recentTracks.size >= RECENT_COLUMNS) {
                    item(key = "recent") { RecentGrid(recentTracks) }
                }
                recommendationSections(viewModel, feed, rows)
                shelfSections(viewModel, shelvesState, homeShelves)
                if (isLoadingMoreShelves) {
                    item { LoadingMoreFooter() }
                }
            }
        }
    }
}

private fun shownVideoIdsOf(recentTracks: List<Track>, feed: LoadState<List<RecommendedItem>>): Set<String> {
    val videoIds = recentTracks.map { track -> track.videoId }.toMutableSet()
    if (feed is LoadState.Content) {
        videoIds.addAll(feed.data.map { item -> item.track.videoId })
    }
    return videoIds
}

/** Recently played songs as two columns of compact tiles, the way back into what was on. */
@Composable
private fun RecentGrid(tracks: List<Track>) {
    val playerController = LocalAppContainer.current.playerController
    Column(
        modifier = Modifier.padding(horizontal = 16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        tracks.chunked(RECENT_COLUMNS).forEachIndexed { rowIndex, rowTracks ->
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                rowTracks.forEachIndexed { columnIndex, track ->
                    val index = rowIndex * RECENT_COLUMNS + columnIndex
                    RecentTile(
                        track = track,
                        onClick = { playerController.playOrToggle(track, tracks, index, RECENT_CONTEXT) },
                        modifier = Modifier.weight(1f),
                    )
                }
            }
        }
    }
}

@Composable
private fun RecentTile(track: Track, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Row(
        modifier = modifier
            .height(RecentTileHeight)
            .clip(RecentTileShape)
            .background(Color.White.copy(alpha = 0.08f))
            .clickable(onClick = onClick),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Cover(thumbnails = track.thumbnails, size = RecentTileHeight, shape = RoundedCornerShape(0.dp))
        Text(
            text = track.title,
            style = MaterialTheme.typography.labelLarge,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f).padding(end = 8.dp),
        )
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
        is LoadState.Content -> rows.data.forEachIndexed { index, row ->
            item(key = "row-${row.key}") {
                RecommendedCardRow(row = row, viewModel = viewModel, isFeatured = index == 0)
            }
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

// The first reason row is the most personal one, so it gets large cards to break up the page.
@Composable
private fun RecommendedCardRow(row: RecommendedRow, viewModel: HomeViewModel, isFeatured: Boolean) {
    val playerController = LocalAppContainer.current.playerController
    val tracks = row.items.map { item -> item.track }
    val context = "recommend:${row.key}"
    val cardSize = cardSizeFor(isFeatured)
    ShelfRow(title = row.title) {
        itemsIndexed(row.items, key = { _, item -> item.track.videoId }) { index, item ->
            LaunchedEffect(item.track.videoId) { viewModel.impress(item, index) }
            MediaCard(
                title = item.track.title,
                subtitle = item.track.artistNames,
                thumbnails = item.track.thumbnails,
                size = cardSize,
                onClick = {
                    viewModel.click(item, index)
                    playerController.playOrToggle(item.track, tracks, index, context)
                },
            )
        }
    }
}

private fun cardSizeFor(isFeatured: Boolean): Dp {
    if (isFeatured) {
        return FeaturedCardSize
    }
    return MediaCardSize
}

private fun LazyListScope.shelfSections(
    viewModel: HomeViewModel,
    shelvesState: LoadState<Unit>,
    shelves: List<Shelf>,
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
        is LoadState.Content -> shelves.forEachIndexed { index, shelf ->
            item(key = "shelf-$index-${shelf.title}") { HomeShelf(shelf) }
        }
    }
}

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
