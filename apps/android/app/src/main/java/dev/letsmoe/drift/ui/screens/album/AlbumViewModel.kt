package dev.letsmoe.drift.ui.screens.album

import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.Album
import dev.letsmoe.drift.core.model.AlbumPage
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.screens.common.RemoteViewModel

class AlbumViewModel(core: DriftCore, private val albumId: String) :
    RemoteViewModel<AlbumPage>({ withPlayableTracks(core.album(albumId), albumId) })

// Album rows carry no cover or album of their own, and no artists when they equal the album's; the queue needs all three.
private fun withPlayableTracks(page: AlbumPage, albumId: String): AlbumPage {
    val album = Album(id = albumId, name = page.title)
    val playableTracks = page.tracks.map { track -> playable(track, page, album) }
    return page.copy(tracks = playableTracks)
}

private fun playable(track: Track, page: AlbumPage, album: Album): Track {
    var artists = track.artists
    if (artists.isEmpty()) {
        artists = page.artists
    }
    return track.copy(thumbnails = page.thumbnails, album = album, artists = artists)
}
