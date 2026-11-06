package dev.letsmoe.drift.ui.components

import android.widget.Toast
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.platform.LocalContext
import dev.letsmoe.drift.core.model.DownloadState
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.downloads.DownloadsState
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.showToastOnFailure
import kotlinx.coroutines.launch

/**
 * The download entry of the track sheet: download, cancel while it is on its way, retry after a
 * failure and remove once it is on the device. Closes the sheet through [dismiss].
 */
@Composable
fun DownloadActionRow(track: Track, kind: ItemKind, dismiss: () -> Unit) {
    val container = LocalAppContainer.current
    val context = LocalContext.current
    val state by rememberDownloadState(track.videoId)
    // Outlives the sheet, which closes right after starting the work.
    val scope = container.applicationScope
    val downloads = container.downloadsState

    fun startAction(toastMessage: String, action: suspend (DownloadsState) -> Unit) {
        scope.launch {
            showToastOnFailure(context) {
                action(downloads)
                Toast.makeText(context, toastMessage, Toast.LENGTH_SHORT).show()
            }
        }
        dismiss()
    }

    when (state) {
        DownloadState.None ->
            ActionRow(DriftIcons.Download, "Download", { startAction("Downloading") { it.download(track, kind) } })
        DownloadState.Done ->
            ActionRow(DriftIcons.Delete, "Remove download", { startAction("Download removed") { it.remove(track.videoId) } })
        DownloadState.Failed ->
            ActionRow(DriftIcons.Download, "Retry download", { startAction("Downloading") { it.retry(track.videoId) } })
        DownloadState.Queued, DownloadState.Downloading, DownloadState.Paused ->
            ActionRow(DriftIcons.Close, "Cancel download", { startAction("Download cancelled") { it.cancel(track.videoId) } })
    }
}
