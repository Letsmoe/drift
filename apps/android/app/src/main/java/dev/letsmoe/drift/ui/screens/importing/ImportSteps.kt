package dev.letsmoe.drift.ui.screens.importing

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.core.model.ImportCategoryCount
import dev.letsmoe.drift.core.model.ImportPlaylistResult
import dev.letsmoe.drift.core.model.ImportUnmatchedItem
import dev.letsmoe.drift.importing.ImportController
import dev.letsmoe.drift.importing.ImportState
import dev.letsmoe.drift.ui.theme.DriftTheme

private const val SOURCE_HINT =
    "Bring a Spotify library over. Everything is matched to YouTube Music; no Spotify account is needed."

internal fun LazyListScope.sourceStep(
    errorMessage: String,
    onPreviewLink: (String) -> Unit,
    onChooseFile: () -> Unit,
) {
    item(key = "source") {
        SourceStep(errorMessage = errorMessage, onPreviewLink = onPreviewLink, onChooseFile = onChooseFile)
    }
}

@Composable
private fun SourceStep(errorMessage: String, onPreviewLink: (String) -> Unit, onChooseFile: () -> Unit) {
    var link by rememberSaveable { mutableStateOf("") }
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        BodyText(SOURCE_HINT)
        OutlinedTextField(
            value = link,
            onValueChange = { link = it },
            singleLine = true,
            label = { Text("Playlist or album link") },
            placeholder = { Text("https://open.spotify.com/playlist/…") },
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Go),
            keyboardActions = KeyboardActions(onGo = { onPreviewLink(link) }),
            modifier = Modifier.fillMaxWidth(),
        )
        DimText("Public links show at most 100 tracks per playlist.")
        Button(onClick = { onPreviewLink(link) }, enabled = link.isNotBlank(), modifier = Modifier.fillMaxWidth()) {
            Text("Preview link")
        }
        Spacer(Modifier.height(8.dp))
        BodyText("For a whole library, pick a Spotify data export (zip or JSON) or an Exportify file (zip or CSV).")
        OutlinedButton(onClick = onChooseFile, modifier = Modifier.fillMaxWidth()) { Text("Choose a file") }
        ErrorText(errorMessage)
    }
}

internal fun LazyListScope.loadingStep() {
    item(key = "loading") {
        Box(modifier = Modifier.fillMaxWidth().padding(vertical = 64.dp), contentAlignment = Alignment.Center) {
            CircularProgressIndicator()
        }
    }
}

internal fun LazyListScope.previewStep(
    state: ImportState,
    controller: ImportController,
    onBack: () -> Unit,
    onImport: () -> Unit,
) {
    val summary = state.summary ?: return
    item(key = "preview-hint") {
        BodyText("Choose what to bring over. Everything is matched to YouTube Music.")
    }
    item(key = "count-liked") {
        CountRow("Liked Songs", summary.liked, state.includeLiked, controller::setIncludeLiked)
    }
    item(key = "count-albums") {
        CountRow("Albums", summary.albums, state.includeAlbums, controller::setIncludeAlbums)
    }
    item(key = "count-artists") {
        CountRow("Artists", summary.artists, state.includeArtists, controller::setIncludeArtists)
    }
    item(key = "count-shows") {
        CountRow("Podcasts", summary.shows, state.includeShows, controller::setIncludeShows)
    }
    if (summary.playlists.isNotEmpty()) {
        item(key = "playlists-header") {
            val allSelected = state.selectedPlaylistKeys.size == summary.playlists.size
            CheckRow(
                label = "Playlists (${summary.playlists.size})",
                checked = allSelected,
                onToggle = { controller.toggleAllPlaylists() },
                isBold = true,
            )
        }
        items(summary.playlists, key = { playlist -> "playlist-${playlist.key}" }) { playlist ->
            CheckRow(
                label = playlist.name,
                detail = trackCountLabel(playlist.trackCount),
                checked = playlist.key in state.selectedPlaylistKeys,
                onToggle = { controller.togglePlaylist(playlist.key) },
                indent = 24,
            )
        }
    }
    item(key = "preview-notes") {
        Column(verticalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.padding(top = 8.dp)) {
            if (summary.trackLimit > 0) {
                DimText("Public links show at most ${summary.trackLimit} tracks per playlist; use the data export for the full playlist.")
            }
            if (state.wasCancelled) {
                BodyText("Import stopped. Tracks matched so far are remembered, so starting again continues quickly.")
            }
            ErrorText(state.errorMessage)
        }
    }
    item(key = "preview-actions") {
        Row(
            modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.End),
        ) {
            TextButton(onClick = onBack) { Text("Back") }
            Button(onClick = onImport, enabled = state.canStart) { Text("Import") }
        }
    }
}

internal fun LazyListScope.runningStep(state: ImportState, onCancel: () -> Unit) {
    item(key = "running") {
        Column(verticalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.padding(top = 16.dp)) {
            Text(state.phaseLabel, style = MaterialTheme.typography.titleMedium)
            ProgressBar(state)
            val progress = state.progress
            if (progress != null && progress.current.isNotBlank()) {
                DimText(progress.current)
            }
            DimText("You can leave this screen. The import keeps running and shows its progress in a notification.")
            OutlinedButton(onClick = onCancel) { Text("Cancel import") }
        }
    }
}

@Composable
private fun ProgressBar(state: ImportState) {
    val progress = state.progress
    if (progress == null || progress.total == 0) {
        LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
        return
    }
    LinearProgressIndicator(progress = { state.progressFraction }, modifier = Modifier.fillMaxWidth())
    DimText("${progress.done} of ${progress.total}")
}

internal fun LazyListScope.doneStep(
    state: ImportState,
    onOpenPlaylist: (String) -> Unit,
    onDone: () -> Unit,
) {
    val result = state.result ?: return
    item(key = "done-title") {
        Text("Import finished", style = MaterialTheme.typography.titleLarge, modifier = Modifier.padding(top = 8.dp))
    }
    val categoryRows = listOf(
        "Liked Songs" to result.liked,
        "Albums" to result.albums,
        "Artists" to result.artists,
        "Podcasts" to result.shows,
    )
    items(categoryRows.filter { (_, count) -> count.matched + count.unmatched > 0 }, key = { (label, _) -> "result-$label" }) {
            (label, count) ->
        ResultRow(label = label, detail = categoryDetail(count), onClick = null)
    }
    if (result.playlists.isNotEmpty()) {
        item(key = "result-playlists") { SectionTitle("Playlists") }
        items(result.playlists, key = { playlist -> "result-playlist-${playlist.key}" }) { playlist ->
            ResultRow(
                label = playlist.name,
                detail = playlistDetail(playlist),
                onClick = playlistClickHandler(playlist, onOpenPlaylist),
            )
        }
    }
    if (result.unmatched.isNotEmpty()) {
        item(key = "unmatched-title") { SectionTitle("Not found on YouTube Music (${result.unmatched.size})") }
        items(result.unmatched.size, key = { index -> "unmatched-$index" }) { index ->
            UnmatchedRow(result.unmatched[index])
        }
    }
    item(key = "done-action") {
        Button(onClick = onDone, modifier = Modifier.fillMaxWidth().padding(top = 16.dp)) { Text("Done") }
    }
}

private fun playlistClickHandler(playlist: ImportPlaylistResult, onOpen: (String) -> Unit): (() -> Unit)? {
    if (playlist.playlistId.isEmpty()) {
        return null
    }
    return { onOpen(playlist.playlistId) }
}

private fun categoryDetail(count: ImportCategoryCount): String {
    if (count.unmatched == 0) {
        return "${count.matched} matched"
    }
    return "${count.matched} matched, ${count.unmatched} not found"
}

private fun playlistDetail(playlist: ImportPlaylistResult): String {
    if (playlist.unmatched == 0) {
        return "${playlist.matched} songs"
    }
    return "${playlist.matched} songs, ${playlist.unmatched} not found"
}

private fun trackCountLabel(count: Int): String {
    if (count <= 0) {
        return ""
    }
    return "$count tracks"
}

@Composable
private fun CountRow(label: String, count: Int, checked: Boolean, onChange: (Boolean) -> Unit) {
    CheckRow(label = label, detail = count.toString(), checked = checked && count > 0, enabled = count > 0, onToggle = { onChange(!checked) })
}

@Composable
private fun CheckRow(
    label: String,
    checked: Boolean,
    onToggle: () -> Unit,
    detail: String = "",
    enabled: Boolean = true,
    isBold: Boolean = false,
    indent: Int = 0,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(enabled = enabled, onClick = onToggle)
            .padding(start = indent.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Checkbox(checked = checked, onCheckedChange = { onToggle() }, enabled = enabled)
        var style = MaterialTheme.typography.bodyLarge
        if (isBold) {
            style = MaterialTheme.typography.titleSmall
        }
        Text(label, style = style, maxLines = 1, overflow = TextOverflow.Ellipsis, modifier = Modifier.weight(1f))
        if (detail.isNotEmpty()) {
            Text(detail, style = MaterialTheme.typography.bodyMedium, color = DriftTheme.colors.textDim)
        }
    }
}

@Composable
private fun ResultRow(label: String, detail: String, onClick: (() -> Unit)?) {
    var modifier = Modifier.fillMaxWidth()
    if (onClick != null) {
        modifier = modifier.clickable(onClick = onClick)
    }
    Column(modifier = modifier.padding(vertical = 10.dp)) {
        Text(label, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
        DimText(detail)
    }
}

@Composable
private fun UnmatchedRow(item: ImportUnmatchedItem) {
    val subtitle = (item.artists + item.playlist).filter { it.isNotBlank() }.joinToString(" • ")
    Column(modifier = Modifier.fillMaxWidth().padding(vertical = 6.dp)) {
        Text(item.title, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
        if (subtitle.isNotEmpty()) {
            DimText(subtitle)
        }
    }
}

@Composable
private fun SectionTitle(text: String) {
    Text(text, style = MaterialTheme.typography.titleSmall, modifier = Modifier.padding(top = 20.dp, bottom = 4.dp))
}

@Composable
private fun BodyText(text: String) {
    Text(text, style = MaterialTheme.typography.bodyMedium, color = DriftTheme.colors.textMuted)
}

@Composable
private fun DimText(text: String) {
    Text(text, style = MaterialTheme.typography.bodySmall, color = DriftTheme.colors.textDim)
}

@Composable
private fun ErrorText(message: String) {
    if (message.isEmpty()) {
        return
    }
    Text(message, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.error)
}
