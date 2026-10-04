package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.theme.OtterTheme

/**
 * A track list row: cover, title, "Artist • Album", a playing indicator when it is the current
 * track and a "more" button that opens the [TrackActionSheet].
 *
 * @param subtitle replaces the default "Artist • Album" line (e.g. only the artist on album pages).
 * @param showCover hide the cover for album track lists.
 * @param onRemoveFromPlaylist adds "Remove from playlist" to the action sheet.
 */
@Composable
fun TrackRow(
    track: Track,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    kind: ItemKind = ItemKind.Song,
    subtitle: String = trackSubtitle(track),
    showCover: Boolean = true,
    onRemoveFromPlaylist: (() -> Unit)? = null,
) {
    val container = LocalAppContainer.current
    val currentTrack by container.playerController.currentTrack.collectAsStateWithLifecycle()
    val isPlaying by container.playerController.isPlaying.collectAsStateWithLifecycle()
    val likedVideoIds by container.libraryState.likedVideoIds.collectAsStateWithLifecycle()
    var showActions by remember { mutableStateOf(false) }

    val isCurrent = currentTrack?.videoId == track.videoId
    val titleColor = if (isCurrent) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface

    Row(
        modifier = modifier
            .fillMaxWidth()
            .heightIn(min = 64.dp)
            .clickable(onClick = onClick)
            .padding(start = 16.dp, end = 4.dp, top = 8.dp, bottom = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (showCover) {
            TrackCover(track = track, isCurrent = isCurrent, isPlaying = isPlaying)
        }
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = track.title,
                style = MaterialTheme.typography.bodyLarge,
                color = titleColor,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = subtitle,
                style = MaterialTheme.typography.bodySmall,
                color = OtterTheme.colors.textDim,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        if (track.videoId in likedVideoIds) {
            Icon(
                imageVector = OtterIcons.Heart.filled,
                contentDescription = "Liked",
                tint = MaterialTheme.colorScheme.primary,
                modifier = Modifier.size(16.dp),
            )
        }
        IconButton(onClick = { showActions = true }) {
            Icon(OtterIcons.More.outline, contentDescription = "More", tint = OtterTheme.colors.textMuted)
        }
    }

    if (showActions) {
        TrackActionSheet(
            track = track,
            kind = kind,
            onDismiss = { showActions = false },
            onRemoveFromPlaylist = onRemoveFromPlaylist,
        )
    }
}

@Composable
private fun TrackCover(track: Track, isCurrent: Boolean, isPlaying: Boolean) {
    Box(contentAlignment = Alignment.Center) {
        Cover(thumbnails = track.thumbnails, size = 48.dp)
        if (isCurrent) {
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(RoundedCornerShape(8.dp))
                    .background(Color.Black.copy(alpha = 0.55f)),
                contentAlignment = Alignment.Center,
            ) {
                PlayingIndicator(isPlaying = isPlaying, color = MaterialTheme.colorScheme.primary)
            }
        }
    }
}
