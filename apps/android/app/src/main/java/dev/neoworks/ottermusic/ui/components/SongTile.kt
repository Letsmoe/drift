package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.gestures.snapping.rememberSnapFlingBehavior
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.theme.OtterTheme

private const val SONGS_PER_COLUMN = 4
private val NextColumnPeek = 40.dp

/** A compact song row (cover, title, artist) for quick-pick grids; the playing track is green. */
@Composable
fun SongTile(track: Track, onClick: () -> Unit, modifier: Modifier = Modifier) {
    val playerController = LocalAppContainer.current.playerController
    val currentTrack by playerController.currentTrack.collectAsStateWithLifecycle()
    val isPlaying by playerController.isPlaying.collectAsStateWithLifecycle()
    val isCurrent = currentTrack?.videoId == track.videoId
    var titleColor = MaterialTheme.colorScheme.onSurface
    if (isCurrent) {
        titleColor = MaterialTheme.colorScheme.primary
    }
    Row(
        modifier = modifier
            .fillMaxWidth()
            .height(64.dp)
            .clickable(onClick = onClick)
            .padding(vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Cover(thumbnails = track.thumbnails, size = 52.dp)
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = track.title,
                style = MaterialTheme.typography.bodyLarge,
                color = titleColor,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = track.artistNames,
                style = MaterialTheme.typography.bodySmall,
                color = OtterTheme.colors.textDim,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        if (isCurrent) {
            PlayingIndicator(isPlaying = isPlaying, color = MaterialTheme.colorScheme.primary)
        }
    }
}

/**
 * YouTube Music style quick picks: songs stacked four to a column in a horizontally snapping
 * row, with a peek of the next column. [onSongVisible] fires once a song's tile is composed.
 */
@Composable
fun SongGrid(
    tracks: List<Track>,
    onSongClick: (index: Int, track: Track) -> Unit,
    modifier: Modifier = Modifier,
    onSongVisible: (index: Int, track: Track) -> Unit = { _, _ -> },
) {
    val listState = rememberLazyListState()
    val columns = tracks.chunked(SONGS_PER_COLUMN)
    BoxWithConstraints(modifier = modifier.fillMaxWidth()) {
        val columnWidth = maxWidth - 16.dp - NextColumnPeek
        LazyRow(
            state = listState,
            flingBehavior = rememberSnapFlingBehavior(listState),
            contentPadding = PaddingValues(horizontal = 16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            itemsIndexed(columns) { columnIndex, columnTracks ->
                Column(modifier = Modifier.width(columnWidth)) {
                    columnTracks.forEachIndexed { rowIndex, track ->
                        val index = columnIndex * SONGS_PER_COLUMN + rowIndex
                        LaunchedEffect(track.videoId) { onSongVisible(index, track) }
                        SongTile(track = track, onClick = { onSongClick(index, track) })
                    }
                }
            }
        }
    }
}
