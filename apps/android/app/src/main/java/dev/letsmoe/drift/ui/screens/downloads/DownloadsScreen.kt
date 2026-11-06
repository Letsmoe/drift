package dev.letsmoe.drift.ui.screens.downloads

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavController
import dev.letsmoe.drift.core.model.CollectionKind
import dev.letsmoe.drift.core.model.DownloadCollection
import dev.letsmoe.drift.core.model.DownloadItem
import dev.letsmoe.drift.core.model.DownloadStatus
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.downloads.DownloadsState
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalBottomOverlayPadding
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.components.ConfirmDialog
import dev.letsmoe.drift.ui.components.DetailTopBar
import dev.letsmoe.drift.ui.components.PlayShuffleButtons
import dev.letsmoe.drift.ui.components.TrackRow
import dev.letsmoe.drift.ui.formatBytes
import dev.letsmoe.drift.ui.formatSongCount
import dev.letsmoe.drift.ui.navigation.openAlbum
import dev.letsmoe.drift.ui.navigation.openLikedSongs
import dev.letsmoe.drift.ui.navigation.openLocalPlaylist
import dev.letsmoe.drift.ui.navigation.openPlaylist
import dev.letsmoe.drift.ui.showToastOnFailure
import dev.letsmoe.drift.ui.theme.DriftTheme
import kotlinx.coroutines.launch

private const val DOWNLOADS_CONTEXT = "downloads"

/** The running queue, the collections kept downloaded and the songs that play without a connection. */
@Composable
fun DownloadsScreen(modifier: Modifier = Modifier) {
    val downloads = LocalAppContainer.current.downloadsState
    val navController = LocalNavController.current

    DisposableEffect(downloads) {
        downloads.startWatchingLists()
        onDispose { downloads.stopWatchingLists() }
    }

    Column(modifier = modifier.fillMaxSize()) {
        DetailTopBar(title = "Downloads", onBack = { navController.popBackStack() })
        DownloadsList(downloads = downloads, navController = navController)
    }
}

@Composable
private fun DownloadsList(downloads: DownloadsState, navController: NavController) {
    val playerController = LocalAppContainer.current.playerController
    val tracks by downloads.downloadedTracks.collectAsStateWithLifecycle()
    val queue by downloads.queue.collectAsStateWithLifecycle()
    val collections by downloads.collections.collectAsStateWithLifecycle()
    val statuses by downloads.statuses.collectAsStateWithLifecycle()
    val summary by downloads.summary.collectAsStateWithLifecycle()
    val usage by downloads.usage.collectAsStateWithLifecycle()
    val errorMessage by downloads.errorMessage.collectAsStateWithLifecycle()

    LazyColumn(contentPadding = PaddingValues(bottom = LocalBottomOverlayPadding.current)) {
        item {
            Text(
                text = "${formatSongCount(usage.count)} • ${formatBytes(usage.bytes)}",
                style = MaterialTheme.typography.bodyMedium,
                color = DriftTheme.colors.textMuted,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
            )
        }
        item {
            PlayShuffleButtons(
                onPlay = { playerController.playContext(tracks, DOWNLOADS_CONTEXT) },
                onShuffle = { playerController.shuffleContext(tracks, DOWNLOADS_CONTEXT) },
            )
        }
        item {
            DownloadsControls(
                downloads = downloads,
                isPaused = summary.paused,
                canDeleteAll = usage.count > 0 || queue.isNotEmpty(),
            )
        }
        notices(isBlocked = summary.blocked && !summary.paused, errorMessage = errorMessage)
        queueSection(downloads = downloads, queue = queue, statuses = statuses)
        collectionsSection(collections = collections, navController = navController)
        songsSection(tracks = tracks)
    }
}

@Composable
private fun DownloadsControls(downloads: DownloadsState, isPaused: Boolean, canDeleteAll: Boolean) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var confirmingDeleteAll by remember { mutableStateOf(false) }

    Row(
        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        OutlinedButton(
            onClick = { scope.launch { showToastOnFailure(context) { togglePause(downloads, isPaused) } } },
        ) {
            Text(pauseLabel(isPaused))
        }
        TextButton(onClick = { confirmingDeleteAll = true }, enabled = canDeleteAll) {
            Text("Delete all", color = MaterialTheme.colorScheme.error)
        }
    }

    if (confirmingDeleteAll) {
        ConfirmDialog(
            title = "Delete all downloads",
            message = "Delete every downloaded song from this device? Your library and playlists are not affected.",
            confirmLabel = "Delete all",
            onConfirm = {
                confirmingDeleteAll = false
                scope.launch { showToastOnFailure(context) { downloads.deleteAll() } }
            },
            onDismiss = { confirmingDeleteAll = false },
        )
    }
}

private suspend fun togglePause(downloads: DownloadsState, isPaused: Boolean) {
    if (isPaused) {
        downloads.resume()
        return
    }
    downloads.pause()
}

private fun pauseLabel(isPaused: Boolean): String {
    if (isPaused) {
        return "Resume downloads"
    }
    return "Pause downloads"
}

private fun LazyListScope.notices(isBlocked: Boolean, errorMessage: String) {
    if (isBlocked) {
        item(key = "blocked") {
            NoticeText(text = "Downloads wait for Wi-Fi. You can change this in Settings.", isError = false)
        }
    }
    if (errorMessage.isNotEmpty()) {
        item(key = "error") { NoticeText(text = errorMessage, isError = true) }
    }
}

@Composable
private fun NoticeText(text: String, isError: Boolean) {
    var color = DriftTheme.colors.textDim
    if (isError) {
        color = MaterialTheme.colorScheme.error
    }
    Text(
        text = text,
        style = MaterialTheme.typography.bodySmall,
        color = color,
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
    )
}

private fun LazyListScope.queueSection(
    downloads: DownloadsState,
    queue: List<DownloadItem>,
    statuses: Map<String, DownloadStatus>,
) {
    if (queue.isEmpty()) {
        return
    }
    item(key = "queue-title") { SectionTitle("Downloading") }
    items(queue, key = { item -> "queue-${item.track.videoId}" }) { item ->
        QueueItem(downloads = downloads, item = item, status = liveStatus(item, statuses))
    }
}

// The live status from the core's updates is fresher than the item loaded with the list.
private fun liveStatus(item: DownloadItem, statuses: Map<String, DownloadStatus>): DownloadStatus {
    val live = statuses[item.track.videoId]
    if (live == null) {
        return item.status
    }
    return live
}

@Composable
private fun QueueItem(downloads: DownloadsState, item: DownloadItem, status: DownloadStatus) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    DownloadQueueRow(
        item = item,
        status = status,
        onCancel = { scope.launch { showToastOnFailure(context) { downloads.cancel(item.track.videoId) } } },
        onRetry = { scope.launch { showToastOnFailure(context) { downloads.retry(item.track.videoId) } } },
    )
}

private fun LazyListScope.collectionsSection(collections: List<DownloadCollection>, navController: NavController) {
    if (collections.isEmpty()) {
        return
    }
    item(key = "collections-title") { SectionTitle("Collections") }
    items(collections, key = { collection -> "collection-${collection.kind}-${collection.id}" }) { collection ->
        CollectionRow(collection = collection, onClick = { openCollection(navController, collection) })
    }
}

@Composable
private fun CollectionRow(collection: DownloadCollection, onClick: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 8.dp),
    ) {
        Text(text = collection.title, style = MaterialTheme.typography.bodyLarge, maxLines = 1)
        Text(
            text = collectionSubtitle(collection),
            style = MaterialTheme.typography.bodySmall,
            color = DriftTheme.colors.textDim,
            maxLines = 1,
        )
    }
}

private fun collectionSubtitle(collection: DownloadCollection): String {
    val progress = "${collection.doneCount} of ${formatSongCount(collection.trackCount)}"
    if (collection.keep) {
        return "$progress • kept up to date"
    }
    return progress
}

private fun openCollection(navController: NavController, collection: DownloadCollection) {
    when (collection.kind) {
        CollectionKind.Album -> navController.openAlbum(collection.id)
        CollectionKind.Playlist -> navController.openPlaylist(collection.id)
        CollectionKind.LocalPlaylist -> navController.openLocalPlaylist(collection.id)
        CollectionKind.Likes -> navController.openLikedSongs()
    }
}

private fun LazyListScope.songsSection(tracks: List<Track>) {
    item(key = "songs-title") { SectionTitle("Downloaded songs") }
    if (tracks.isEmpty()) {
        item(key = "songs-empty") {
            Text(
                text = "Songs you download play without a connection and show up here.",
                style = MaterialTheme.typography.bodyMedium,
                color = DriftTheme.colors.textDim,
                textAlign = TextAlign.Center,
                modifier = Modifier.fillMaxWidth().padding(horizontal = 32.dp, vertical = 32.dp),
            )
        }
        return
    }
    itemsIndexed(tracks, key = { _, track -> "song-${track.videoId}" }) { index, track ->
        DownloadedTrackRow(track = track, tracks = tracks, index = index)
    }
}

@Composable
private fun DownloadedTrackRow(track: Track, tracks: List<Track>, index: Int) {
    val playerController = LocalAppContainer.current.playerController
    TrackRow(
        track = track,
        onClick = { playerController.playOrToggle(track, tracks, index, DOWNLOADS_CONTEXT) },
    )
}

@Composable
private fun SectionTitle(title: String) {
    Text(
        text = title,
        style = MaterialTheme.typography.titleMedium,
        modifier = Modifier.padding(start = 16.dp, end = 16.dp, top = 20.dp, bottom = 4.dp),
    )
}
