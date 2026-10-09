package dev.letsmoe.drift.ui.screens.podcast

import androidx.lifecycle.viewModelScope
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.Episode
import dev.letsmoe.drift.core.model.PodcastPage
import dev.letsmoe.drift.ui.screens.common.ContinuationPage
import dev.letsmoe.drift.ui.screens.common.ContinuationPager
import dev.letsmoe.drift.ui.screens.common.RemoteViewModel

class PodcastViewModel(private val core: DriftCore, podcastId: String) :
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
