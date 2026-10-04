package dev.neoworks.ottermusic.ui.screens.playlist

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
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.PlaylistPage
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.appViewModel
import dev.neoworks.ottermusic.ui.components.EnsureRatings
import dev.neoworks.ottermusic.ui.components.CollectionMoreButton
import dev.neoworks.ottermusic.ui.components.ExpandableText
import dev.neoworks.ottermusic.ui.components.LoadingMoreFooter
import dev.neoworks.ottermusic.ui.components.OnNearEnd
import dev.neoworks.ottermusic.ui.components.PlayShuffleButtons
import dev.neoworks.ottermusic.ui.components.SaveButton
import dev.neoworks.ottermusic.ui.components.TrackRow
import dev.neoworks.ottermusic.ui.formatSongCount
import dev.neoworks.ottermusic.ui.screens.detail.DetailActionRow
import dev.neoworks.ottermusic.ui.screens.detail.DetailHeader
import dev.neoworks.ottermusic.ui.screens.detail.DetailLoadState
import dev.neoworks.ottermusic.ui.screens.detail.DetailScaffold
import dev.neoworks.ottermusic.ui.screens.detail.detailListPadding
import dev.neoworks.ottermusic.ui.theme.OtterTheme

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
        color = OtterTheme.colors.textMuted,
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
