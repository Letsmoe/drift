package dev.neoworks.ottermusic.ui.components

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.LocalAppContainer

/** Looks up the like state of [tracks] in one batch so their rows can show the heart. */
@Composable
fun EnsureRatings(tracks: List<Track>) {
    val libraryState = LocalAppContainer.current.libraryState
    LaunchedEffect(tracks) {
        libraryState.ensureRatings(tracks.map { track -> track.videoId })
    }
}
