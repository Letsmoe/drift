package dev.letsmoe.drift.ui.components

import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.State
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.DownloadState
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.icons.DriftIcons
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.map

/** The download state of [videoId]; looks it up once and then follows the core's updates. */
@Composable
fun rememberDownloadState(videoId: String): State<DownloadState> {
    val downloads = LocalAppContainer.current.downloadsState
    LaunchedEffect(videoId) { downloads.ensureStates(listOf(videoId)) }
    // Only this track's state is collected, so progress of other downloads does not recompose the row.
    val stateFlow = remember(videoId) {
        downloads.statuses
            .map { downloads.stateOf(videoId) }
            .distinctUntilChanged()
    }
    return stateFlow.collectAsStateWithLifecycle(initialValue = downloads.stateOf(videoId))
}

/** A small marker for track rows: a filled arrow once downloaded, a spinner while it is on its way. */
@Composable
fun DownloadMark(state: DownloadState, modifier: Modifier = Modifier) {
    when (state) {
        DownloadState.Done -> Icon(
            imageVector = DriftIcons.Download.filled,
            contentDescription = "Downloaded",
            tint = MaterialTheme.colorScheme.primary,
            modifier = modifier.size(16.dp),
        )
        DownloadState.Queued, DownloadState.Downloading, DownloadState.Paused -> CircularProgressIndicator(
            modifier = modifier.size(14.dp),
            strokeWidth = 2.dp,
        )
        DownloadState.Failed -> Icon(
            imageVector = DriftIcons.Warning.filled,
            contentDescription = "Download failed",
            tint = MaterialTheme.colorScheme.error,
            modifier = modifier.size(16.dp),
        )
        DownloadState.None -> Unit
    }
}
