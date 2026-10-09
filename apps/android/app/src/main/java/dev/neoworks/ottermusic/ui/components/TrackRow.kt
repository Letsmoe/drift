package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
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
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalHapticFeedback
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.formatDuration
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.theme.OtterTheme

/**
 * A track list row: cover, title, "Artist • Album", a playing indicator when it is the current
 * track and a "more" button that opens the [TrackActionSheet].
 *
 * @param subtitle replaces the default "Artist • Album" line (e.g. only the artist on album pages).
 * @param showCover hide the cover for album track lists.
 * @param trackNumber shows a number slot before the cover (album and artist lists); the playing
 * indicator replaces the number for the current track.
 * @param showDuration shows the track length before the "more" button.
 * @param trailing extra content before the "more" button, e.g. a drag handle.
 * @param onRemoveFromPlaylist adds "Remove from playlist" to the action sheet.
 */
@OptIn(ExperimentalFoundationApi::class)
@Composable
fun TrackRow(
    track: Track,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    kind: ItemKind = ItemKind.Song,
    subtitle: String = trackSubtitle(track),
    showCover: Boolean = true,
    trackNumber: Int? = null,
    showDuration: Boolean = false,
    trailing: (@Composable () -> Unit)? = null,
    onRemoveFromPlaylist: (() -> Unit)? = null,
) {
    val container = LocalAppContainer.current
    val currentTrack by container.playerController.currentTrack.collectAsStateWithLifecycle()
    val isPlaying by container.playerController.isPlaying.collectAsStateWithLifecycle()
    val likedVideoIds by container.libraryState.likedVideoIds.collectAsStateWithLifecycle()
    var showActions by remember { mutableStateOf(false) }
    val haptics = LocalHapticFeedback.current

    val isCurrent = currentTrack?.videoId == track.videoId
    val titleColor = if (isCurrent) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface

    Row(
        modifier = modifier
            .fillMaxWidth()
            .heightIn(min = 64.dp)
            .combinedClickable(
                onClick = onClick,
                onLongClick = {
                    haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                    showActions = true
                },
            )
            .padding(start = 16.dp, end = 4.dp, top = 8.dp, bottom = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (trackNumber != null) {
            TrackNumber(number = trackNumber, isCurrent = isCurrent, isPlaying = isPlaying)
        }
        if (showCover) {
            TrackCover(track = track, isCurrent = isCurrent && trackNumber == null, isPlaying = isPlaying)
        }
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = track.title,
                style = MaterialTheme.typography.bodyLarge,
                color = titleColor,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            if (subtitle.isNotEmpty()) {
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = OtterTheme.colors.textDim,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
        if (track.videoId in likedVideoIds) {
            Icon(
                imageVector = OtterIcons.Heart.filled,
                contentDescription = "Liked",
                tint = MaterialTheme.colorScheme.primary,
                modifier = Modifier.size(16.dp),
            )
        }
        if (showDuration && track.durationSeconds > 0) {
            Text(
                text = formatDuration(track.durationSeconds),
                style = MaterialTheme.typography.bodySmall,
                color = OtterTheme.colors.textDim,
            )
        }
        trailing?.invoke()
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
private fun TrackNumber(number: Int, isCurrent: Boolean, isPlaying: Boolean) {
    Box(modifier = Modifier.width(28.dp), contentAlignment = Alignment.Center) {
        if (isCurrent) {
            PlayingIndicator(isPlaying = isPlaying, color = MaterialTheme.colorScheme.primary, size = 16.dp)
            return@Box
        }
        Text(
            text = number.toString(),
            style = MaterialTheme.typography.bodyMedium,
            color = OtterTheme.colors.textDim,
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
