package dev.letsmoe.drift.core.model

import kotlinx.serialization.Serializable

/** Where an import reads from; [wireName] is the core's source kind. */
enum class ImportSourceKind(val wireName: String) {
    Export("export"),
    Url("url"),
    YouTubeAccount("youtube-account"),
}

@Serializable
data class ImportPlaylistSummary(
    val key: String,
    val name: String = "",
    val trackCount: Int = 0,
)

/** What a source holds, counted without any matching. */
@Serializable
data class ImportSummary(
    val liked: Int = 0,
    val playlists: List<ImportPlaylistSummary> = emptyList(),
    val albums: Int = 0,
    val artists: Int = 0,
    val shows: Int = 0,
    /** Set when the source shows at most that many tracks per playlist (public links). */
    val trackLimit: Int = 0,
)

/** What an import acts on; [playlists] holds [ImportPlaylistSummary.key]s. */
@Serializable
data class ImportOptions(
    val liked: Boolean = false,
    val playlists: List<String> = emptyList(),
    val albums: Boolean = false,
    val artists: Boolean = false,
    val shows: Boolean = false,
    val retryUnmatched: Boolean = false,
)

@Serializable
data class ImportProgress(
    /** One of "tracks", "albums", "artists", "shows" and "saving". */
    val phase: String = "",
    val done: Int = 0,
    val total: Int = 0,
    val current: String = "",
)

@Serializable
data class ImportCategoryCount(
    val matched: Int = 0,
    val unmatched: Int = 0,
)

@Serializable
data class ImportPlaylistResult(
    val key: String = "",
    val name: String = "",
    val playlistId: String = "",
    val matched: Int = 0,
    val unmatched: Int = 0,
)

@Serializable
data class ImportUnmatchedItem(
    /** One of "track", "album", "artist", "show" and "playlist". */
    val kind: String = "",
    val title: String = "",
    val artists: List<String> = emptyList(),
    /** The playlist a track belongs to; empty for liked tracks and other kinds. */
    val playlist: String = "",
)

@Serializable
data class ImportResult(
    val liked: ImportCategoryCount = ImportCategoryCount(),
    val albums: ImportCategoryCount = ImportCategoryCount(),
    val artists: ImportCategoryCount = ImportCategoryCount(),
    val shows: ImportCategoryCount = ImportCategoryCount(),
    val playlists: List<ImportPlaylistResult> = emptyList(),
    val unmatched: List<ImportUnmatchedItem> = emptyList(),
)

@Serializable
data class YouTubeAccountInfo(
    val name: String = "",
    val handle: String = "",
    val email: String = "",
    val photoUrl: String = "",
)

/** The stored YouTube Music sign-in; times are RFC 3339 and empty when unknown. */
@Serializable
data class YouTubeAccountStatus(
    val signedIn: Boolean = false,
    val account: YouTubeAccountInfo = YouTubeAccountInfo(),
    val validatedAt: String = "",
    val lastSyncedAt: String = "",
)
