package dev.neoworks.ottermusic.ui.screens.album

import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.Album
import dev.neoworks.ottermusic.core.model.AlbumPage
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.screens.common.RemoteViewModel

class AlbumViewModel(core: OtterCore, private val albumId: String) :
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
