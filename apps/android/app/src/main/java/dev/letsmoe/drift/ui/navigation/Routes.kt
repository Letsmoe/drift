package dev.letsmoe.drift.ui.navigation

import kotlinx.serialization.Serializable

/** Type-safe navigation destinations. Ids are YouTube browse IDs, except LocalPlaylist (library playlist id). */
object Routes {
    @Serializable
    data object Home

    /** Search as a page; [podcasts] searches shows and episodes instead of music. */
    @Serializable
    data class Search(val podcasts: Boolean = false)

    @Serializable
    data object Podcasts

    @Serializable
    data object Library

    @Serializable
    data class Album(val id: String)

    @Serializable
    data class Playlist(val id: String)

    @Serializable
    data class Artist(val id: String)

    @Serializable
    data class Podcast(val id: String)

    @Serializable
    data object LikedSongs

    @Serializable
    data object Import

    @Serializable
    data class LocalPlaylist(val id: String)

    @Serializable
    data object Downloads

    @Serializable
    data object Settings
}
