package dev.letsmoe.drift.ui.components

import android.widget.Toast
import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.CollectionKind
import dev.letsmoe.drift.core.model.DownloadCollection
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.showToastOnFailure
import dev.letsmoe.drift.ui.theme.DriftTheme
import kotlinx.coroutines.launch

/**
 * Keeps an album, playlist, local playlist or Liked Songs downloaded: first tap downloads it and
 * keeps it in sync, a tap on a downloaded one asks before removing the downloads.
 */
@Composable
fun CollectionDownloadButton(kind: CollectionKind, id: String, title: String, modifier: Modifier = Modifier) {
    val downloads = LocalAppContainer.current.downloadsState
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val collections by downloads.collections.collectAsStateWithLifecycle()
    val collection = collections.firstOrNull { candidate -> candidate.kind == kind && candidate.id == id }
    var confirmingRemoval by remember { mutableStateOf(false) }

    // A kept YouTube collection looks for new tracks whenever its page opens.
    LaunchedEffect(kind, id) { downloads.refreshOnOpen(kind, id) }

    IconButton(
        modifier = modifier,
        onClick = {
            if (collection == null) {
                scope.launch {
                    showToastOnFailure(context) {
                        downloads.keepDownloaded(kind, id)
                        Toast.makeText(context, "Downloading $title", Toast.LENGTH_SHORT).show()
                    }
                }
                return@IconButton
            }
            confirmingRemoval = true
        },
    ) {
        CollectionDownloadIcon(collection = collection, title = title)
    }

    if (confirmingRemoval) {
        ConfirmDialog(
            title = "Remove downloads",
            message = "Delete the downloaded songs of \"$title\" and stop keeping it downloaded? " +
                "Songs that another downloaded collection or a single download still needs stay.",
            confirmLabel = "Remove",
            onConfirm = {
                confirmingRemoval = false
                scope.launch { showToastOnFailure(context) { downloads.removeCollection(kind, id) } }
            },
            onDismiss = { confirmingRemoval = false },
        )
    }
}

@Composable
private fun CollectionDownloadIcon(collection: DownloadCollection?, title: String) {
    if (collection == null) {
        Icon(
            DriftIcons.Download.outline,
            contentDescription = "Keep $title downloaded",
            tint = DriftTheme.colors.textMuted,
        )
        return
    }
    if (collection.isComplete) {
        Icon(
            DriftIcons.Download.filled,
            contentDescription = "Remove downloads of $title",
            tint = MaterialTheme.colorScheme.primary,
        )
        return
    }
    CircularProgressIndicator(
        progress = { collection.fraction },
        modifier = Modifier.size(24.dp),
        strokeWidth = 2.5.dp,
    )
}
