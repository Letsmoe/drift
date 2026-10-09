package dev.letsmoe.drift.ui.screens.playlist

import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.PlaylistPage
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.EnsureRatings
import dev.letsmoe.drift.ui.components.CollectionMoreButton
import dev.letsmoe.drift.ui.components.ExpandableText
import dev.letsmoe.drift.ui.components.LoadingMoreFooter
import dev.letsmoe.drift.ui.components.OnNearEnd
import dev.letsmoe.drift.ui.components.PlayShuffleButtons
import dev.letsmoe.drift.ui.components.SaveButton
import dev.letsmoe.drift.ui.components.TrackRow
import dev.letsmoe.drift.ui.formatSongCount
import dev.letsmoe.drift.ui.screens.detail.DetailActionRow
import dev.letsmoe.drift.ui.screens.detail.DetailHeader
import dev.letsmoe.drift.ui.screens.detail.DetailLoadState
import dev.letsmoe.drift.ui.screens.detail.DetailScaffold
import dev.letsmoe.drift.ui.screens.detail.detailListPadding
import dev.letsmoe.drift.ui.theme.DriftTheme

@Composable
fun PlaylistScreen(playlistId: String, modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> PlaylistViewModel(container.core, playlistId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    DetailLoadState(state = state, onRetry = viewModel::retry, modifier = modifier) { page ->
        PlaylistContent(playlistId = playlistId, page = page, viewModel = viewModel, modifier = modifier)
    }
}

@Composable
private fun PlaylistContent(
    playlistId: String,
    page: PlaylistPage,
    viewModel: PlaylistViewModel,
    modifier: Modifier,
) {
    val playerController = LocalAppContainer.current.playerController
    val listState = rememberLazyListState()
    val tracks by viewModel.trackPager.entries.collectAsStateWithLifecycle()
    val hasMore by viewModel.trackPager.hasMore.collectAsStateWithLifecycle()
    val isLoadingMore by viewModel.trackPager.isLoadingMore.collectAsStateWithLifecycle()
    val context = "playlist:$playlistId"
    val playlistItem = Item(
        kind = ItemKind.Playlist,
        id = playlistId,
        title = page.title,
        subtitle = page.author,
        thumbnails = page.thumbnails,
    )
    listState.OnNearEnd(isEnabled = hasMore, onNearEnd = viewModel.trackPager::loadMore)
    EnsureRatings(tracks)

    DetailScaffold(title = page.title, listState = listState, modifier = modifier) {
        LazyColumn(state = listState, contentPadding = detailListPadding()) {
            item {
                DetailHeader(title = page.title, thumbnails = page.thumbnails) {
                    PlaylistMeta(page = page, trackCount = tracks.size)
                }
            }
            item {
                DetailActionRow {
                    SaveButton(item = playlistItem)
                    CollectionMoreButton(title = page.title, tracks = tracks)
                }
            }
            item {
                PlayShuffleButtons(
                    onPlay = { playerController.playContext(tracks, context) },
                    onShuffle = { playerController.shuffleContext(tracks, context) },
                )
            }
            itemsIndexed(tracks, key = { index, track -> "$index-${track.videoId}" }) { index, track ->
                TrackRow(
                    track = track,
                    onClick = { playerController.playOrToggle(track, tracks, index, context) },
                )
            }
            if (isLoadingMore) {
                item { LoadingMoreFooter() }
            }
        }
    }
}

@Composable
private fun PlaylistMeta(page: PlaylistPage, trackCount: Int) {
    val summary = playlistSummary(page.author, trackCount)
    Text(
        text = summary,
        style = MaterialTheme.typography.bodyMedium,
        color = DriftTheme.colors.textMuted,
    )
    if (page.description.isNotEmpty()) {
        ExpandableText(text = page.description, modifier = Modifier.padding(top = 8.dp))
    }
}

private fun playlistSummary(author: String, trackCount: Int): String {
    if (author.isEmpty()) {
        return formatSongCount(trackCount)
    }
    return "$author • ${formatSongCount(trackCount)}"
}
