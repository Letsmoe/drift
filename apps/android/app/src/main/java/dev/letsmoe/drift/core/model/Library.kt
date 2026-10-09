package dev.letsmoe.drift.core.model

import kotlinx.serialization.Serializable

@Serializable
data class PlaylistSummary(
    val id: String,
    val name: String = "",
    val description: String = "",
    val trackCount: Int = 0,
    val thumbnails: List<Thumbnail> = emptyList(),
    val updatedAt: String = "",
)

@Serializable
data class PlaylistEntry(
    val id: Long,
    val track: Track,
    val addedAt: String = "",
)

@Serializable
data class LibraryPlaylist(
    val id: String,
    val name: String = "",
    val description: String = "",
    val createdAt: String = "",
    val updatedAt: String = "",
    val entries: List<PlaylistEntry> = emptyList(),
)

@Serializable
data class SavedItem(
    val item: Item,
    val savedAt: String = "",
)

@Serializable
data class Impression(
    val videoId: String,
    /** What produced the recommendation, e.g. "related:radio". */
    val source: String,
    /** Where it was shown, e.g. "home". */
    val surface: String,
    val position: Int? = null,
    val clicked: Boolean = false,
)

@Serializable
data class ResumePosition(
    val seconds: Int = 0,
    val found: Boolean = false,
)
