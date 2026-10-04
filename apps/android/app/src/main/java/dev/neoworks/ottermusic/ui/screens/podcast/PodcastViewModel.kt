package dev.neoworks.ottermusic.ui.screens.podcast

import androidx.lifecycle.viewModelScope
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.Episode
import dev.neoworks.ottermusic.core.model.PodcastPage
import dev.neoworks.ottermusic.ui.screens.common.ContinuationPage
import dev.neoworks.ottermusic.ui.screens.common.ContinuationPager
import dev.neoworks.ottermusic.ui.screens.common.RemoteViewModel

class PodcastViewModel(private val core: OtterCore, podcastId: String) :
    RemoteViewModel<PodcastPage>({ core.podcast(podcastId) }) {

    /** All loaded episodes: the first page plus continuation pages. */
    val episodePager = ContinuationPager<Episode>(viewModelScope) { token -> fetchMoreEpisodes(token) }

    override fun onLoaded(data: PodcastPage) {
        episodePager.reset(data.episodes, data.continuation)
    }

    private suspend fun fetchMoreEpisodes(token: String): ContinuationPage<Episode> {
        val next = core.podcastContinuation(token)
        return ContinuationPage(next.episodes, next.continuation)
    }
}
