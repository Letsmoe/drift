package dev.neoworks.ottermusic.ui.components

import android.content.Context
import android.widget.Toast
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.Album
import dev.neoworks.ottermusic.core.model.Artist
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.PlaylistSummary
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.player.PlayerController
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.LocalNavController
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.navigation.openAlbum
import dev.neoworks.ottermusic.ui.navigation.openArtist
import dev.neoworks.ottermusic.ui.theme.OtterTheme
import kotlinx.coroutines.launch

/**
 * The "more" sheet of a track, modeled on YouTube Music's: header with like toggle, then play
 * next, add to queue, save to playlist, radio, go to album / artist and (when
 * [onRemoveFromPlaylist] is given) remove from playlist.
 *
 * [onOpenAlbum] and [onOpenArtist] default to navigating with the app's NavController; override
 * them to also close an overlay.
 */
@Composable
fun TrackActionSheet(
    track: Track,
    onDismiss: () -> Unit,
    kind: ItemKind = ItemKind.Song,
    onRemoveFromPlaylist: (() -> Unit)? = null,
    onOpenAlbum: ((Album) -> Unit)? = null,
    onOpenArtist: ((Artist) -> Unit)? = null,
) {
    val navController = LocalNavController.current
    var showPlaylistPicker by remember { mutableStateOf(false) }
    var showArtistPicker by remember { mutableStateOf(false) }

    val openAlbum = onOpenAlbum ?: { album: Album -> navController.openAlbum(album.id) }
    val openArtist = onOpenArtist ?: { artist: Artist -> navController.openArtist(artist.id) }
    val linkedArtists = track.artists.filter { artist -> artist.id.isNotEmpty() }

    // The picker replaces the action sheet instead of stacking on it, so nothing shows through.
    if (!showPlaylistPicker) {
        TrackActionRows(
            track = track,
            kind = kind,
            onDismiss = onDismiss,
            onRemoveFromPlaylist = onRemoveFromPlaylist,
            onOpenAlbum = openAlbum,
            onOpenArtist = openArtist,
            onSaveToPlaylist = { showPlaylistPicker = true },
            onPickArtist = { showArtistPicker = true },
        )
    }

    if (showPlaylistPicker) {
        PlaylistPickerSheet(
            tracks = listOf(track),
            onDismiss = {
                showPlaylistPicker = false
                onDismiss()
            },
        )
    }
    if (showArtistPicker) {
        ArtistPickerDialog(
            artists = linkedArtists,
            onPick = { artist ->
                showArtistPicker = false
                onDismiss()
                openArtist(artist)
            },
            onDismiss = { showArtistPicker = false },
        )
    }
}

@Composable
private fun TrackActionRows(
    track: Track,
    kind: ItemKind,
    onDismiss: () -> Unit,
    onRemoveFromPlaylist: (() -> Unit)?,
    onOpenAlbum: (Album) -> Unit,
    onOpenArtist: (Artist) -> Unit,
    onSaveToPlaylist: () -> Unit,
    onPickArtist: () -> Unit,
) {
    val container = LocalAppContainer.current
    val context = LocalContext.current
    // Outlives the sheet, which closes right after starting the radio.
    val scope = container.applicationScope
    val linkedArtists = track.artists.filter { artist -> artist.id.isNotEmpty() }
    val linkedAlbum = track.album?.takeIf { album -> album.id.isNotEmpty() }

    FloatingSheet(onDismiss = onDismiss) { dismiss ->
        TrackSheetHeader(track = track, kind = kind)
        SheetDivider()
        ActionRow(OtterIcons.SkipForward, "Play next", {
            container.playerController.playNext(listOf(track))
            Toast.makeText(context, "Playing next", Toast.LENGTH_SHORT).show()
            dismiss()
        })
        ActionRow(OtterIcons.Queue, "Add to queue", {
            container.playerController.addToQueue(listOf(track))
            Toast.makeText(context, "Added to queue", Toast.LENGTH_SHORT).show()
            dismiss()
        })
        ActionRow(OtterIcons.AddToPlaylist, "Save to playlist", { onSaveToPlaylist() })
        ActionRow(OtterIcons.Radio, "Start radio", {
            scope.launch { startRadio(container.core, container.playerController, track, context) }
            dismiss()
        })
        if (linkedAlbum != null) {
            ActionRow(OtterIcons.Album, "Go to album", {
                dismiss()
                onOpenAlbum(linkedAlbum)
            })
        }
        if (linkedArtists.size == 1) {
            ActionRow(OtterIcons.Artist, "Go to artist", {
                dismiss()
                onOpenArtist(linkedArtists.first())
            })
        }
        if (linkedArtists.size > 1) {
            ActionRow(OtterIcons.Artist, "Go to artist", { onPickArtist() })
        }
        if (onRemoveFromPlaylist != null) {
            ActionRow(OtterIcons.Remove, "Remove from playlist", {
                onRemoveFromPlaylist()
                dismiss()
            })
        }
    }
}

private suspend fun startRadio(core: OtterCore, player: PlayerController, seed: Track, context: Context) {
    try {
        val radioTracks = core.radio(seed.videoId).filter { track -> track.videoId != seed.videoId }
        player.playTracks(listOf(seed) + radioTracks, ItemKind.Song, 0, "radio")
    } catch (error: Exception) {
        Toast.makeText(context, error.message ?: "Could not start radio", Toast.LENGTH_SHORT).show()
    }
}

@Composable
private fun TrackSheetHeader(track: Track, kind: ItemKind, modifier: Modifier = Modifier) {
    val library = LocalAppContainer.current.libraryState
    val likedVideoIds by library.likedVideoIds.collectAsStateWithLifecycle()
    LaunchedEffect(track.videoId) { library.ensureRatings(listOf(track.videoId)) }
    val isLiked = track.videoId in likedVideoIds

    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Cover(track.thumbnails, size = 52.dp)
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = track.title,
                style = MaterialTheme.typography.titleSmall,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = trackSubtitle(track),
                style = MaterialTheme.typography.bodySmall,
                color = OtterTheme.colors.textDim,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        IconButton(onClick = { library.toggleLike(track, kind) }) {
            Icon(
                imageVector = OtterIcons.Heart.select(isLiked),
                contentDescription = "Like",
                tint = if (isLiked) MaterialTheme.colorScheme.primary else OtterTheme.colors.textMuted,
            )
        }
    }
}

/** "Artist • Album" for track rows and sheets. */
fun trackSubtitle(track: Track): String {
    val albumName = track.album?.name.orEmpty()
    if (albumName.isEmpty()) {
        return track.artistNames
    }
    if (track.artists.isEmpty()) {
        return albumName
    }
    return "${track.artistNames} • $albumName"
}

@Composable
private fun ArtistPickerDialog(artists: List<Artist>, onPick: (Artist) -> Unit, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Go to artist") },
        text = {
            Column {
                for (artist in artists) {
                    Text(
                        text = artist.name,
                        style = MaterialTheme.typography.bodyLarge,
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(12.dp))
                            .clickable { onPick(artist) }
                            .padding(vertical = 12.dp, horizontal = 8.dp),
                    )
                }
            }
        },
        confirmButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}

/** Lists the user's playlists plus "New playlist"; picking one adds [tracks] to it. */
@Composable
fun PlaylistPickerSheet(tracks: List<Track>, onDismiss: () -> Unit) {
    val library = LocalAppContainer.current.libraryState
    val playlists by library.playlists.collectAsStateWithLifecycle()
    val context = LocalContext.current
    // Outlives the sheet or dialog, which closes right after starting the work.
    val scope = LocalAppContainer.current.applicationScope
    var showNewPlaylistDialog by remember { mutableStateOf(false) }

    fun addTo(playlist: PlaylistSummary, dismiss: () -> Unit) {
        scope.launch {
            try {
                library.addToPlaylist(playlist.id, tracks)
                Toast.makeText(context, "Saved to ${playlist.name}", Toast.LENGTH_SHORT).show()
            } catch (error: Exception) {
                Toast.makeText(context, error.message ?: "Could not save", Toast.LENGTH_SHORT).show()
            }
        }
        dismiss()
    }

    FloatingSheet(onDismiss = onDismiss) { dismiss ->
        Text(
            text = "Save to playlist",
            style = MaterialTheme.typography.titleSmall,
            modifier = Modifier.padding(horizontal = 20.dp, vertical = 8.dp),
        )
        SheetDivider()
        ActionRow(OtterIcons.Add, "New playlist", { showNewPlaylistDialog = true })
        LazyColumn(modifier = Modifier.height(280.dp)) {
            items(playlists, key = { playlist -> playlist.id }) { playlist ->
                PlaylistPickerRow(playlist = playlist, onClick = { addTo(playlist, dismiss) })
            }
        }
    }

    if (showNewPlaylistDialog) {
        NewPlaylistDialog(
            onDismiss = { showNewPlaylistDialog = false },
            onCreate = { name ->
                showNewPlaylistDialog = false
                scope.launch {
                    try {
                        val created = library.createPlaylist(name)
                        library.addToPlaylist(created.id, tracks)
                        Toast.makeText(context, "Saved to $name", Toast.LENGTH_SHORT).show()
                    } catch (error: Exception) {
                        Toast.makeText(context, error.message ?: "Could not save", Toast.LENGTH_SHORT).show()
                    }
                }
                onDismiss()
            },
        )
    }
}

@Composable
private fun PlaylistPickerRow(playlist: PlaylistSummary, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .padding(horizontal = 8.dp)
            .fillMaxWidth()
            .clip(RoundedCornerShape(16.dp))
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Cover(playlist.thumbnails, size = 40.dp, shape = RoundedCornerShape(6.dp))
        Column {
            Text(playlist.name, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(
                text = "${playlist.trackCount} songs",
                style = MaterialTheme.typography.bodySmall,
                color = OtterTheme.colors.textDim,
            )
        }
    }
}

/** Asks for a playlist name; [onCreate] receives the trimmed, non-empty name. */
@Composable
fun NewPlaylistDialog(onDismiss: () -> Unit, onCreate: (String) -> Unit) {
    var name by remember { mutableStateOf("") }
    val focusRequester = remember { FocusRequester() }
    LaunchedEffect(Unit) { focusRequester.requestFocus() }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("New playlist") },
        text = {
            OutlinedTextField(
                value = name,
                onValueChange = { name = it },
                singleLine = true,
                label = { Text("Name") },
                modifier = Modifier.focusRequester(focusRequester),
            )
        },
        confirmButton = {
            TextButton(onClick = { onCreate(name.trim()) }, enabled = name.isNotBlank()) { Text("Create") }
        },
        dismissButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}
