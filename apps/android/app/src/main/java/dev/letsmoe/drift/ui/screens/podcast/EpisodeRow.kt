package dev.letsmoe.drift.ui.screens.podcast

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.Episode
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.formatDuration
import dev.letsmoe.drift.ui.components.playPauseIcon
import dev.letsmoe.drift.ui.theme.DriftTheme

/**
 * A podcast episode: cover, date, title, description, a round play button with the duration and,
 * when the episode was started before, a thin progress bar from the library's resume position.
 */
@Composable
fun EpisodeRow(episode: Episode, onPlay: () -> Unit, modifier: Modifier = Modifier) {
    val container = LocalAppContainer.current
    val track = episode.track
    val currentTrack by container.playerController.currentTrack.collectAsStateWithLifecycle()
    val isPlayerPlaying by container.playerController.isPlaying.collectAsStateWithLifecycle()
    val resumeSeconds = rememberResumeSeconds(track.videoId)
    val isCurrent = currentTrack?.videoId == track.videoId
    var titleColor = MaterialTheme.colorScheme.onSurface
    if (isCurrent) {
        titleColor = MaterialTheme.colorScheme.primary
    }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onPlay)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            Cover(thumbnails = track.thumbnails, size = 64.dp)
            Column(modifier = Modifier.weight(1f)) {
                if (episode.published.isNotEmpty()) {
                    Text(episode.published, style = MaterialTheme.typography.labelMedium, color = DriftTheme.colors.textDim)
                }
                Text(
                    text = track.title,
                    style = MaterialTheme.typography.titleSmall,
                    color = titleColor,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
        if (episode.description.isNotEmpty()) {
            Text(
                text = episode.description,
                style = MaterialTheme.typography.bodySmall,
                color = DriftTheme.colors.textMuted,
                maxLines = 3,
                overflow = TextOverflow.Ellipsis,
            )
        }
        EpisodeFooter(
            durationSeconds = track.durationSeconds,
            resumeSeconds = resumeSeconds,
            isPlaying = isCurrent && isPlayerPlaying,
            onPlay = onPlay,
        )
    }
}

@Composable
private fun EpisodeFooter(durationSeconds: Int, resumeSeconds: Int, isPlaying: Boolean, onPlay: () -> Unit) {
    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        Box(
            modifier = Modifier
                .size(40.dp)
                .clip(CircleShape)
                .background(DriftTheme.colors.surfaceHover)
                .clickable(onClick = onPlay),
            contentAlignment = Alignment.Center,
        ) {
            Icon(playPauseIcon(isPlaying), contentDescription = "Play", modifier = Modifier.size(20.dp))
        }
        Text(
            text = remainingLabel(durationSeconds, resumeSeconds),
            style = MaterialTheme.typography.bodySmall,
            color = DriftTheme.colors.textDim,
        )
        if (resumeSeconds > 0 && durationSeconds > 0) {
            LinearProgressIndicator(
                progress = { (resumeSeconds.toFloat() / durationSeconds).coerceIn(0f, 1f) },
                modifier = Modifier.weight(1f).height(3.dp),
                color = MaterialTheme.colorScheme.onSurface,
                trackColor = DriftTheme.colors.surfaceHover,
            )
        }
    }
}

private fun remainingLabel(durationSeconds: Int, resumeSeconds: Int): String {
    if (resumeSeconds <= 0 || resumeSeconds >= durationSeconds) {
        return formatDuration(durationSeconds)
    }
    return "${formatDuration(durationSeconds - resumeSeconds)} left"
}

// One local database read per visible row; cheap enough to avoid batching.
@Composable
private fun rememberResumeSeconds(videoId: String): Int {
    val core = LocalAppContainer.current.core
    var resumeSeconds by remember(videoId) { mutableIntStateOf(0) }
    LaunchedEffect(videoId) {
        val position = runCatching { core.resumePosition(videoId) }.getOrNull()
        if (position != null && position.found) {
            resumeSeconds = position.seconds
        }
    }
    return resumeSeconds
}
