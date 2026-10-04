package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.ui.theme.Canvas
import dev.neoworks.ottermusic.ui.theme.OtterTheme

/** A filter pill: raised when idle, inverted (light on dark) when selected. Green is kept for playback state. */
@Composable
fun OtterChip(
    label: String,
    isSelected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    var background = OtterTheme.colors.surfaceRaised
    var textColor = MaterialTheme.colorScheme.onSurface
    if (isSelected) {
        background = MaterialTheme.colorScheme.onSurface
        textColor = Canvas
    }
    Text(
        text = label,
        style = MaterialTheme.typography.labelLarge,
        color = textColor,
        maxLines = 1,
        modifier = modifier
            .clip(CircleShape)
            .background(background)
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 8.dp),
    )
}
