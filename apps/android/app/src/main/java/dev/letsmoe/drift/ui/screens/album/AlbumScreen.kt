package dev.letsmoe.drift.ui.screens.album

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.AlbumPage
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.EnsureRatings
import dev.letsmoe.drift.ui.components.CollectionMoreButton
import dev.letsmoe.drift.ui.components.PlayShuffleButtons
import dev.letsmoe.drift.ui.components.SaveButton
import dev.letsmoe.drift.ui.components.TrackRow
import dev.letsmoe.drift.ui.formatSongCount
import dev.letsmoe.drift.ui.formatTotalDuration
import dev.letsmoe.drift.ui.navigation.openArtist
import dev.letsmoe.drift.ui.screens.detail.DetailActionRow
import dev.letsmoe.drift.ui.screens.detail.DetailHeader
import dev.letsmoe.drift.ui.screens.detail.DetailScaffold
import dev.letsmoe.drift.ui.screens.detail.DetailLoadState
import dev.letsmoe.drift.ui.screens.detail.detailListPadding
import dev.letsmoe.drift.ui.theme.DriftTheme

@Composable
fun AlbumScreen(albumId: String, modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> AlbumViewModel(container.core, albumId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    DetailLoadState(state = state, onRetry = viewModel::retry, modifier = modifier) { page ->
        AlbumContent(albumId = albumId, page = page, modifier = modifier)
    }
}

@Composable
private fun AlbumContent(albumId: String, page: AlbumPage, modifier: Modifier) {
    val playerController = LocalAppContainer.current.playerController
    val navController = LocalNavController.current
    val listState = rememberLazyListState()
    val context = "album:$albumId"
    val albumItem = Item(
        kind = ItemKind.Album,
        id = albumId,
        title = page.title,
        subtitle = page.subtitle,
        thumbnails = page.thumbnails,
    )
    val linkedArtist = page.artists.firstOrNull { artist -> artist.id.isNotEmpty() }
    val albumArtistNames = page.artists.joinToString(", ") { artist -> artist.name }
    EnsureRatings(page.tracks)

    DetailScaffold(title = page.title, listState = listState, modifier = modifier) {
        LazyColumn(state = listState, contentPadding = detailListPadding()) {
            item {
                DetailHeader(title = page.title, thumbnails = page.thumbnails) {
                    Text(
                        text = albumArtistNames,
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.onSurface,
                        modifier = Modifier
                            .clickable(enabled = linkedArtist != null) {
                                linkedArtist?.let { artist -> navController.openArtist(artist.id) }
                            }
                            .padding(vertical = 2.dp),
                    )
                    Text(
                        text = albumSubtitle(page),
                        style = MaterialTheme.typography.bodySmall,
                        color = DriftTheme.colors.textDim,
                    )
                }
            }
            item {
                DetailActionRow {
                    SaveButton(item = albumItem)
                    CollectionMoreButton(title = page.title, tracks = page.tracks)
                }
            }
            item {
                PlayShuffleButtons(
                    onPlay = { playerController.playContext(page.tracks, context) },
                    onShuffle = { playerController.shuffleContext(page.tracks, context) },
                )
            }
            itemsIndexed(page.tracks, key = { index, track -> "$index-${track.videoId}" }) { index, track ->
                TrackRow(
                    track = track,
                    onClick = { playerController.playOrToggle(track, page.tracks, index, context) },
                    subtitle = trackArtistsUnlessAlbumArtists(track, albumArtistNames),
                    showCover = false,
                    trackNumber = index + 1,
                    showDuration = true,
                )
            }
            item { AlbumFooter(page) }
        }
    }
}

private fun albumSubtitle(page: AlbumPage): String {
    if (page.year.isEmpty()) {
        return "Album"
    }
    return "Album • ${page.year}"
}

// Tracks by the album's own artists show only their title, like Apple Music.
private fun trackArtistsUnlessAlbumArtists(track: Track, albumArtistNames: String): String {
    if (track.artistNames == albumArtistNames) {
        return ""
    }
    return track.artistNames
}

@Composable
private fun AlbumFooter(page: AlbumPage) {
    if (page.tracks.isEmpty()) {
        return
    }
    Box(modifier = Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.CenterStart) {
        Text(
            text = "${formatSongCount(page.tracks.size)}, ${formatTotalDuration(page.tracks)}",
            style = MaterialTheme.typography.bodySmall,
            color = DriftTheme.colors.textDim,
        )
    }
}
