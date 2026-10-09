package dev.letsmoe.drift.ui.player

/** Human readable name of the context a queue was started from (`album:<id>`, `radio`, ...). */
fun contextLabel(context: String): String {
    val kind = context.substringBefore(':')
    return when (kind) {
        "" -> "Your queue"
        "home" -> "Home"
        "recommend" -> "Recommended for you"
        "radio" -> "Radio"
        "liked" -> "Liked songs"
        "album" -> "Album"
        "playlist", "library" -> "Playlist"
        "artist" -> "Artist"
        "podcast" -> "Podcast"
        "search" -> "Search"
        "track" -> "Your queue"
        else -> context
    }
}
