package dev.letsmoe.drift.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import coil3.compose.AsyncImage
import dev.letsmoe.drift.core.model.Thumbnail
import dev.letsmoe.drift.core.model.bestFor
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme
import androidx.compose.ui.graphics.Shape

val CoverShape = RoundedCornerShape(8.dp)

/** A square cover of fixed [size]; the best thumbnail for that size is loaded. Use [circle] for artists. */
@Composable
fun Cover(
    thumbnails: List<Thumbnail>,
    size: Dp,
    modifier: Modifier = Modifier,
    shape: Shape = CoverShape,
    circle: Boolean = false,
) {
    CoverFill(
        thumbnails = thumbnails,
        requestSize = size,
        modifier = modifier.size(size),
        shape = shape,
        circle = circle,
    )
}

/**
 * A cover whose layout size comes from [modifier] (e.g. `Modifier.fillMaxWidth().aspectRatio(1f)`);
 * [requestSize] only picks the thumbnail resolution.
 */
@Composable
fun CoverFill(
    thumbnails: List<Thumbnail>,
    requestSize: Dp,
    modifier: Modifier = Modifier,
    shape: Shape = CoverShape,
    circle: Boolean = false,
) {
    val density = LocalDensity.current
    val url = remember(thumbnails, requestSize, density) {
        thumbnails.bestFor(with(density) { requestSize.roundToPx() })?.url
    }
    val clipShape = if (circle) CircleShape else shape
    Box(
        modifier = modifier
            .clip(clipShape)
            .background(DriftTheme.colors.surfaceRaised),
        contentAlignment = Alignment.Center,
    ) {
        if (url == null) {
            Icon(DriftIcons.MusicNote.outline, contentDescription = null, tint = DriftTheme.colors.textDim)
            return@Box
        }
        AsyncImage(
            model = url,
            contentDescription = null,
            contentScale = ContentScale.Crop,
            modifier = Modifier.matchParentSize(),
        )
    }
}
