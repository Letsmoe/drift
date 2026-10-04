package dev.neoworks.ottermusic.ui.screens.playlist

import androidx.lifecycle.viewModelScope
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.PlaylistPage
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.screens.common.ContinuationPage
import dev.neoworks.ottermusic.ui.screens.common.ContinuationPager
import dev.neoworks.ottermusic.ui.screens.common.RemoteViewModel

class PlaylistViewModel(private val core: OtterCore, playlistId: String) :
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
