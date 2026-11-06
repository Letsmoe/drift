package dev.letsmoe.drift.importing

import dev.letsmoe.drift.core.model.ImportOptions
import dev.letsmoe.drift.core.model.ImportProgress
import dev.letsmoe.drift.core.model.ImportResult
import dev.letsmoe.drift.core.model.ImportSourceKind
import dev.letsmoe.drift.core.model.ImportSummary

enum class ImportStep {
    Source,
    Loading,
    Preview,
    Running,
    Done,
}

/** Everything the import screens show. */
data class ImportState(
    val step: ImportStep = ImportStep.Source,
    val sourceKind: ImportSourceKind = ImportSourceKind.Url,
    val summary: ImportSummary? = null,
    val includeLiked: Boolean = false,
    val includeAlbums: Boolean = false,
    val includeArtists: Boolean = false,
    val includeShows: Boolean = false,
    val selectedPlaylistKeys: Set<String> = emptySet(),
    val progress: ImportProgress? = null,
    val result: ImportResult? = null,
    val errorMessage: String = "",
    val wasCancelled: Boolean = false,
) {
    /** True while an import runs or its summary has not been read yet. */
    val hasBackgroundStatus: Boolean
        get() = step == ImportStep.Running || step == ImportStep.Done

    val isAccountSource: Boolean
        get() = sourceKind == ImportSourceKind.YouTubeAccount

    val selectedItemCount: Int
        get() {
            val loadedSummary = summary ?: return 0
            var count = selectedPlaylistKeys.size
            count += flagCount(includeLiked, loadedSummary.liked)
            count += flagCount(includeAlbums, loadedSummary.albums)
            count += flagCount(includeArtists, loadedSummary.artists)
            count += flagCount(includeShows, loadedSummary.shows)
            return count
        }

    val canStart: Boolean
        get() = selectedItemCount > 0

    /** 0 to 1 share of the items handled so far. */
    val progressFraction: Float
        get() {
            val current = progress ?: return 0f
            if (current.total == 0) {
                return 0f
            }
            return current.done.toFloat() / current.total
        }

    /** What the running import does right now, e.g. "Matching songs". */
    val phaseLabel: String
        get() {
            val current = progress ?: return "Starting"
            var labels = MATCHING_LABELS
            if (isAccountSource) {
                labels = READING_LABELS
            }
            return labels[current.phase] ?: current.phase
        }

    fun toOptions(): ImportOptions =
        ImportOptions(
            liked = includeLiked,
            playlists = selectedPlaylistKeys.toList(),
            albums = includeAlbums,
            artists = includeArtists,
            shows = includeShows,
        )

    private fun flagCount(included: Boolean, count: Int): Int {
        if (included && count > 0) {
            return 1
        }
        return 0
    }

    private companion object {
        val MATCHING_LABELS = mapOf(
            "tracks" to "Matching songs",
            "albums" to "Matching albums",
            "artists" to "Matching artists",
            "shows" to "Matching podcasts",
            "saving" to "Saving to your library",
        )
        val READING_LABELS = mapOf(
            "tracks" to "Reading songs and playlists",
            "albums" to "Reading albums",
            "artists" to "Reading artists",
            "shows" to "Reading podcasts",
            "saving" to "Saving to your library",
        )
    }
}
