package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.LocalIndication
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.remember
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.minimumInteractiveComponentSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.ui.theme.Canvas
import dev.neoworks.ottermusic.ui.theme.OtterTheme

private const val CHIP_PRESSED_SCALE = 0.94f

/** A filter pill: raised when idle, inverted (light on dark) when selected. Green is kept for playback state. */
@Composable
fun OtterChip(
    label: String,
    isSelected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val interactionSource = remember { MutableInteractionSource() }
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
            .minimumInteractiveComponentSize()
            .pressScale(interactionSource, CHIP_PRESSED_SCALE)
            .clip(CircleShape)
            .background(background)
            .clickable(interactionSource = interactionSource, indication = LocalIndication.current, onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 8.dp),
    )
}
