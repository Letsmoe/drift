package dev.letsmoe.drift.ui.components

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.LocalAppContainer

/** Looks up the like state of [tracks] in one batch so their rows can show the heart. */
@Composable
fun EnsureRatings(tracks: List<Track>) {
    val libraryState = LocalAppContainer.current.libraryState
    LaunchedEffect(tracks) {
        libraryState.ensureRatings(tracks.map { track -> track.videoId })
    }
}
