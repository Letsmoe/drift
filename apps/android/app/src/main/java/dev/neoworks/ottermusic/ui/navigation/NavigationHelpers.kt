package dev.neoworks.ottermusic.ui.navigation

import androidx.navigation.NavController
import androidx.navigation.NavGraph.Companion.findStartDestination
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.core.model.ItemKind

fun NavController.openAlbum(browseId: String) {
    navigate(Routes.Album(browseId))
}

fun NavController.openPlaylist(browseId: String) {
    navigate(Routes.Playlist(browseId))
}

fun NavController.openArtist(browseId: String) {
    navigate(Routes.Artist(browseId))
}

fun NavController.openPodcast(browseId: String) {
    navigate(Routes.Podcast(browseId))
}

fun NavController.openLikedSongs() {
    navigate(Routes.LikedSongs)
}

fun NavController.openLocalPlaylist(playlistId: String) {
    navigate(Routes.LocalPlaylist(playlistId))
}

/**
 * Opens the page behind a shelf item. Songs and episodes have no page: they are handed to
 * [onPlay] instead (typically `playerController.playTracks(...)`).
 */
fun NavController.openItem(item: Item, onPlay: (Item) -> Unit) {
    when (item.kind) {
        ItemKind.Song, ItemKind.Episode -> onPlay(item)
        ItemKind.Album -> openAlbum(item.id)
        ItemKind.Playlist -> openPlaylist(item.id)
        ItemKind.Artist -> openArtist(item.id)
        ItemKind.Podcast -> openPodcast(item.id)
    }
}

/** Switches the bottom tab, keeping each tab's back stack. */
fun NavController.switchToTab(route: Any) {
    navigate(route) {
        popUpTo(graph.findStartDestination().id) { saveState = true }
        launchSingleTop = true
        restoreState = true
    }
}
