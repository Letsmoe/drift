package dev.neoworks.ottermusic.ui.screens.localplaylist

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.Button
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.SnackbarResult
import androidx.compose.material3.SwipeToDismissBox
import androidx.compose.material3.SwipeToDismissBoxValue
import androidx.compose.material3.Text
import androidx.compose.material3.rememberSwipeToDismissBoxState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import android.widget.Toast
import dev.neoworks.ottermusic.core.model.LibraryPlaylist
import dev.neoworks.ottermusic.core.model.PlaylistEntry
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.LocalBottomOverlayPadding
import dev.neoworks.ottermusic.ui.LocalNavController
import dev.neoworks.ottermusic.ui.appViewModel
import dev.neoworks.ottermusic.ui.components.CollectionMoreButton
import dev.neoworks.ottermusic.ui.components.ConfirmDialog
import dev.neoworks.ottermusic.ui.components.CoverMosaic
import dev.neoworks.ottermusic.ui.components.EditPlaylistDialog
import dev.neoworks.ottermusic.ui.components.EnsureRatings
import dev.neoworks.ottermusic.ui.components.ExpandableText
import dev.neoworks.ottermusic.ui.components.PlayShuffleButtons
import dev.neoworks.ottermusic.ui.components.TrackRow
import dev.neoworks.ottermusic.ui.formatSongCount
import dev.neoworks.ottermusic.ui.formatTotalDuration
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.navigation.Routes
import dev.neoworks.ottermusic.ui.navigation.switchToTab
import dev.neoworks.ottermusic.ui.screens.detail.DetailActionRow
import dev.neoworks.ottermusic.ui.screens.detail.DetailHeader
import dev.neoworks.ottermusic.ui.screens.detail.DetailLoadState
import dev.neoworks.ottermusic.ui.screens.detail.DetailScaffold
import dev.neoworks.ottermusic.ui.screens.detail.detailListPadding
import dev.neoworks.ottermusic.ui.showToastOnFailure
import dev.neoworks.ottermusic.ui.theme.Canvas
import dev.neoworks.ottermusic.ui.theme.OtterTheme
import kotlinx.coroutines.launch
import sh.calvin.reorderable.ReorderableItem
import sh.calvin.reorderable.rememberReorderableLazyListState

private enum class PlaylistDialog { None, Edit, Delete }

@Composable
fun LocalPlaylistScreen(playlistId: String, modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container ->
        LocalPlaylistViewModel(container.core, container.libraryState, playlistId)
    }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val context = LocalContext.current
    LaunchedEffect(viewModel) {
        viewModel.failures.collect { message -> Toast.makeText(context, message, Toast.LENGTH_SHORT).show() }
    }
    DetailLoadState(state = state, onRetry = viewModel::retry, modifier = modifier) { playlist ->
        LocalPlaylistContent(playlist = playlist, viewModel = viewModel, modifier = modifier)
    }
}

@Composable
private fun LocalPlaylistContent(playlist: LibraryPlaylist, viewModel: LocalPlaylistViewModel, modifier: Modifier) {
    val entries by viewModel.entries.collectAsStateWithLifecycle()
    val navController = LocalNavController.current
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val listState = rememberLazyListState()
    val snackbarHostState = remember { SnackbarHostState() }
    var openDialog by remember { mutableStateOf(PlaylistDialog.None) }
    val tracks = entries.map { entry -> entry.track }
    EnsureRatings(tracks)

    fun removeWithUndo(entry: PlaylistEntry) {
        val removed = viewModel.removeEntry(entry) ?: return
        scope.launch {
            snackbarHostState.currentSnackbarData?.dismiss()
            val result = snackbarHostState.showSnackbar(
                message = "Removed ${entry.track.title}",
                actionLabel = "Undo",
                withDismissAction = true,
            )
            if (result == SnackbarResult.ActionPerformed) {
                viewModel.undoRemoval(removed)
            }
        }
    }

    Box(modifier = modifier) {
        DetailScaffold(title = playlist.name, listState = listState) {
            PlaylistList(
                playlist = playlist,
                entries = entries,
                listState = listState,
                viewModel = viewModel,
                onEdit = { openDialog = PlaylistDialog.Edit },
                onDelete = { openDialog = PlaylistDialog.Delete },
                onAddSongs = { navController.switchToTab(Routes.Search) },
                onRemove = ::removeWithUndo,
            )
        }
        SnackbarHost(
            hostState = snackbarHostState,
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .padding(bottom = LocalBottomOverlayPadding.current),
        )
    }

    if (openDialog == PlaylistDialog.Edit) {
        EditPlaylistDialog(
            name = playlist.name,
            description = playlist.description,
            onSave = { name, description ->
                openDialog = PlaylistDialog.None
                scope.launch { showToastOnFailure(context) { viewModel.saveDetails(name, description) } }
            },
            onDismiss = { openDialog = PlaylistDialog.None },
        )
    }
    if (openDialog == PlaylistDialog.Delete) {
        ConfirmDialog(
            title = "Delete playlist",
            message = "Delete \"${playlist.name}\" from your library? This cannot be undone.",
            confirmLabel = "Delete",
            onConfirm = {
                openDialog = PlaylistDialog.None
                scope.launch {
                    showToastOnFailure(context) {
                        viewModel.delete()
                        navController.popBackStack()
                    }
                }
            },
            onDismiss = { openDialog = PlaylistDialog.None },
        )
    }
}

@Composable
private fun PlaylistList(
    playlist: LibraryPlaylist,
    entries: List<PlaylistEntry>,
    listState: LazyListState,
    viewModel: LocalPlaylistViewModel,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
    onAddSongs: () -> Unit,
    onRemove: (PlaylistEntry) -> Unit,
) {
    val playerController = LocalAppContainer.current.playerController
    val context = "library:${playlist.id}"
    val tracks = entries.map { entry -> entry.track }
    val reorderableState = rememberReorderableLazyListState(listState) { from, to ->
        val fromId = from.key as? Long
        val toId = to.key as? Long
        if (fromId != null && toId != null) {
            viewModel.moveLocally(fromId, toId)
        }
    }

    LazyColumn(state = listState, contentPadding = detailListPadding()) {
        item(key = "header") { PlaylistHeader(playlist = playlist, entries = entries) }
        item(key = "actions") {
            PlaylistActions(
                playlistName = playlist.name,
                tracks = tracks,
                onEdit = onEdit,
                onDelete = onDelete,
                onAddSongs = onAddSongs,
            )
        }
        item(key = "playShuffle") {
            PlayShuffleButtons(
                onPlay = { playerController.playContext(tracks, context) },
                onShuffle = { playerController.shuffleContext(tracks, context) },
            )
        }
        if (entries.isEmpty()) {
            item(key = "empty") { EmptyPlaylist(onFindSongs = onAddSongs) }
        }
        items(entries, key = { entry -> entry.id }) { entry ->
            ReorderableItem(reorderableState, key = entry.id) { isDragging ->
                SwipeToRemoveRow(onRemove = { onRemove(entry) }) {
                    TrackRow(
                        track = entry.track,
                        onClick = {
                            val index = tracks.indexOfFirst { track -> track.videoId == entry.track.videoId }
                            playerController.playOrToggle(entry.track, tracks, index.coerceAtLeast(0), context)
                        },
                        modifier = Modifier.background(rowBackground(isDragging)),
                        onRemoveFromPlaylist = { onRemove(entry) },
                        trailing = {
                            IconButton(
                                onClick = {},
                                modifier = Modifier.draggableHandle(onDragStopped = { viewModel.commitMove(entry.id) }),
                            ) {
                                Icon(OtterIcons.DragHandle.outline, contentDescription = "Reorder", tint = OtterTheme.colors.textDim)
                            }
                        },
                    )
                }
            }
        }
    }
}

@Composable
private fun rowBackground(isDragging: Boolean): Color {
    if (isDragging) {
        return OtterTheme.colors.surfaceHover
    }
    return Canvas
}

@Composable
private fun PlaylistHeader(playlist: LibraryPlaylist, entries: List<PlaylistEntry>) {
    val coverThumbnails = entries.map { entry -> entry.track.thumbnails }.filter { thumbnails -> thumbnails.isNotEmpty() }
    val tracks = entries.map { entry -> entry.track }
    DetailHeader(
        title = playlist.name,
        thumbnails = coverThumbnails.firstOrNull() ?: emptyList(),
        cover = { CoverMosaic(coverThumbnails = coverThumbnails, size = 240.dp) },
    ) {
        Text(
            text = playlistSummary(tracks),
            style = MaterialTheme.typography.bodyMedium,
            color = OtterTheme.colors.textMuted,
        )
        if (playlist.description.isNotEmpty()) {
            ExpandableText(text = playlist.description, modifier = Modifier.padding(top = 8.dp))
        }
    }
}

private fun playlistSummary(tracks: List<Track>): String {
    if (tracks.isEmpty()) {
        return formatSongCount(0)
    }
    return "${formatSongCount(tracks.size)} • ${formatTotalDuration(tracks)}"
}

@Composable
private fun PlaylistActions(
    playlistName: String,
    tracks: List<Track>,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
    onAddSongs: () -> Unit,
) {
    DetailActionRow {
        IconButton(onClick = onAddSongs) {
            Icon(OtterIcons.Add.outline, contentDescription = "Add songs", tint = OtterTheme.colors.textMuted)
        }
        IconButton(onClick = onEdit) {
            Icon(OtterIcons.Edit.outline, contentDescription = "Edit details", tint = OtterTheme.colors.textMuted)
        }
        IconButton(onClick = onDelete) {
            Icon(OtterIcons.Delete.outline, contentDescription = "Delete playlist", tint = OtterTheme.colors.textMuted)
        }
        CollectionMoreButton(title = playlistName, tracks = tracks)
    }
}

@Composable
private fun EmptyPlaylist(onFindSongs: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxWidth().padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = "This playlist is empty. Add songs from the \"...\" menu of any track.",
            style = MaterialTheme.typography.bodyMedium,
            color = OtterTheme.colors.textDim,
            textAlign = TextAlign.Center,
            modifier = Modifier.padding(bottom = 16.dp),
        )
        Button(onClick = onFindSongs) { Text("Find songs") }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun SwipeToRemoveRow(onRemove: () -> Unit, content: @Composable () -> Unit) {
    val dismissState = rememberSwipeToDismissBoxState()
    LaunchedEffect(dismissState.currentValue) {
        if (dismissState.currentValue == SwipeToDismissBoxValue.EndToStart) {
            onRemove()
        }
    }
    SwipeToDismissBox(
        state = dismissState,
        enableDismissFromStartToEnd = false,
        backgroundContent = {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.error.copy(alpha = 0.85f))
                    .padding(horizontal = 24.dp),
                contentAlignment = Alignment.CenterEnd,
            ) {
                Icon(OtterIcons.Delete.outline, contentDescription = "Remove", tint = Color.White)
            }
        },
    ) {
        content()
    }
}
