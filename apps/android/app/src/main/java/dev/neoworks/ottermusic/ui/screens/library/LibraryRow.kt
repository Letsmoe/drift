package dev.neoworks.ottermusic.ui.screens.library

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.core.model.Thumbnail
import dev.neoworks.ottermusic.ui.components.Cover
import dev.neoworks.ottermusic.ui.components.PlayingIndicator
import dev.neoworks.ottermusic.ui.components.rememberIsContextCurrent
import dev.neoworks.ottermusic.ui.components.rememberIsContextPlaying
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.theme.OtterTheme

private val LibraryCoverSize = 56.dp
val LikedSongsGradient = Brush.linearGradient(listOf(Color(0xFF6D4AFF), Color(0xFF3B82F6)))

/** A library list row: 56dp cover (round for artists), title, subtitle and a playing indicator. */
@OptIn(ExperimentalFoundationApi::class)
@Composable
fun LibraryRow(
    title: String,
    subtitle: String,
    playContext: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    thumbnails: List<Thumbnail> = emptyList(),
    isCircle: Boolean = false,
    onLongClick: (() -> Unit)? = null,
    cover: (@Composable () -> Unit)? = null,
) {
    val isCurrent = rememberIsContextCurrent(playContext)
    val isPlaying = rememberIsContextPlaying(playContext)
    var titleColor = MaterialTheme.colorScheme.onSurface
    if (isCurrent) {
        titleColor = MaterialTheme.colorScheme.primary
    }
    Row(
        modifier = modifier
            .fillMaxWidth()
            .combinedClickable(onClick = onClick, onLongClick = onLongClick)
            .padding(horizontal = 16.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        if (cover != null) {
            cover()
        } else {
            Cover(thumbnails = thumbnails, size = LibraryCoverSize, circle = isCircle)
        }
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = title,
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
        if (isCurrent) {
            PlayingIndicator(isPlaying = isPlaying, color = MaterialTheme.colorScheme.primary)
        }
    }
}

/** The gradient square with a heart used for Liked Songs. */
@Composable
fun LikedSongsCover(modifier: Modifier = Modifier, size: Dp = LibraryCoverSize) {
    Box(
        modifier = modifier
            .size(size)
            .clip(RoundedCornerShape(8.dp))
            .background(LikedSongsGradient),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = OtterIcons.Heart.filled,
            contentDescription = null,
            tint = Color.White,
            modifier = Modifier.size(size / 2),
        )
    }
}
