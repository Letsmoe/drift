package dev.letsmoe.drift.ui.screens.library

import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.PlaylistSummary
import dev.letsmoe.drift.core.model.SavedItem
import dev.letsmoe.drift.core.model.Thumbnail

/** The kinds the library chips filter by. */
enum class LibraryFilter(val label: String) {
    Playlists("Playlists"),
    Albums("Albums"),
    Artists("Artists"),
    Podcasts("Podcasts"),
}

enum class LibrarySort(val label: String) {
    Recents("Recents"),
    Alphabetical("Alphabetical"),
}

/** One row of the library list: a local playlist or a saved album, YouTube playlist, artist or podcast. */
sealed interface LibraryEntry {
    val key: String
    val title: String
    val subtitle: String
    val thumbnails: List<Thumbnail>
    val isCircle: Boolean

    /** Queue context that plays this entry, to highlight it while it plays. */
    val playContext: String
    val filter: LibraryFilter

    /** Sort key for "Recents": the core's ISO timestamps order lexicographically. */
    val recencyTimestamp: String

    data class LocalPlaylist(val summary: PlaylistSummary) : LibraryEntry {
        override val key = "library:${summary.id}"
        override val title = summary.name
        override val subtitle = "Playlist • ${summary.trackCount} songs"
        override val thumbnails = summary.thumbnails
        override val isCircle = false
        override val playContext = "library:${summary.id}"
        override val filter = LibraryFilter.Playlists
        override val recencyTimestamp = summary.updatedAt
    }

    data class Saved(val saved: SavedItem) : LibraryEntry {
        val item: Item = saved.item
        override val key = "${item.kind.wireName}:${item.id}"
        override val title = item.title
        override val subtitle = savedSubtitle(item)
        override val thumbnails = item.thumbnails
        override val isCircle = item.kind == ItemKind.Artist
        override val playContext = "${item.kind.wireName}:${item.id}"
        override val filter = filterOf(item.kind)
        override val recencyTimestamp = saved.savedAt
    }
}

private fun kindLabel(kind: ItemKind): String =
    when (kind) {
        ItemKind.Album -> "Album"
        ItemKind.Playlist -> "Playlist"
        ItemKind.Artist -> "Artist"
        ItemKind.Podcast -> "Podcast"
        ItemKind.Song -> "Song"
        ItemKind.Episode -> "Episode"
    }

private fun savedSubtitle(item: Item): String {
    val label = kindLabel(item.kind)
    if (item.subtitle.isEmpty() || item.kind == ItemKind.Artist) {
        return label
    }
    return "$label • ${item.subtitle}"
}

private fun filterOf(kind: ItemKind): LibraryFilter =
    when (kind) {
        ItemKind.Album -> LibraryFilter.Albums
        ItemKind.Artist -> LibraryFilter.Artists
        ItemKind.Podcast -> LibraryFilter.Podcasts
        ItemKind.Playlist, ItemKind.Song, ItemKind.Episode -> LibraryFilter.Playlists
    }
