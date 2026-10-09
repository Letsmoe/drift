package dev.letsmoe.drift.ui.screens.playlist

import androidx.lifecycle.viewModelScope
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.PlaylistPage
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.screens.common.ContinuationPage
import dev.letsmoe.drift.ui.screens.common.ContinuationPager
import dev.letsmoe.drift.ui.screens.common.RemoteViewModel

class PlaylistViewModel(private val core: DriftCore, playlistId: String) :
    RemoteViewModel<PlaylistPage>({ core.playlist(playlistId) }) {

    /** All loaded tracks: the first page plus continuation pages. */
    val trackPager = ContinuationPager<Track>(viewModelScope) { token -> fetchMoreTracks(token) }

    override fun onLoaded(data: PlaylistPage) {
        trackPager.reset(data.tracks, data.continuation)
    }

    private suspend fun fetchMoreTracks(token: String): ContinuationPage<Track> {
        val next = core.playlistContinuation(token)
        return ContinuationPage(next.tracks, next.continuation)
    }
}
