package dev.neoworks.ottermusic.ui.navigation

import kotlinx.serialization.Serializable

/** Type-safe navigation destinations. Ids are YouTube browse IDs, except LocalPlaylist (library playlist id). */
object Routes {
    @Serializable
    data object Home

    @Serializable
    data object Search

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
    data class LocalPlaylist(val id: String)
}
