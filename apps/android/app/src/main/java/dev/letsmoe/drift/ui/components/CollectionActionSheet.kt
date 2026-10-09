package dev.letsmoe.drift.ui.components

import android.widget.Toast
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.foundation.layout.padding
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme

/** The "more" sheet of an album, playlist or artist: play next, add to queue, save to playlist. */
@Composable
fun CollectionActionSheet(
    title: String,
    tracks: List<Track>,
    onDismiss: () -> Unit,
) {
    val playerController = LocalAppContainer.current.playerController
    val context = LocalContext.current
    var showPlaylistPicker by remember { mutableStateOf(false) }

    FloatingSheet(onDismiss = onDismiss) { dismiss ->
        Text(
            text = title,
            style = MaterialTheme.typography.titleSmall,
            maxLines = 1,
            modifier = Modifier.padding(horizontal = 20.dp, vertical = 8.dp),
        )
        SheetDivider()
        ActionRow(DriftIcons.SkipForward, "Play next", {
            playerController.playNext(tracks)
            Toast.makeText(context, "Playing next", Toast.LENGTH_SHORT).show()
            dismiss()
        })
        ActionRow(DriftIcons.Queue, "Add to queue", {
            playerController.addToQueue(tracks)
            Toast.makeText(context, "Added to queue", Toast.LENGTH_SHORT).show()
            dismiss()
        })
        ActionRow(DriftIcons.AddToPlaylist, "Save to playlist", { showPlaylistPicker = true })
    }

    if (showPlaylistPicker) {
        PlaylistPickerSheet(
            tracks = tracks,
            onDismiss = {
                showPlaylistPicker = false
                onDismiss()
            },
        )
    }
}

/** A "more" icon button that opens a [CollectionActionSheet] for [tracks]. */
@Composable
fun CollectionMoreButton(title: String, tracks: List<Track>, modifier: Modifier = Modifier) {
    var showSheet by remember { mutableStateOf(false) }
    IconButton(onClick = { showSheet = true }, modifier = modifier, enabled = tracks.isNotEmpty()) {
        Icon(DriftIcons.More.outline, contentDescription = "More", tint = DriftTheme.colors.textMuted)
    }
    if (showSheet) {
        CollectionActionSheet(title = title, tracks = tracks, onDismiss = { showSheet = false })
    }
}
