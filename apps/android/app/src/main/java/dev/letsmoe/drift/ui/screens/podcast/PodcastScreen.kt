package dev.letsmoe.drift.ui.screens.podcast

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.PodcastPage
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.ExpandableText
import dev.letsmoe.drift.ui.components.FollowButton
import dev.letsmoe.drift.ui.components.LoadingMoreFooter
import dev.letsmoe.drift.ui.components.OnNearEnd
import dev.letsmoe.drift.ui.components.RoundPlayButton
import dev.letsmoe.drift.ui.components.SectionHeader
import dev.letsmoe.drift.ui.components.rememberIsContextPlaying
import dev.letsmoe.drift.ui.navigation.openArtist
import dev.letsmoe.drift.ui.screens.detail.DetailHeader
import dev.letsmoe.drift.ui.screens.detail.DetailLoadState
import dev.letsmoe.drift.ui.screens.detail.DetailScaffold
import dev.letsmoe.drift.ui.screens.detail.detailListPadding
import dev.letsmoe.drift.ui.theme.DriftTheme

@Composable
fun PodcastScreen(podcastId: String, modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> PodcastViewModel(container.core, podcastId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    DetailLoadState(state = state, onRetry = viewModel::retry, modifier = modifier) { page ->
        PodcastContent(podcastId = podcastId, page = page, viewModel = viewModel, modifier = modifier)
    }
}

@Composable
private fun PodcastContent(
    podcastId: String,
    page: PodcastPage,
    viewModel: PodcastViewModel,
    modifier: Modifier,
) {
    val playerController = LocalAppContainer.current.playerController
    val listState = rememberLazyListState()
    val episodes by viewModel.episodePager.entries.collectAsStateWithLifecycle()
    val hasMore by viewModel.episodePager.hasMore.collectAsStateWithLifecycle()
    val isLoadingMore by viewModel.episodePager.isLoadingMore.collectAsStateWithLifecycle()
    val episodeTracks = episodes.map { episode -> episode.track }
    val context = "podcast:$podcastId"
    val isPlaying = rememberIsContextPlaying(context)
    val podcastItem = Item(
        kind = ItemKind.Podcast,
        id = podcastId,
        title = page.title,
        subtitle = page.author.name,
        thumbnails = page.thumbnails,
    )
    listState.OnNearEnd(isEnabled = hasMore, onNearEnd = viewModel.episodePager::loadMore)

    DetailScaffold(title = page.title, listState = listState, modifier = modifier) {
        LazyColumn(state = listState, contentPadding = detailListPadding()) {
            item {
                DetailHeader(title = page.title, thumbnails = page.thumbnails) {
                    PodcastAuthor(page = page)
                    if (page.description.isNotEmpty()) {
                        ExpandableText(text = page.description, modifier = Modifier.padding(top = 8.dp))
                    }
                }
            }
            item {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    FollowButton(item = podcastItem)
                    RoundPlayButton(
                        isPlaying = isPlaying,
                        onClick = { playerController.playContext(episodeTracks, context, ItemKind.Episode) },
                        contentDescription = "Play ${page.title}",
                    )
                }
            }
            item { SectionHeader(title = "All episodes") }
            itemsIndexed(episodes, key = { index, episode -> "$index-${episode.track.videoId}" }) { index, episode ->
                EpisodeRow(
                    episode = episode,
                    onPlay = {
                        playerController.playOrToggle(episode.track, episodeTracks, index, context, ItemKind.Episode)
                    },
                )
            }
            if (isLoadingMore) {
                item { LoadingMoreFooter() }
            }
        }
    }
}

@Composable
private fun PodcastAuthor(page: PodcastPage) {
    val navController = LocalNavController.current
    val author = page.author
    if (author.name.isEmpty()) {
        return
    }
    Text(
        text = author.name,
        style = MaterialTheme.typography.titleMedium,
        color = MaterialTheme.colorScheme.onSurface,
        modifier = Modifier
            .clickable(enabled = author.id.isNotEmpty()) { navController.openArtist(author.id) }
            .padding(vertical = 2.dp),
    )
    Text("Podcast", style = MaterialTheme.typography.bodySmall, color = DriftTheme.colors.textDim)
}
