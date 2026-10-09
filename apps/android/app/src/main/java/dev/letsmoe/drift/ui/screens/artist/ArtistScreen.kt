package dev.letsmoe.drift.ui.screens.artist

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.ArtistPage
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.appViewModel
import dev.letsmoe.drift.ui.components.EnsureRatings
import dev.letsmoe.drift.ui.components.CollectionMoreButton
import dev.letsmoe.drift.ui.components.ExpandableText
import dev.letsmoe.drift.ui.components.FollowButton
import dev.letsmoe.drift.ui.components.RoundPlayButton
import dev.letsmoe.drift.ui.components.SectionHeader
import dev.letsmoe.drift.ui.components.ShelfRow
import dev.letsmoe.drift.ui.components.TrackRow
import dev.letsmoe.drift.ui.components.rememberIsContextPlaying
import dev.letsmoe.drift.ui.components.rememberItemOpener
import dev.letsmoe.drift.ui.screens.detail.DetailLoadState
import dev.letsmoe.drift.ui.screens.detail.DetailScaffold
import dev.letsmoe.drift.ui.screens.detail.detailListPadding
import dev.letsmoe.drift.ui.theme.DriftTheme

private const val POPULAR_COLLAPSED_COUNT = 5
private const val POPULAR_EXPANDED_COUNT = 10

@Composable
fun ArtistScreen(artistId: String, modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> ArtistViewModel(container.core, artistId) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    DetailLoadState(state = state, onRetry = viewModel::retry, modifier = modifier) { page ->
        ArtistContent(artistId = artistId, page = page, modifier = modifier)
    }
}

@Composable
private fun ArtistContent(artistId: String, page: ArtistPage, modifier: Modifier) {
    val listState = rememberLazyListState()
    val context = "artist:$artistId"
    val openItem = rememberItemOpener(context)
    var showMorePopular by rememberSaveable { mutableStateOf(false) }
    EnsureRatings(page.topSongs)

    DetailScaffold(
        title = page.name,
        listState = listState,
        modifier = modifier,
        titleScrollThreshold = ArtistHeroHeight - 120.dp,
    ) {
        LazyColumn(state = listState, contentPadding = detailListPadding()) {
            item { ArtistHero(name = page.name, thumbnails = page.thumbnails) }
            item { ArtistActions(artistId = artistId, page = page, context = context) }
            popularSection(
                page = page,
                context = context,
                showMore = showMorePopular,
                onToggleShowMore = { showMorePopular = !showMorePopular },
            )
            for (shelf in page.shelves) {
                if (shelf.items.isNotEmpty()) {
                    item { ShelfRow(shelf = shelf, onItemClick = openItem, modifier = Modifier.padding(top = 16.dp)) }
                }
            }
            if (page.description.isNotEmpty()) {
                item { AboutCard(page = page) }
            }
        }
    }
}

@Composable
private fun ArtistActions(artistId: String, page: ArtistPage, context: String) {
    val playerController = LocalAppContainer.current.playerController
    val isPlaying = rememberIsContextPlaying(context)
    val artistItem = Item(
        kind = ItemKind.Artist,
        id = artistId,
        title = page.name,
        subtitle = "Artist",
        thumbnails = page.thumbnails,
    )
    Row(
        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        FollowButton(item = artistItem)
        CollectionMoreButton(title = page.name, tracks = page.topSongs)
        Row(modifier = Modifier.weight(1f), horizontalArrangement = Arrangement.End) {
            RoundPlayButton(
                isPlaying = isPlaying,
                onClick = { playerController.playContext(page.topSongs, context) },
                contentDescription = "Play ${page.name}",
            )
        }
    }
}

private fun LazyListScope.popularSection(
    page: ArtistPage,
    context: String,
    showMore: Boolean,
    onToggleShowMore: () -> Unit,
) {
    if (page.topSongs.isEmpty()) {
        return
    }
    var visibleCount = POPULAR_COLLAPSED_COUNT
    if (showMore) {
        visibleCount = POPULAR_EXPANDED_COUNT
    }
    val visibleSongs = page.topSongs.take(visibleCount)
    item { SectionHeader(title = "Popular", modifier = Modifier.padding(top = 8.dp)) }
    itemsIndexed(visibleSongs, key = { index, track -> "popular-$index-${track.videoId}" }) { index, track ->
        PopularRow(track = track, topSongs = page.topSongs, context = context, index = index)
    }
    if (page.topSongs.size > POPULAR_COLLAPSED_COUNT) {
        item { ShowMoreLabel(isExpanded = showMore, onClick = onToggleShowMore) }
    }
}

@Composable
private fun PopularRow(track: Track, topSongs: List<Track>, context: String, index: Int) {
    val playerController = LocalAppContainer.current.playerController
    TrackRow(
        track = track,
        onClick = { playerController.playOrToggle(track, topSongs, index, context) },
        trackNumber = index + 1,
        subtitle = "",
    )
}

@Composable
private fun ShowMoreLabel(isExpanded: Boolean, onClick: () -> Unit) {
    var label = "See more"
    if (isExpanded) {
        label = "Show less"
    }
    Text(
        text = label,
        style = MaterialTheme.typography.labelLarge,
        color = DriftTheme.colors.textMuted,
        modifier = Modifier
            .clip(RoundedCornerShape(50))
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 12.dp),
    )
}

@Composable
private fun AboutCard(page: ArtistPage) {
    Column(
        modifier = Modifier
            .padding(16.dp)
            .fillMaxWidth()
            .clip(RoundedCornerShape(20.dp))
            .background(DriftTheme.colors.surfaceRaised)
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(text = "About", style = MaterialTheme.typography.titleLarge)
        ExpandableText(text = page.description, collapsedLines = 4)
    }
}
