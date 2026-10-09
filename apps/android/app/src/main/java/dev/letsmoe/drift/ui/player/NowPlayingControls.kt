package dev.letsmoe.drift.ui.player

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.ui.unit.DpSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Slider
import androidx.compose.material3.SliderDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.RepeatMode
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme
import java.util.Locale

private val SEEK_THUMB_SIZE = 14.dp
private val SEEK_TRACK_HEIGHT = 4.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun SeekBar() {
    val player = LocalAppContainer.current.playerController
    val positionMs by player.positionMs.collectAsStateWithLifecycle()
    val durationMs by player.durationMs.collectAsStateWithLifecycle()
    var isDragging by remember { mutableStateOf(false) }
    var dragValueMs by remember { mutableFloatStateOf(0f) }

    val safeDuration = durationMs.coerceAtLeast(1L).toFloat()
    var shownMs = positionMs.toFloat()
    if (isDragging) {
        shownMs = dragValueMs
    }
    shownMs = shownMs.coerceIn(0f, safeDuration)

    val sliderColors = SliderDefaults.colors(
        thumbColor = Color.White,
        activeTrackColor = Color.White,
        inactiveTrackColor = Color.White.copy(alpha = 0.25f),
    )
    val interactionSource = remember { MutableInteractionSource() }
    Slider(
        value = shownMs,
        onValueChange = { value ->
            isDragging = true
            dragValueMs = value
        },
        onValueChangeFinished = {
            player.seekTo(dragValueMs.toLong())
            isDragging = false
        },
        valueRange = 0f..safeDuration,
        interactionSource = interactionSource,
        colors = sliderColors,
        thumb = {
            SliderDefaults.Thumb(
                interactionSource = interactionSource,
                colors = sliderColors,
                thumbSize = DpSize(SEEK_THUMB_SIZE, SEEK_THUMB_SIZE),
            )
        },
        track = { sliderState ->
            SliderDefaults.Track(
                sliderState = sliderState,
                colors = sliderColors,
                thumbTrackGapSize = 0.dp,
                trackInsideCornerSize = 0.dp,
                drawStopIndicator = null,
                modifier = Modifier.height(SEEK_TRACK_HEIGHT),
            )
        },
    )
    Row(modifier = Modifier.fillMaxWidth().padding(horizontal = 4.dp), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(formatTime(shownMs.toLong()), style = MaterialTheme.typography.labelMedium, color = DriftTheme.colors.textMuted)
        Text(
            text = "-" + formatTime((safeDuration - shownMs).toLong()),
            style = MaterialTheme.typography.labelMedium,
            color = DriftTheme.colors.textMuted,
        )
    }
}

internal fun formatTime(milliseconds: Long): String {
    val totalSeconds = (milliseconds / 1000).coerceAtLeast(0)
    return String.format(Locale.US, "%d:%02d", totalSeconds / 60, totalSeconds % 60)
}

/** Shuffle, previous, big round play/pause, next, repeat. */
@Composable
internal fun TransportControls() {
    val player = LocalAppContainer.current.playerController
    val queue by player.queue.collectAsStateWithLifecycle()
    val isPlaying by player.isPlaying.collectAsStateWithLifecycle()
    val isBuffering by player.isBuffering.collectAsStateWithLifecycle()
    val accent = MaterialTheme.colorScheme.primary
    val idle = MaterialTheme.colorScheme.onSurface

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        IconButton(onClick = player::toggleShuffle) {
            Icon(DriftIcons.Shuffle.outline, contentDescription = "Shuffle", tint = if (queue.shuffle) accent else idle)
        }
        IconButton(onClick = player::previous, modifier = Modifier.size(56.dp)) {
            Icon(DriftIcons.SkipBack.filled, contentDescription = "Previous", modifier = Modifier.size(32.dp))
        }
        PlayPauseCircle(isPlaying = isPlaying, isBuffering = isBuffering, onClick = player::toggle)
        IconButton(onClick = player::next, modifier = Modifier.size(56.dp)) {
            Icon(DriftIcons.SkipForward.filled, contentDescription = "Next", modifier = Modifier.size(32.dp))
        }
        IconButton(onClick = player::cycleRepeat) {
            val icon = if (queue.repeat == RepeatMode.One) DriftIcons.RepeatOnce else DriftIcons.Repeat
            val tint = if (queue.repeat == RepeatMode.Off) idle else accent
            Icon(icon.outline, contentDescription = "Repeat", tint = tint)
        }
    }
}

@Composable
private fun PlayPauseCircle(isPlaying: Boolean, isBuffering: Boolean, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .size(72.dp)
            .clip(CircleShape)
            .background(Color.White)
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        if (isBuffering) {
            CircularProgressIndicator(modifier = Modifier.size(32.dp), strokeWidth = 3.dp, color = Color.Black)
            return@Box
        }
        val icon = if (isPlaying) DriftIcons.Pause else DriftIcons.Play
        Icon(icon.filled, contentDescription = if (isPlaying) "Pause" else "Play", tint = Color.Black, modifier = Modifier.size(34.dp))
    }
}

/** Autoplay toggle on the left, queue button on the right. */
@Composable
internal fun NowPlayingBottomRow(onOpenQueue: () -> Unit) {
    val player = LocalAppContainer.current.playerController
    val queue by player.queue.collectAsStateWithLifecycle()
    val autoplayTint = if (queue.autoplay) MaterialTheme.colorScheme.primary else DriftTheme.colors.textMuted

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        IconButton(onClick = { player.setAutoplay(!queue.autoplay) }) {
            Icon(DriftIcons.Infinity.outline, contentDescription = "Autoplay", tint = autoplayTint)
        }
        IconButton(onClick = onOpenQueue) {
            Icon(DriftIcons.Queue.outline, contentDescription = "Queue")
        }
    }
}
