package dev.neoworks.ottermusic.ui.components

import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.ui.Alignment
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.icons.OtterIcons

/** True while the queue was started from [context] and is playing; recomposes on queue and transport changes. */
@Composable
fun rememberIsContextPlaying(context: String): Boolean {
    val playerController = LocalAppContainer.current.playerController
    val queue by playerController.queue.collectAsStateWithLifecycle()
    val currentTrack by playerController.currentTrack.collectAsStateWithLifecycle()
    val isPlaying by playerController.isPlaying.collectAsStateWithLifecycle()
    return queue.context == context && currentTrack != null && isPlaying
}

/** True while the queue was started from [context] and has a current track, playing or paused. */
@Composable
fun rememberIsContextCurrent(context: String): Boolean {
    val playerController = LocalAppContainer.current.playerController
    val queue by playerController.queue.collectAsStateWithLifecycle()
    val currentTrack by playerController.currentTrack.collectAsStateWithLifecycle()
    return queue.context == context && currentTrack != null
}

/** A round green play / pause button; the only filled accent control besides Play pills. */
@Composable
fun RoundPlayButton(
    isPlaying: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    size: Dp = 56.dp,
    contentDescription: String = "Play",
) {
    Box(
        modifier = modifier
            .size(size)
            .clip(CircleShape)
            .background(MaterialTheme.colorScheme.primary)
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = playPauseIcon(isPlaying),
            contentDescription = contentDescription,
            tint = MaterialTheme.colorScheme.onPrimary,
            modifier = Modifier.size(size / 2),
        )
    }
}

/** The filled pause icon while playing, otherwise the filled play icon. */
fun playPauseIcon(isPlaying: Boolean): ImageVector {
    if (isPlaying) {
        return OtterIcons.Pause.filled
    }
    return OtterIcons.Play.filled
}
