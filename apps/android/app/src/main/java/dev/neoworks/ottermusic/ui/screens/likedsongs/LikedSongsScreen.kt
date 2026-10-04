package dev.neoworks.ottermusic.ui.screens.likedsongs

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
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.appViewModel
import dev.neoworks.ottermusic.ui.components.CollectionMoreButton
import dev.neoworks.ottermusic.ui.components.EnsureRatings
import dev.neoworks.ottermusic.ui.components.PlayShuffleButtons
import dev.neoworks.ottermusic.ui.components.TrackRow
import dev.neoworks.ottermusic.ui.formatSongCount
import dev.neoworks.ottermusic.ui.screens.detail.DetailActionRow
import dev.neoworks.ottermusic.ui.screens.detail.DetailHeader
import dev.neoworks.ottermusic.ui.screens.detail.DetailLoadState
import dev.neoworks.ottermusic.ui.screens.detail.DetailScaffold
import dev.neoworks.ottermusic.ui.screens.detail.detailListPadding
import dev.neoworks.ottermusic.ui.screens.library.LikedSongsCover
import dev.neoworks.ottermusic.ui.theme.OtterTheme

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
                        color = OtterTheme.colors.textMuted,
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
                        color = OtterTheme.colors.textDim,
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
