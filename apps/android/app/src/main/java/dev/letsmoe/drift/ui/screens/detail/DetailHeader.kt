package dev.letsmoe.drift.ui.screens.detail

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.core.model.Thumbnail
import dev.letsmoe.drift.core.model.bestFor
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.components.CoverShape
import dev.letsmoe.drift.ui.player.rememberCoverColor
import dev.letsmoe.drift.ui.theme.Canvas

private val TopBarHeight = 56.dp
private const val COVER_COLOR_REQUEST_PIXELS = 256

/**
 * The header of album, playlist, podcast and similar pages: a large centered cover with a
 * soft gradient from the cover's color into the background, the [title] and [meta] lines below.
 * Leaves room at the top for the floating [DetailTopBar].
 *
 * @param cover replaces the cover image, e.g. for the Liked Songs tile or a cover mosaic.
 */
@Composable
fun DetailHeader(
    title: String,
    thumbnails: List<Thumbnail>,
    modifier: Modifier = Modifier,
    coverSize: Dp = 240.dp,
    gradientColor: Color? = null,
    cover: (@Composable () -> Unit)? = null,
    meta: @Composable ColumnScope.() -> Unit = {},
) {
    val coverUrl = thumbnails.bestFor(COVER_COLOR_REQUEST_PIXELS)?.url
    val extractedColor by rememberCoverColor(coverUrl)
    val topColor = gradientColor ?: extractedColor
    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(Brush.verticalGradient(listOf(topColor, Canvas)))
            .statusBarsPadding()
            .padding(top = TopBarHeight, bottom = 8.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Column(modifier = Modifier.shadow(elevation = 24.dp, shape = CoverShape)) {
            if (cover != null) {
                cover()
            } else {
                Cover(thumbnails = thumbnails, size = coverSize)
            }
        }
        Text(
            text = title,
            style = MaterialTheme.typography.headlineSmall,
            textAlign = TextAlign.Center,
            maxLines = 3,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.padding(start = 24.dp, end = 24.dp, top = 20.dp),
        )
        Column(
            modifier = Modifier.padding(horizontal = 24.dp, vertical = 4.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            content = meta,
        )
    }
}

/** A centered row of round buttons (save, more) between the header and the Play / Shuffle pills. */
@Composable
fun DetailActionRow(modifier: Modifier = Modifier, content: @Composable () -> Unit) {
    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.CenterHorizontally),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        content()
    }
}
