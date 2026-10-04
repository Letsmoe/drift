package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.TextOverflow
import dev.neoworks.ottermusic.ui.theme.OtterTheme

/** Text clamped to [collapsedLines]; tapping it expands it when it did not fit. */
@Composable
fun ExpandableText(
    text: String,
    modifier: Modifier = Modifier,
    collapsedLines: Int = 3,
    style: TextStyle = MaterialTheme.typography.bodyMedium,
    color: Color = OtterTheme.colors.textMuted,
) {
    var isExpanded by rememberSaveable { mutableStateOf(false) }
    var isOverflowing by rememberSaveable { mutableStateOf(false) }
    val visibleLines = if (isExpanded) Int.MAX_VALUE else collapsedLines
    Column(modifier = modifier.clickable(enabled = isOverflowing || isExpanded) { isExpanded = !isExpanded }) {
        Text(
            text = text,
            style = style,
            color = color,
            maxLines = visibleLines,
            overflow = TextOverflow.Ellipsis,
            onTextLayout = { layout ->
                if (!isExpanded) {
                    isOverflowing = layout.hasVisualOverflow
                }
            },
        )
        if (isOverflowing || isExpanded) {
            val toggleLabel = if (isExpanded) "Less" else "More"
            Text(text = toggleLabel, style = MaterialTheme.typography.labelLarge)
        }
    }
}
