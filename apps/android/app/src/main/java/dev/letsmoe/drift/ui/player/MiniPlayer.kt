package dev.letsmoe.drift.ui.player

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionScope
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.layout.Spacer
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.hapticfeedback.HapticFeedback
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalHapticFeedback
import kotlin.math.abs
import kotlinx.coroutines.launch
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.player.PlayerController
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.components.glassSurface
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme

private const val SWIPE_THRESHOLD_PIXELS = 120f
private const val SWIPE_VELOCITY_THRESHOLD = 900f
private const val DRAG_RESISTANCE = 3f

/**
 * Floating card above the tab bar: cover, title and artist, play/pause, next and a 2dp progress
 * line. Swipe sideways for next/previous, tap to open Now Playing. Hidden when the queue is empty
 * or while [isVisible] is false (Now Playing is open and owns the shared cover).
 */
@OptIn(ExperimentalSharedTransitionApi::class)
@Composable
fun MiniPlayer(
    isVisible: Boolean,
    sharedTransitionScope: SharedTransitionScope,
    onOpen: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val player = LocalAppContainer.current.playerController
    val track by player.currentTrack.collectAsStateWithLifecycle()
    val shownTrack = rememberLastNonNull(track)

    AnimatedVisibility(
        visible = track != null && isVisible,
        modifier = modifier,
        enter = slideInVertically { fullHeight -> fullHeight } + fadeIn(),
        exit = fadeOut(tween(durationMillis = 150)),
    ) {
        if (shownTrack != null) {
            val sharedCover = rememberSharedCoverModifier(sharedTransitionScope, this)
            MiniPlayerCard(track = shownTrack, player = player, sharedCover = sharedCover, onOpen = onOpen)
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

@Suppress("ModifierParameter")
@Composable
private fun MiniPlayerCard(track: Track, player: PlayerController, sharedCover: Modifier, onOpen: () -> Unit) {
    val isPlaying by player.isPlaying.collectAsStateWithLifecycle()
    val isBuffering by player.isBuffering.collectAsStateWithLifecycle()
    val dragOffset = remember { Animatable(0f) }
    val scope = rememberCoroutineScope()
    val haptics = LocalHapticFeedback.current
    val dragState = rememberDraggableState { delta ->
        scope.launch { dragOffset.snapTo(dragOffset.value + delta) }
    }

    Box(
        modifier = Modifier
            .padding(start = 12.dp, end = 12.dp, bottom = 8.dp)
            .graphicsLayer { translationX = dragOffset.value / DRAG_RESISTANCE }
            .fillMaxWidth()
            .glassSurface()
            .draggable(
                state = dragState,
                orientation = Orientation.Horizontal,
                onDragStopped = { velocity ->
                    scope.launch { settleSwipe(dragOffset, velocity, player, haptics) }
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
            Cover(track.thumbnails, size = 40.dp, modifier = sharedCover)
            TrackTitles(track = track, modifier = Modifier.weight(1f))
            PlayPauseButton(isPlaying = isPlaying, isBuffering = isBuffering, onClick = player::toggle)
            IconButton(onClick = player::next) {
                Icon(DriftIcons.SkipForward.filled, contentDescription = "Next")
            }
        }
        MiniProgressLine(player = player, modifier = Modifier.align(Alignment.BottomStart))
    }
}

// The text cross-fades when the track changes, e.g. after a swipe or a skip.
@Composable
private fun TrackTitles(track: Track, modifier: Modifier = Modifier) {
    AnimatedContent(
        targetState = track,
        modifier = modifier,
        transitionSpec = { fadeIn(tween(200)) togetherWith fadeOut(tween(120)) },
        contentKey = { shownTrack -> shownTrack.videoId },
        label = "miniTrackTitles",
    ) { shownTrack ->
        Column {
            Text(shownTrack.title, style = MaterialTheme.typography.titleSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(
                text = shownTrack.artistNames,
                style = MaterialTheme.typography.bodySmall,
                color = DriftTheme.colors.textMuted,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

private suspend fun settleSwipe(
    dragOffset: Animatable<Float, *>,
    velocity: Float,
    player: PlayerController,
    haptics: HapticFeedback,
) {
    val offset = dragOffset.value
    val isFarEnough = abs(offset) > SWIPE_THRESHOLD_PIXELS
    val isFastEnough = abs(velocity) > SWIPE_VELOCITY_THRESHOLD && velocity * offset >= 0f
    if (isFarEnough || isFastEnough) {
        haptics.performHapticFeedback(HapticFeedbackType.SegmentTick)
        skipInDirection(offset, velocity, player)
    }
    dragOffset.animateTo(0f, spring(dampingRatio = Spring.DampingRatioMediumBouncy, stiffness = Spring.StiffnessMedium), velocity)
}

private fun skipInDirection(offset: Float, velocity: Float, player: PlayerController) {
    val direction = if (offset != 0f) offset else velocity
    if (direction < 0f) {
        player.next()
        return
    }
    player.previous()
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
            Icon(DriftIcons.Pause.filled, contentDescription = "Pause")
            return@IconButton
        }
        Icon(DriftIcons.Play.filled, contentDescription = "Play")
    }
}

@Composable
private fun MiniProgressLine(player: PlayerController, modifier: Modifier = Modifier) {
    val positionMs by player.positionMs.collectAsStateWithLifecycle()
    val durationMs by player.durationMs.collectAsStateWithLifecycle()
    val trackColor = DriftTheme.colors.border
    val progressColor = MaterialTheme.colorScheme.primary
    // Position is read while drawing, so the ticker never recomposes the card.
    Spacer(
        modifier = modifier
            .fillMaxWidth()
            .height(2.dp)
            .drawBehind {
                drawRect(trackColor)
                if (durationMs > 0) {
                    val fraction = (positionMs.toFloat() / durationMs).coerceIn(0f, 1f)
                    drawRect(progressColor, size = Size(size.width * fraction, size.height))
                }
            },
    )
}
