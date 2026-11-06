package dev.letsmoe.drift.ui.screens.downloads

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.core.model.DownloadItem
import dev.letsmoe.drift.core.model.DownloadState
import dev.letsmoe.drift.core.model.DownloadStatus
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.formatBytes
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme

/** A download that is waiting, running or failed, with its progress and the cancel and retry buttons. */
@Composable
fun DownloadQueueRow(
    item: DownloadItem,
    status: DownloadStatus,
    onCancel: () -> Unit,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val isFailed = status.state == DownloadState.Failed
    var statusColor = DriftTheme.colors.textDim
    if (isFailed) {
        statusColor = MaterialTheme.colorScheme.error
    }
    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(start = 16.dp, end = 4.dp, top = 8.dp, bottom = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Cover(item.track.thumbnails, size = 48.dp)
        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text(
                text = item.track.title,
                style = MaterialTheme.typography.bodyLarge,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = describeStatus(status),
                style = MaterialTheme.typography.bodySmall,
                color = statusColor,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            if (status.state == DownloadState.Downloading) {
                LinearProgressIndicator(progress = { status.fraction }, modifier = Modifier.fillMaxWidth())
            }
        }
        if (isFailed) {
            IconButton(onClick = onRetry) {
                Icon(
                    DriftIcons.Download.outline,
                    contentDescription = "Retry download of ${item.track.title}",
                    tint = DriftTheme.colors.textMuted,
                )
            }
        }
        IconButton(onClick = onCancel) {
            Icon(
                DriftIcons.Close.outline,
                contentDescription = "Cancel download of ${item.track.title}",
                tint = DriftTheme.colors.textMuted,
            )
        }
    }
}

private fun describeStatus(status: DownloadStatus): String {
    return when (status.state) {
        DownloadState.Downloading -> "${formatBytes(status.bytesDone)} of ${formatBytes(status.bytesTotal)}"
        DownloadState.Paused -> "Paused"
        DownloadState.Failed -> status.error
        else -> "Waiting"
    }
}
