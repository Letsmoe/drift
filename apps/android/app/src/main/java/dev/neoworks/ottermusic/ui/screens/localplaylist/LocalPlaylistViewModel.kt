package dev.neoworks.ottermusic.ui.screens.localplaylist

import androidx.lifecycle.viewModelScope
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.LibraryPlaylist
import dev.neoworks.ottermusic.core.model.PlaylistEntry
import dev.neoworks.ottermusic.library.LibraryState
import dev.neoworks.ottermusic.ui.describeError
import dev.neoworks.ottermusic.ui.screens.common.RemoteViewModel
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/** An entry that was just removed, with where it was, so the removal can be undone. */
data class RemovedEntry(val entry: PlaylistEntry, val position: Int)

class LocalPlaylistViewModel(
    private val core: OtterCore,
    val libraryState: LibraryState,
    private val playlistId: String,
) : RemoteViewModel<LibraryPlaylist>({ core.libraryPlaylist(playlistId) }) {
    private val mutableEntries = MutableStateFlow<List<PlaylistEntry>>(emptyList())

    /** The entries as shown: reordering and removal apply here at once and are sent to the core after. */
    val entries: StateFlow<List<PlaylistEntry>> = mutableEntries.asStateFlow()

    private val mutableFailures = MutableSharedFlow<String>(extraBufferCapacity = 1)
    val failures: SharedFlow<String> = mutableFailures.asSharedFlow()

    override fun onLoaded(data: LibraryPlaylist) {
        mutableEntries.value = data.entries
    }

    /** Moves the entry with [fromId] to where the entry with [toId] is. */
    fun moveLocally(fromId: Long, toId: Long) {
        val reordered = mutableEntries.value.toMutableList()
        val fromPosition = reordered.indexOfFirst { entry -> entry.id == fromId }
        val toPosition = reordered.indexOfFirst { entry -> entry.id == toId }
        if (fromPosition < 0 || toPosition < 0) {
            return
        }
        reordered.add(toPosition, reordered.removeAt(fromPosition))
        mutableEntries.value = reordered
    }

    /** Sends the entry's current position to the core after a drag ended. */
    fun commitMove(entryId: Long) {
        val newIndex = mutableEntries.value.indexOfFirst { entry -> entry.id == entryId }
        if (newIndex < 0) {
            return
        }
        runCommand { core.movePlaylistEntry(playlistId, entryId, newIndex) }
    }

    fun removeEntry(entry: PlaylistEntry): RemovedEntry? {
        val position = mutableEntries.value.indexOfFirst { candidate -> candidate.id == entry.id }
        if (position < 0) {
            return null
        }
        mutableEntries.value = mutableEntries.value.filter { candidate -> candidate.id != entry.id }
        runCommand { libraryState.removeFromPlaylist(playlistId, entry.id) }
        return RemovedEntry(entry, position)
    }

    /** Adds the track back (the core appends it) and moves it to its old position. */
    fun undoRemoval(removed: RemovedEntry) {
        runCommand(refreshOnSuccess = true) {
            libraryState.addToPlaylist(playlistId, listOf(removed.entry.track))
            val appendedEntry = core.libraryPlaylist(playlistId).entries.last()
            core.movePlaylistEntry(playlistId, appendedEntry.id, removed.position)
        }
    }

    suspend fun saveDetails(name: String, description: String) {
        libraryState.renamePlaylist(playlistId, name, description)
        refreshQuietly()
    }

    suspend fun delete() {
        libraryState.deletePlaylist(playlistId)
    }

    // A failed command leaves the screen out of sync, so it reloads the truth from the core.
    private fun runCommand(refreshOnSuccess: Boolean = false, command: suspend () -> Unit) {
        viewModelScope.launch {
            try {
                command()
                if (refreshOnSuccess) {
                    refreshQuietly()
                }
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (error: Exception) {
                mutableFailures.tryEmit(describeError(error))
                refreshQuietly()
            }
        }
    }
}
