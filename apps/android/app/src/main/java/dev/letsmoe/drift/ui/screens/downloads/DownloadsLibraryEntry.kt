package dev.letsmoe.drift.ui.screens.downloads

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
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.core.model.DownloadSummary
import dev.letsmoe.drift.ui.formatSongCount
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme

/** The Library's Downloads row: what is on the device and what is still on its way. Opens the Downloads page. */
@Composable
fun DownloadsLibraryEntry(summary: DownloadSummary, onClick: () -> Unit, modifier: Modifier = Modifier) {
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
            Icon(DriftIcons.Download.filled, contentDescription = null, modifier = Modifier.size(26.dp))
        }
        Column(modifier = Modifier.weight(1f)) {
            Text("Downloads", style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(
                text = downloadsSubtitle(summary),
                style = MaterialTheme.typography.bodySmall,
                color = DriftTheme.colors.textDim,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

private fun downloadsSubtitle(summary: DownloadSummary): String {
    val pending = summary.queued + summary.downloading
    if (pending > 0) {
        return "Downloading ${formatSongCount(pending)}"
    }
    if (summary.done == 0) {
        return "Songs that play without a connection"
    }
    return "${formatSongCount(summary.done)} on this device"
}
