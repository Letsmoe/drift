package dev.letsmoe.drift.ui.screens.importing

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.importing.ImportState
import dev.letsmoe.drift.importing.ImportStep
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme

/**
 * The Library's import row: an invitation to import when idle, the progress of a running import,
 * and a pointer to the summary once it finished. Opens the import flow.
 */
@Composable
fun ImportLibraryEntry(state: ImportState, onClick: () -> Unit, modifier: Modifier = Modifier) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Box(
            modifier = Modifier.size(56.dp).clip(RoundedCornerShape(8.dp)).background(DriftTheme.colors.surfaceRaised),
            contentAlignment = Alignment.Center,
        ) {
            Icon(DriftIcons.Download.outline, contentDescription = null, modifier = Modifier.size(26.dp))
        }
        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text(title(state), style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(
                subtitle(state),
                style = MaterialTheme.typography.bodySmall,
                color = DriftTheme.colors.textDim,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            if (state.step == ImportStep.Running) {
                RunningBar(state)
            }
        }
    }
}

@Composable
private fun RunningBar(state: ImportState) {
    val progress = state.progress
    if (progress == null || progress.total == 0) {
        LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
        return
    }
    LinearProgressIndicator(progress = { state.progressFraction }, modifier = Modifier.fillMaxWidth())
}

private fun title(state: ImportState): String {
    return when (state.step) {
        ImportStep.Running -> state.phaseLabel
        ImportStep.Done -> "Import finished"
        else -> "Import from Spotify"
    }
}

private fun subtitle(state: ImportState): String {
    return when (state.step) {
        ImportStep.Running -> runningSubtitle(state)
        ImportStep.Done -> "See summary"
        else -> "Playlists, liked songs, albums and artists"
    }
}

private fun runningSubtitle(state: ImportState): String {
    val progress = state.progress
    if (progress == null || progress.total == 0) {
        return "Working"
    }
    return "${progress.done} of ${progress.total}"
}
