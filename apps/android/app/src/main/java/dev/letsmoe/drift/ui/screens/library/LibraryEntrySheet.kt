package dev.letsmoe.drift.ui.screens.library

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.library.LibraryState
import dev.letsmoe.drift.ui.components.ActionRow
import dev.letsmoe.drift.ui.components.ConfirmDialog
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.components.EditPlaylistDialog
import dev.letsmoe.drift.ui.components.FloatingSheet
import dev.letsmoe.drift.ui.components.SheetDivider
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.showToastOnFailure
import dev.letsmoe.drift.ui.theme.DriftTheme
import dev.letsmoe.drift.ui.LocalAppContainer
import kotlinx.coroutines.launch

private enum class EntryDialog { None, Edit, Delete }

/** The long-press sheet of a library row: rename / delete for local playlists, unsave / unfollow for saved items. */
@Composable
fun LibraryEntrySheet(entry: LibraryEntry, libraryState: LibraryState, onDismiss: () -> Unit) {
    var openDialog by remember { mutableStateOf(EntryDialog.None) }
    val context = LocalContext.current
    // Outlives the sheet or dialog, which closes right after starting the work.
    val scope = LocalAppContainer.current.applicationScope

    if (openDialog == EntryDialog.None) {
        FloatingSheet(onDismiss = onDismiss) { dismiss ->
            EntryHeader(entry)
            SheetDivider()
            when (entry) {
                is LibraryEntry.LocalPlaylist -> LocalPlaylistActions(
                    onEdit = { openDialog = EntryDialog.Edit },
                    onDelete = { openDialog = EntryDialog.Delete },
                )
                is LibraryEntry.Saved -> ActionRow(
                    icon = DriftIcons.Remove,
                    label = removeLabel(entry),
                    onClick = {
                        scope.launch { showToastOnFailure(context) { libraryState.toggleSaved(entry.item) } }
                        dismiss()
                    },
                )
            }
        }
    }
    if (entry is LibraryEntry.LocalPlaylist && openDialog == EntryDialog.Edit) {
        EditPlaylistDialog(
            name = entry.summary.name,
            description = entry.summary.description,
            onSave = { name, description ->
                scope.launch {
                    showToastOnFailure(context) { libraryState.renamePlaylist(entry.summary.id, name, description) }
                }
                onDismiss()
            },
            onDismiss = onDismiss,
        )
    }
    if (entry is LibraryEntry.LocalPlaylist && openDialog == EntryDialog.Delete) {
        ConfirmDialog(
            title = "Delete playlist",
            message = "Delete \"${entry.summary.name}\" from your library? This cannot be undone.",
            confirmLabel = "Delete",
            onConfirm = {
                scope.launch { showToastOnFailure(context) { libraryState.deletePlaylist(entry.summary.id) } }
                onDismiss()
            },
            onDismiss = onDismiss,
        )
    }
}

@Composable
private fun LocalPlaylistActions(onEdit: () -> Unit, onDelete: () -> Unit) {
    ActionRow(DriftIcons.Edit, "Edit details", onEdit)
    ActionRow(
        icon = DriftIcons.Delete,
        label = "Delete playlist",
        onClick = onDelete,
        tint = MaterialTheme.colorScheme.error,
    )
}

private fun removeLabel(entry: LibraryEntry.Saved): String {
    val kind = entry.item.kind
    if (kind == ItemKind.Artist || kind == ItemKind.Podcast) {
        return "Unfollow"
    }
    return "Remove from library"
}

@Composable
private fun EntryHeader(entry: LibraryEntry) {
    Row(
        modifier = Modifier.padding(horizontal = 20.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Cover(thumbnails = entry.thumbnails, size = 48.dp, circle = entry.isCircle)
        Column {
            Text(entry.title, style = MaterialTheme.typography.titleSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
            Text(
                entry.subtitle,
                style = MaterialTheme.typography.bodySmall,
                color = DriftTheme.colors.textDim,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}
