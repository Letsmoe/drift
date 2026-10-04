package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.core.model.ItemKind
import dev.neoworks.ottermusic.core.model.Shelf
import dev.neoworks.ottermusic.core.model.Thumbnail
import dev.neoworks.ottermusic.ui.theme.OtterTheme

val MediaCardSize = 160.dp

/** A square cover with a title and subtitle underneath; the cover is a circle for artists. */
@Composable
fun MediaCard(
    title: String,
    subtitle: String,
    thumbnails: List<Thumbnail>,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    circle: Boolean = false,
    size: Dp = MediaCardSize,
) {
    Column(
        modifier = modifier
            .width(size)
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onClick),
    ) {
        Cover(thumbnails = thumbnails, size = size, circle = circle, shape = RoundedCornerShape(10.dp))
        Text(
            text = title,
            style = MaterialTheme.typography.titleSmall,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            textAlign = if (circle) TextAlign.Center else TextAlign.Start,
            modifier = Modifier
                .width(size)
                .padding(top = 8.dp),
        )
        if (subtitle.isNotEmpty()) {
            Text(
                text = subtitle,
                style = MaterialTheme.typography.bodySmall,
                color = OtterTheme.colors.textDim,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
                textAlign = if (circle) TextAlign.Center else TextAlign.Start,
                modifier = Modifier.width(size),
            )
        }
    }
}

/** A card for a shelf [Item]; artists get a circle. */
@Composable
fun MediaCard(item: Item, onClick: () -> Unit, modifier: Modifier = Modifier, size: Dp = MediaCardSize) {
    MediaCard(
        title = item.title,
        subtitle = item.subtitle,
        thumbnails = item.thumbnails,
        onClick = onClick,
        modifier = modifier,
        circle = item.kind == ItemKind.Artist,
        size = size,
    )
}

/** A titled, horizontally scrolling row; fill it with [content] (`items(...) { MediaCard(...) }`). */
@Composable
fun ShelfRow(
    title: String,
    modifier: Modifier = Modifier,
    content: LazyListScope.() -> Unit,
) {
    Column(modifier = modifier) {
        SectionHeader(title = title)
        LazyRow(
            contentPadding = PaddingValues(horizontal = 16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            content = content,
        )
    }
}

/** A shelf of cards for a [Shelf]; the caller decides what tapping an item does. */
@Composable
fun ShelfRow(shelf: Shelf, onItemClick: (Item) -> Unit, modifier: Modifier = Modifier) {
    ShelfRow(title = shelf.title, modifier = modifier) {
        items(shelf.items, key = null) { item ->
            MediaCard(item = item, onClick = { onItemClick(item) })
        }
    }
}
