package dev.letsmoe.drift.ui.screens.search

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.core.model.Item
import dev.letsmoe.drift.core.model.ItemKind
import dev.letsmoe.drift.core.model.Shelf
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.components.EnsureRatings
import dev.letsmoe.drift.ui.components.DriftChip
import dev.letsmoe.drift.ui.components.SectionHeader
import dev.letsmoe.drift.ui.components.TrackRow
import dev.letsmoe.drift.ui.components.rememberItemOpener
import dev.letsmoe.drift.ui.theme.DriftTheme

private const val SHELF_PREVIEW_COUNT = 4

/** Everything the results part of the search screen needs to render one finished search. */
data class SearchResultsContent(
    val shelves: List<Shelf>,
    val query: String,
    val filterTitle: String,
    val onSelectFilter: (String) -> Unit,
)

/**
 * The results list: filter chips (All plus one per returned shelf), the top result and the
 * shelves as vertical lists. A selected chip shows only its shelf, in full.
 */
fun LazyListScope.searchResults(content: SearchResultsContent) {
    val shelves = content.shelves
    item(key = "filters") {
        FilterChips(titles = shelves.filterTitles(), selected = content.filterTitle, onSelect = content.onSelectFilter)
    }
    val filtered = shelves.firstOrNull { shelf -> shelf.title == content.filterTitle }
    if (filtered != null) {
        shelfItems(shelf = filtered, query = content.query, limit = Int.MAX_VALUE, onSeeAll = null)
        return
    }
    val topResult = shelves.topResult()
    if (topResult != null) {
        item(key = "top-result") { TopResult(item = topResult, query = content.query) }
    }
    for (shelf in shelves.listShelves()) {
        shelfItems(
            shelf = shelf,
            query = content.query,
            limit = SHELF_PREVIEW_COUNT,
            onSeeAll = { content.onSelectFilter(shelf.title) },
        )
    }
}

@Composable
private fun FilterChips(titles: List<String>, selected: String, onSelect: (String) -> Unit) {
    LazyRow(
        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        item { DriftChip(label = "All", isSelected = selected.isEmpty(), onClick = { onSelect("") }) }
        items(titles.size) { index ->
            DriftChip(label = titles[index], isSelected = titles[index] == selected, onClick = { onSelect(titles[index]) })
        }
    }
}

@Composable
private fun TopResult(item: Item, query: String) {
    val openItem = rememberItemOpener("search:$query")
    Column {
        SectionHeader(title = "Top result")
        TopResultCard(item = item, onClick = { openItem(item) }, modifier = Modifier.padding(horizontal = 16.dp))
    }
}

private fun LazyListScope.shelfItems(shelf: Shelf, query: String, limit: Int, onSeeAll: (() -> Unit)?) {
    val visibleItems = shelf.items.take(limit)
    var actionLabel: String? = null
    if (onSeeAll != null && shelf.items.size > limit) {
        actionLabel = "See all"
    }
    item(key = "header-${shelf.title}") {
        SectionHeader(title = shelf.title, actionLabel = actionLabel, onAction = onSeeAll)
    }
    val tracks = shelf.items.mapNotNull { shelfItem -> shelfItem.track }
    val isTrackShelf = tracks.size == shelf.items.size
    items(visibleItems.size, key = { index -> "${shelf.title}-$index-${visibleItems[index].id}" }) { index ->
        if (isTrackShelf) {
            TrackResultRow(item = visibleItems[index], index = index, tracks = tracks, query = query)
            return@items
        }
        ItemResultRow(item = visibleItems[index], query = query)
    }
}

@Composable
private fun TrackResultRow(item: Item, index: Int, tracks: List<Track>, query: String) {
    val playerController = LocalAppContainer.current.playerController
    val track = tracks[index]
    EnsureRatings(listOf(track))
    TrackRow(
        track = track,
        kind = item.kind,
        onClick = { playerController.playOrToggle(track, tracks, index, "search:$query", item.kind) },
    )
}

/** An album, playlist, artist or podcast result: cover (round for artists), title and subtitle. */
@Composable
private fun ItemResultRow(item: Item, query: String) {
    val openItem = rememberItemOpener("search:$query")
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { openItem(item) }
            .padding(horizontal = 16.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Cover(thumbnails = item.thumbnails, size = 56.dp, circle = item.kind == ItemKind.Artist)
        Column(modifier = Modifier.weight(1f)) {
            Text(item.title, style = MaterialTheme.typography.bodyLarge, maxLines = 1, overflow = TextOverflow.Ellipsis)
            if (item.subtitle.isNotEmpty()) {
                Text(
                    item.subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = DriftTheme.colors.textDim,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}
