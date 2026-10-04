package dev.neoworks.ottermusic.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Thumbnail(
    val url: String = "",
    val width: Int = 0,
    val height: Int = 0,
)

@Serializable
data class Artist(
    val id: String = "",
    val name: String = "",
)

@Serializable
data class Album(
    val id: String = "",
    val name: String = "",
)

@Serializable
data class Track(
    val videoId: String,
    val title: String = "",
    val artists: List<Artist> = emptyList(),
    val album: Album? = null,
    val durationSeconds: Int = 0,
    val thumbnails: List<Thumbnail> = emptyList(),
) {
    /** "Artist, Artist" for subtitles. */
    val artistNames: String
        get() = artists.joinToString(", ") { artist -> artist.name }
}

/** What a shelf item is: songs and episodes play, the others open a page. */
@Serializable
enum class ItemKind(val wireName: String) {
    @SerialName("song")
    Song("song"),

    @SerialName("album")
    Album("album"),

    @SerialName("playlist")
    Playlist("playlist"),

    @SerialName("artist")
    Artist("artist"),

    @SerialName("podcast")
    Podcast("podcast"),

    @SerialName("episode")
    Episode("episode"),
}

@Serializable
data class Item(
    val kind: ItemKind = ItemKind.Song,
    /** Video ID for songs and episodes, browse ID otherwise. */
    val id: String = "",
    val title: String = "",
    val subtitle: String = "",
    val thumbnails: List<Thumbnail> = emptyList(),
    /** Set for the playable kinds (song, episode). */
    val track: Track? = null,
)

@Serializable
data class Shelf(
    val title: String = "",
    val items: List<Item> = emptyList(),
)

@Serializable
data class HomePage(
    val shelves: List<Shelf> = emptyList(),
    /** Empty on the last page. */
    val continuation: String = "",
)

@Serializable
data class SearchResults(
    val shelves: List<Shelf> = emptyList(),
)

@Serializable
data class AlbumPage(
    val id: String = "",
    val playlistId: String = "",
    val title: String = "",
    val subtitle: String = "",
    val artists: List<Artist> = emptyList(),
    val year: String = "",
    val thumbnails: List<Thumbnail> = emptyList(),
    val tracks: List<Track> = emptyList(),
)

@Serializable
data class PlaylistPage(
    val id: String = "",
    val title: String = "",
    val author: String = "",
    val description: String = "",
    val thumbnails: List<Thumbnail> = emptyList(),
    val tracks: List<Track> = emptyList(),
    val continuation: String = "",
)

@Serializable
data class PlaylistTracks(
    val tracks: List<Track> = emptyList(),
    val continuation: String = "",
)

@Serializable
data class ArtistPage(
    val id: String = "",
    val name: String = "",
    val description: String = "",
    val thumbnails: List<Thumbnail> = emptyList(),
    val topSongs: List<Track> = emptyList(),
    val shelves: List<Shelf> = emptyList(),
)

@Serializable
data class Episode(
    val track: Track,
    val description: String = "",
    /** YouTube's localized date text, e.g. "3 days ago". */
    val published: String = "",
)

@Serializable
data class PodcastPage(
    val id: String = "",
    val title: String = "",
    val author: Artist = Artist(),
    val description: String = "",
    val thumbnails: List<Thumbnail> = emptyList(),
    val episodes: List<Episode> = emptyList(),
    val continuation: String = "",
)

@Serializable
data class PodcastEpisodes(
    val episodes: List<Episode> = emptyList(),
    val continuation: String = "",
)
