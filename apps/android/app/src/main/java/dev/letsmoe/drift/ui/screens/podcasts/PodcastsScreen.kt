package dev.letsmoe.drift.ui.screens.podcasts

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
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.Episode
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.Shelf
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalBottomOverlayPadding
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.ErrorState
import dev.letsmoe.drift.ui.components.LoadingMoreFooter
import dev.letsmoe.drift.ui.components.MediaCard
import dev.letsmoe.drift.ui.components.OnNearEnd
import dev.letsmoe.drift.ui.components.RootTopBar
import dev.letsmoe.drift.ui.components.SectionHeader
import dev.letsmoe.drift.ui.components.ShelfRow
import dev.letsmoe.drift.ui.components.Skeleton
import dev.letsmoe.drift.ui.components.rememberItemOpener
import dev.letsmoe.drift.ui.navigation.openPodcast
import dev.letsmoe.drift.ui.screens.common.LoadState
import dev.letsmoe.drift.ui.screens.podcast.EpisodeRow

private const val PODCASTS_CONTEXT = "podcasts"
private const val LATEST_EPISODES_CONTEXT = "podcasts:latest"

@Composable
fun PodcastsScreen(modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> PodcastsViewModel(container.core, container.libraryState) }
    val savedShows by viewModel.savedShows.collectAsStateWithLifecycle()
    val latestEpisodes by viewModel.latestEpisodes.collectAsStateWithLifecycle()
    val shelvesState by viewModel.shelvesState.collectAsStateWithLifecycle()
    val shelves by viewModel.shelfPager.entries.collectAsStateWithLifecycle()
    val hasMoreShelves by viewModel.shelfPager.hasMore.collectAsStateWithLifecycle()
    val isLoadingMoreShelves by viewModel.shelfPager.isLoadingMore.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    listState.OnNearEnd(isEnabled = hasMoreShelves, onNearEnd = viewModel.shelfPager::loadMore)

    Column(modifier = modifier.fillMaxSize()) {
        RootTopBar(title = "Podcasts")
        LazyColumn(
            state = listState,
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(bottom = LocalBottomOverlayPadding.current),
            verticalArrangement = Arrangement.spacedBy(24.dp),
        ) {
            if (savedShows.isNotEmpty()) {
                item(key = "saved-shows") { SavedShowsRow(savedShows) }
            }
            if (latestEpisodes.isNotEmpty()) {
                item(key = "latest-episodes") { LatestEpisodes(latestEpisodes) }
            }
            shelfSections(shelvesState, shelves, onRetry = viewModel::retryShelves)
            if (isLoadingMoreShelves) {
                item { LoadingMoreFooter() }
            }
        }
    }
}

@Composable
private fun SavedShowsRow(shows: List<Item>) {
    val navController = LocalNavController.current
    ShelfRow(title = "Your shows") {
        items(shows, key = { show -> show.id }) { show ->
            MediaCard(item = show, onClick = { navController.openPodcast(show.id) })
        }
    }
}

@Composable
private fun LatestEpisodes(episodes: List<Episode>) {
    val playerController = LocalAppContainer.current.playerController
    val tracks = episodes.map { episode -> episode.track }
    Column {
        SectionHeader(title = "Latest episodes")
        episodes.forEachIndexed { index, episode ->
            EpisodeRow(
                episode = episode,
                onPlay = {
                    playerController.playOrToggle(episode.track, tracks, index, LATEST_EPISODES_CONTEXT, ItemKind.Episode)
                },
            )
        }
    }
}

private fun LazyListScope.shelfSections(shelvesState: LoadState<Unit>, shelves: List<Shelf>, onRetry: () -> Unit) {
    when (shelvesState) {
        LoadState.Loading -> items(2) { ShelfSkeleton() }
        is LoadState.Error -> item(key = "shelves-error") {
            Box(modifier = Modifier.fillMaxWidth().height(280.dp)) {
                ErrorState(message = shelvesState.message, onRetry = onRetry)
            }
        }
        is LoadState.Content -> itemsIndexed(shelves, key = { index, shelf -> "shelf-$index-${shelf.title}" }) { _, shelf ->
            ShelfRow(shelf = shelf, onItemClick = rememberItemOpener(PODCASTS_CONTEXT))
        }
    }
}

@Composable
private fun ShelfSkeleton() {
    Column(modifier = Modifier.padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Skeleton(modifier = Modifier.fillMaxWidth(0.4f).height(22.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            repeat(3) { Skeleton(modifier = Modifier.size(160.dp)) }
        }
    }
}
