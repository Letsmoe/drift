package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.Dp
import dev.neoworks.ottermusic.core.model.Thumbnail

private const val MOSAIC_CELLS = 4

/**
 * A 2x2 mosaic of covers for a local playlist; with fewer than four covers it shows the first
 * one full size (empty list: the placeholder cover).
 */
@Composable
fun CoverMosaic(coverThumbnails: List<List<Thumbnail>>, size: Dp, modifier: Modifier = Modifier) {
    if (coverThumbnails.size < MOSAIC_CELLS) {
        val first = coverThumbnails.firstOrNull() ?: emptyList()
        Cover(thumbnails = first, size = size, modifier = modifier)
        return
    }
    val cellSize = size / 2
    val cellShape = RoundedCornerShape(0)
    Column(modifier = modifier.size(size)) {
        for (rowCovers in coverThumbnails.take(MOSAIC_CELLS).chunked(2)) {
            Row {
                for (thumbnails in rowCovers) {
                    Cover(thumbnails = thumbnails, size = cellSize, shape = cellShape)
                }
            }
        }
    }
}
