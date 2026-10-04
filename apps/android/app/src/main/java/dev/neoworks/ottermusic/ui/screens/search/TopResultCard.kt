package dev.neoworks.ottermusic.ui.screens.search

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.components.Cover
import dev.neoworks.ottermusic.ui.components.RoundPlayButton
import dev.neoworks.ottermusic.ui.theme.OtterTheme

/** The best match of a search: big cover, title, kind and, for songs and episodes, a play button. */
@Composable
fun TopResultCard(item: Item, onClick: () -> Unit, modifier: Modifier = Modifier) {
    val playerController = LocalAppContainer.current.playerController
    val currentTrack by playerController.currentTrack.collectAsStateWithLifecycle()
    val isPlayerPlaying by playerController.isPlaying.collectAsStateWithLifecycle()
    val isTrackPlaying = item.track != null && currentTrack?.videoId == item.track.videoId && isPlayerPlaying

    Row(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(20.dp))
            .background(OtterTheme.colors.surfaceRaised)
            .clickable(onClick = onClick)
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Cover(thumbnails = item.thumbnails, size = 96.dp, circle = item.kind == ItemKind.Artist)
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = item.title,
                style = MaterialTheme.typography.titleLarge,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = topResultSubtitle(item),
                style = MaterialTheme.typography.bodySmall,
                color = OtterTheme.colors.textMuted,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
        }
        if (item.track != null) {
            RoundPlayButton(isPlaying = isTrackPlaying, onClick = onClick, size = 48.dp, contentDescription = "Play ${item.title}")
        }
    }
}

private fun topResultSubtitle(item: Item): String {
    val kindLabel = item.kind.name
    if (item.subtitle.isEmpty()) {
        return kindLabel
    }
    return item.subtitle
}
