package dev.letsmoe.drift.ui.screens.likedsongs

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.CollectionMoreButton
import dev.letsmoe.drift.ui.components.EnsureRatings
import dev.letsmoe.drift.ui.components.PlayShuffleButtons
import dev.letsmoe.drift.ui.components.TrackRow
import dev.letsmoe.drift.ui.formatSongCount
import dev.letsmoe.drift.ui.screens.detail.DetailActionRow
import dev.letsmoe.drift.ui.screens.detail.DetailHeader
import dev.letsmoe.drift.ui.screens.detail.DetailLoadState
import dev.letsmoe.drift.ui.screens.detail.DetailScaffold
import dev.letsmoe.drift.ui.screens.detail.detailListPadding
import dev.letsmoe.drift.ui.screens.library.LikedSongsCover
import dev.letsmoe.drift.ui.theme.DriftTheme

private const val LIKED_CONTEXT = "liked"
private val LikedHeaderColor = Color(0xFF2E2470)
private const val LIKED_RATING = 1

@Composable
fun LikedSongsScreen(modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> LikedSongsViewModel(container.core) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    DetailLoadState(state = state, onRetry = viewModel::retry, modifier = modifier) { likedTracks ->
        LikedSongsContent(loadedTracks = likedTracks, modifier = modifier)
    }
}

@Composable
private fun LikedSongsContent(loadedTracks: List<Track>, modifier: Modifier) {
    val container = LocalAppContainer.current
    val playerController = container.playerController
    val ratings by container.libraryState.ratings.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    EnsureRatings(loadedTracks)
    // A track whose rating changed away from like drops out without refetching.
    val tracks = loadedTracks.filter { track -> (ratings[track.videoId] ?: LIKED_RATING) == LIKED_RATING }

    DetailScaffold(title = "Liked Songs", listState = listState, modifier = modifier) {
        LazyColumn(state = listState, contentPadding = detailListPadding()) {
            item {
                DetailHeader(
                    title = "Liked Songs",
                    thumbnails = emptyList(),
                    gradientColor = LikedHeaderColor,
                    cover = { LikedSongsCover(size = 200.dp) },
                ) {
                    Text(
                        text = formatSongCount(tracks.size),
                        style = MaterialTheme.typography.bodyMedium,
                        color = DriftTheme.colors.textMuted,
                    )
                }
            }
            item {
                DetailActionRow { CollectionMoreButton(title = "Liked Songs", tracks = tracks) }
            }
            item {
                PlayShuffleButtons(
                    onPlay = { playerController.playContext(tracks, LIKED_CONTEXT) },
                    onShuffle = { playerController.shuffleContext(tracks, LIKED_CONTEXT) },
                )
            }
            if (tracks.isEmpty()) {
                item {
                    Text(
                        text = "Songs you like will appear here.",
                        style = MaterialTheme.typography.bodyMedium,
                        color = DriftTheme.colors.textDim,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.fillMaxWidth().padding(48.dp),
                    )
                }
            }
            itemsIndexed(tracks, key = { _, track -> track.videoId }) { index, track ->
                TrackRow(
                    track = track,
                    onClick = { playerController.playOrToggle(track, tracks, index, LIKED_CONTEXT) },
                )
            }
        }
    }
}
