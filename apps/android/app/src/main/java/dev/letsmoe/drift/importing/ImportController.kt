package dev.letsmoe.drift.importing

import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.IMPORT_CANCELLED_MESSAGE
import dev.letsmoe.drift.core.ImportEventListener
import dev.letsmoe.drift.core.model.ImportProgress
import dev.letsmoe.drift.core.model.ImportResult
import dev.letsmoe.drift.core.model.ImportSourceKind
import dev.letsmoe.drift.core.model.ImportSummary
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/**
 * The import flow: choose a source, pick what to import, watch it run, read the summary. It
 * outlives its screen, so an import keeps reporting while the screen is closed. The core runs
 * one import at a time, and so does this.
 */
class ImportController(
    private val core: DriftCore,
    private val fileStore: ImportFileStore,
    private val keepAlive: ImportKeepAlive,
    private val scope: CoroutineScope,
    private val onLibraryChanged: suspend () -> Unit,
) {
    private val mutableState = MutableStateFlow(ImportState())
    val state: StateFlow<ImportState> = mutableState.asStateFlow()

    private var location = ""
    private var copiedFilePath: String? = null

    init {
        scope.launch { fileStore.deleteAll() }
    }

    /** Reads a public Spotify playlist or album link and moves on to choosing what to import. */
    fun previewLink(link: String) {
        val trimmedLink = link.trim()
        if (trimmedLink.isEmpty() || !canChooseSource()) {
            return
        }
        scope.launch {
            location = trimmedLink
            loadPreview(ImportSourceKind.Url) { core.spotifyPreviewUrl(trimmedLink) }
        }
    }

    /** Reads a picked data export or Exportify file; [documentReference] is what [ImportFileStore] resolves. */
    fun previewFile(documentReference: String) {
        if (!canChooseSource()) {
            return
        }
        scope.launch {
            mutableState.update { it.copy(step = ImportStep.Loading, errorMessage = "") }
            val path = copyPickedFile(documentReference) ?: return@launch
            location = path
            copiedFilePath = path
            loadPreview(ImportSourceKind.Export) { core.spotifyPreviewExport(path) }
        }
    }

    fun setIncludeLiked(included: Boolean) = mutableState.update { it.copy(includeLiked = included) }

    fun setIncludeAlbums(included: Boolean) = mutableState.update { it.copy(includeAlbums = included) }

    fun setIncludeArtists(included: Boolean) = mutableState.update { it.copy(includeArtists = included) }

    fun setIncludeShows(included: Boolean) = mutableState.update { it.copy(includeShows = included) }

    fun togglePlaylist(key: String) {
        mutableState.update { current ->
            if (key in current.selectedPlaylistKeys) {
                current.copy(selectedPlaylistKeys = current.selectedPlaylistKeys - key)
            } else {
                current.copy(selectedPlaylistKeys = current.selectedPlaylistKeys + key)
            }
        }
    }

    fun toggleAllPlaylists() {
        mutableState.update { current ->
            val playlists = current.summary?.playlists.orEmpty()
            if (current.selectedPlaylistKeys.size == playlists.size) {
                current.copy(selectedPlaylistKeys = emptySet())
            } else {
                current.copy(selectedPlaylistKeys = playlists.map { it.key }.toSet())
            }
        }
    }

    fun start() {
        val current = mutableState.value
        if (current.step != ImportStep.Preview || !current.canStart) {
            return
        }
        mutableState.update {
            it.copy(step = ImportStep.Running, progress = null, errorMessage = "", wasCancelled = false)
        }
        keepAlive.start()
        scope.launch {
            try {
                core.spotifyImport(current.sourceKind, location, current.toOptions(), eventListener)
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (error: Exception) {
                fail(error.message ?: error.javaClass.simpleName)
            }
        }
    }

    fun cancel() {
        scope.launch { core.cancelImport() }
    }

    /** Back to choosing a source. Does nothing while an import runs. */
    fun restart() {
        if (mutableState.value.step == ImportStep.Running) {
            return
        }
        scope.launch { deleteCopiedFile() }
        mutableState.value = ImportState()
    }

    /** Clears a finished import's summary once it has been read. */
    fun dismissFinished() {
        if (mutableState.value.step == ImportStep.Done) {
            restart()
        }
    }

    private fun canChooseSource(): Boolean {
        val step = mutableState.value.step
        return step == ImportStep.Source || step == ImportStep.Preview
    }

    private suspend fun copyPickedFile(documentReference: String): String? {
        try {
            return fileStore.copyIntoWorkspace(documentReference)
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            returnToSource(error.message ?: error.javaClass.simpleName)
            return null
        }
    }

    private suspend fun loadPreview(sourceKind: ImportSourceKind, load: suspend () -> ImportSummary) {
        mutableState.update { it.copy(step = ImportStep.Loading, errorMessage = "") }
        try {
            val summary = load()
            mutableState.value = ImportState(
                step = ImportStep.Preview,
                sourceKind = sourceKind,
                summary = summary,
                includeLiked = summary.liked > 0,
                includeAlbums = summary.albums > 0,
                includeArtists = summary.artists > 0,
                includeShows = summary.shows > 0,
                selectedPlaylistKeys = summary.playlists.map { it.key }.toSet(),
            )
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (error: Exception) {
            returnToSource(error.message ?: error.javaClass.simpleName)
        }
    }

    private suspend fun returnToSource(message: String) {
        deleteCopiedFile()
        mutableState.value = ImportState(errorMessage = message)
    }

    private suspend fun deleteCopiedFile() {
        val path = copiedFilePath ?: return
        copiedFilePath = null
        fileStore.delete(path)
    }

    private val eventListener = object : ImportEventListener {
        override fun onProgress(progress: ImportProgress) {
            mutableState.update { it.copy(progress = progress) }
        }

        override fun onFinished(result: ImportResult) {
            finish(result)
        }

        override fun onFailed(message: String) {
            fail(message)
        }
    }

    private fun finish(result: ImportResult) {
        keepAlive.stop()
        scope.launch {
            deleteCopiedFile()
            mutableState.update { it.copy(step = ImportStep.Done, result = result, progress = null) }
            onLibraryChanged()
        }
    }

    // A failed or cancelled import goes back to the preview, so it can be retried with the same file.
    private fun fail(message: String) {
        keepAlive.stop()
        if (message == IMPORT_CANCELLED_MESSAGE) {
            mutableState.update { it.copy(step = ImportStep.Preview, progress = null, wasCancelled = true) }
            return
        }
        mutableState.update {
            it.copy(step = ImportStep.Preview, progress = null, wasCancelled = false, errorMessage = message)
        }
    }
}
