package dev.neoworks.ottermusic.ui.player

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.Orientation
import androidx.compose.foundation.gestures.draggable
import androidx.compose.foundation.gestures.rememberDraggableState
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.player.PlayerController
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.components.Cover
import dev.neoworks.ottermusic.ui.components.glassSurface
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.theme.OtterTheme
import kotlin.math.roundToInt

private const val SWIPE_THRESHOLD_PIXELS = 120f

/**
 * Floating card above the tab bar: cover, title and artist, play/pause, next and a 2dp progress
 * line. Swipe sideways for next/previous, tap to open Now Playing. Hidden when the queue is empty.
 */
@Composable
fun MiniPlayer(onOpen: () -> Unit, modifier: Modifier = Modifier) {
    val player = LocalAppContainer.current.playerController
    val track by player.currentTrack.collectAsStateWithLifecycle()
    val shownTrack = rememberLastNonNull(track)

    AnimatedVisibility(
        visible = track != null,
        modifier = modifier,
        enter = slideInVertically { fullHeight -> fullHeight } + fadeIn(),
        exit = slideOutVertically { fullHeight -> fullHeight } + fadeOut(),
    ) {
        if (shownTrack != null) {
            MiniPlayerCard(track = shownTrack, player = player, onOpen = onOpen)
        }
    }
}

@Composable
private fun <T : Any> rememberLastNonNull(value: T?): T? {
    val holder = remember { arrayOfNulls<Any>(1) }
    if (value != null) {
        holder[0] = value
    }
    @Suppress("UNCHECKED_CAST")
    return holder[0] as T?
}

@Composable
private fun MiniPlayerCard(track: Track, player: PlayerController, onOpen: () -> Unit) {
    val isPlaying by player.isPlaying.collectAsStateWithLifecycle()
    val isBuffering by player.isBuffering.collectAsStateWithLifecycle()
    var dragOffset by remember { mutableFloatStateOf(0f) }
    val dragState = rememberDraggableState { delta -> dragOffset += delta }

    Box(
        modifier = Modifier
            .padding(start = 12.dp, end = 12.dp, bottom = 8.dp)
            .offset { IntOffset(dragOffset.roundToInt() / 3, 0) }
            .fillMaxWidth()
            .glassSurface()
            .draggable(
                state = dragState,
                orientation = Orientation.Horizontal,
                onDragStopped = {
                    handleSwipe(dragOffset, player)
                    dragOffset = 0f
                },
            )
            .clickable(onClick = onOpen),
    ) {
        Row(
            modifier = Modifier
                .padding(start = 8.dp, end = 4.dp, top = 8.dp, bottom = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Cover(track.thumbnails, size = 40.dp)
            Column(modifier = Modifier.weight(1f)) {
                Text(track.title, style = MaterialTheme.typography.titleSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
                Text(
                    text = track.artistNames,
                    style = MaterialTheme.typography.bodySmall,
                    color = OtterTheme.colors.textMuted,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
            PlayPauseButton(isPlaying = isPlaying, isBuffering = isBuffering, onClick = player::toggle)
            IconButton(onClick = player::next) {
                Icon(OtterIcons.SkipForward.filled, contentDescription = "Next")
            }
        }
        MiniProgressLine(player = player, modifier = Modifier.align(Alignment.BottomStart))
    }
}

private fun handleSwipe(dragOffset: Float, player: PlayerController) {
    if (dragOffset < -SWIPE_THRESHOLD_PIXELS) {
        player.next()
        return
    }
    if (dragOffset > SWIPE_THRESHOLD_PIXELS) {
        player.previous()
    }
}

@Composable
private fun PlayPauseButton(isPlaying: Boolean, isBuffering: Boolean, onClick: () -> Unit) {
    IconButton(onClick = onClick) {
        if (isBuffering) {
            CircularProgressIndicator(
                modifier = Modifier.size(24.dp),
                strokeWidth = 2.dp,
                color = MaterialTheme.colorScheme.onSurface,
            )
            return@IconButton
        }
        if (isPlaying) {
            Icon(OtterIcons.Pause.filled, contentDescription = "Pause")
            return@IconButton
        }
        Icon(OtterIcons.Play.filled, contentDescription = "Play")
    }
}

@Composable
private fun MiniProgressLine(player: PlayerController, modifier: Modifier = Modifier) {
    val positionMs by player.positionMs.collectAsStateWithLifecycle()
    val durationMs by player.durationMs.collectAsStateWithLifecycle()
    var fraction = 0f
    if (durationMs > 0) {
        fraction = (positionMs.toFloat() / durationMs).coerceIn(0f, 1f)
    }
    Box(
        modifier = modifier
            .fillMaxWidth()
            .height(2.dp)
            .background(OtterTheme.colors.border),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth(fraction)
                .height(2.dp)
                .background(MaterialTheme.colorScheme.primary),
        )
    }
}
