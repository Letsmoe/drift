package dev.neoworks.ottermusic.ui.screens.library

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
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
import dev.neoworks.ottermusic.ui.LocalBottomOverlayPadding
import dev.neoworks.ottermusic.ui.LocalNavController
import dev.neoworks.ottermusic.ui.appViewModel
import dev.neoworks.ottermusic.ui.components.NewPlaylistDialog
import dev.neoworks.ottermusic.ui.components.OtterChip
import dev.neoworks.ottermusic.ui.components.RootTopBar
import dev.neoworks.ottermusic.ui.formatSongCount
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.navigation.openAlbum
import dev.neoworks.ottermusic.ui.navigation.openArtist
import dev.neoworks.ottermusic.ui.navigation.openLikedSongs
import dev.neoworks.ottermusic.ui.navigation.openLocalPlaylist
import dev.neoworks.ottermusic.ui.navigation.openPlaylist
import dev.neoworks.ottermusic.ui.navigation.openPodcast
import dev.neoworks.ottermusic.ui.showToastOnFailure
import dev.neoworks.ottermusic.ui.theme.OtterTheme
import androidx.navigation.NavController
import dev.neoworks.ottermusic.core.model.ItemKind
import kotlinx.coroutines.launch

@Composable
fun LibraryScreen(modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> LibraryViewModel(container.core, container.libraryState) }
    val navController = LocalNavController.current
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val entries by viewModel.entries.collectAsStateWithLifecycle()
    val filter by viewModel.filter.collectAsStateWithLifecycle()
    val sort by viewModel.sort.collectAsStateWithLifecycle()
    val likedCount by viewModel.likedCount.collectAsStateWithLifecycle()
    var showNewPlaylistDialog by remember { mutableStateOf(false) }
    var sheetEntry by remember { mutableStateOf<LibraryEntry?>(null) }

    Column(modifier = modifier.fillMaxSize()) {
        RootTopBar(title = "Your Library") {
            IconButton(onClick = { showNewPlaylistDialog = true }) {
                Icon(OtterIcons.Add.outline, contentDescription = "New playlist")
            }
        }
        LazyColumn(contentPadding = PaddingValues(bottom = LocalBottomOverlayPadding.current)) {
            item { FilterBar(selected = filter, onToggle = viewModel::toggleFilter) }
            item { SortRow(sort = sort, onToggle = viewModel::toggleSort) }
            pinnedEntries(filter = filter, likedCount = likedCount, onOpenLikedSongs = navController::openLikedSongs)
            items(entries, key = { entry -> entry.key }) { entry ->
                LibraryRow(
                    title = entry.title,
                    subtitle = entry.subtitle,
                    playContext = entry.playContext,
                    thumbnails = entry.thumbnails,
                    isCircle = entry.isCircle,
                    onClick = { openEntry(navController, entry) },
                    onLongClick = { sheetEntry = entry },
                )
            }
            if (entries.isEmpty()) {
                item { EmptyLibrary(isFiltered = filter != null) }
            }
        }
    }

    if (showNewPlaylistDialog) {
        NewPlaylistDialog(
            onDismiss = { showNewPlaylistDialog = false },
            onCreate = { name ->
                showNewPlaylistDialog = false
                scope.launch {
                    showToastOnFailure(context) {
                        val created = viewModel.libraryState.createPlaylist(name)
                        navController.openLocalPlaylist(created.id)
                    }
                }
            },
        )
    }
    val entryForSheet = sheetEntry
    if (entryForSheet != null) {
        LibraryEntrySheet(
            entry = entryForSheet,
            libraryState = viewModel.libraryState,
            onDismiss = { sheetEntry = null },
        )
    }
}

private fun openEntry(navController: NavController, entry: LibraryEntry) {
    when (entry) {
        is LibraryEntry.LocalPlaylist -> navController.openLocalPlaylist(entry.summary.id)
        is LibraryEntry.Saved -> openSavedItem(navController, entry)
    }
}

private fun openSavedItem(navController: NavController, entry: LibraryEntry.Saved) {
    val id = entry.item.id
    when (entry.item.kind) {
        ItemKind.Album -> navController.openAlbum(id)
        ItemKind.Playlist -> navController.openPlaylist(id)
        ItemKind.Artist -> navController.openArtist(id)
        ItemKind.Podcast -> navController.openPodcast(id)
        ItemKind.Song, ItemKind.Episode -> Unit
    }
}

@Composable
private fun FilterBar(selected: LibraryFilter?, onToggle: (LibraryFilter) -> Unit) {
    LazyRow(
        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        items(LibraryFilter.entries) { filter ->
            OtterChip(label = filter.label, isSelected = filter == selected, onClick = { onToggle(filter) })
        }
    }
}

@Composable
private fun SortRow(sort: LibrarySort, onToggle: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onToggle)
            .padding(horizontal = 16.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Icon(OtterIcons.Clock.outline, contentDescription = null, tint = OtterTheme.colors.textMuted)
        Text(text = sort.label, style = MaterialTheme.typography.labelLarge)
    }
}

/**
 * Rows pinned above the list. Liked Songs is the only one today; a later "Import from Spotify"
 * entry belongs here, next to it (keep it visible with and without a filter).
 */
private fun LazyListScope.pinnedEntries(filter: LibraryFilter?, likedCount: Int, onOpenLikedSongs: () -> Unit) {
    val showsLikedSongs = filter == null || filter == LibraryFilter.Playlists
    if (!showsLikedSongs) {
        return
    }
    item(key = "pinned-liked") {
        LibraryRow(
            title = "Liked Songs",
            subtitle = "Playlist • ${formatSongCount(likedCount)}",
            playContext = "liked",
            onClick = onOpenLikedSongs,
            cover = { LikedSongsCover() },
        )
    }
}

@Composable
private fun EmptyLibrary(isFiltered: Boolean) {
    var message = "Playlists you create and albums, artists and podcasts you save will show up here."
    if (isFiltered) {
        message = "Nothing here yet."
    }
    Text(
        text = message,
        style = MaterialTheme.typography.bodyMedium,
        color = OtterTheme.colors.textDim,
        textAlign = TextAlign.Center,
        modifier = Modifier.fillMaxWidth().padding(horizontal = 32.dp, vertical = 48.dp),
    )
}
