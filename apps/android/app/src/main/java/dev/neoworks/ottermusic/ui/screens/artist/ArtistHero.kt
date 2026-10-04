package dev.neoworks.ottermusic.ui.screens.artist

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.core.model.Thumbnail
import dev.neoworks.ottermusic.ui.components.CoverFill
import dev.neoworks.ottermusic.ui.theme.Canvas

val ArtistHeroHeight = 360.dp
private const val HERO_REQUEST_WIDTH_DP = 480

/** Full-bleed artist image fading into the background, with the name bottom-left. */
@Composable
fun ArtistHero(name: String, thumbnails: List<Thumbnail>, modifier: Modifier = Modifier) {
    Box(modifier = modifier.fillMaxWidth().height(ArtistHeroHeight)) {
        CoverFill(
            thumbnails = thumbnails,
            requestSize = HERO_REQUEST_WIDTH_DP.dp,
            modifier = Modifier.fillMaxSize(),
            shape = RoundedCornerShape(0.dp),
        )
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        0.0f to Color.Black.copy(alpha = 0.35f),
                        0.45f to Color.Transparent,
                        1.0f to Canvas,
                    ),
                ),
        )
        Text(
            text = name,
            style = MaterialTheme.typography.displaySmall,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier
                .align(Alignment.BottomStart)
                .padding(start = 16.dp, end = 16.dp, bottom = 12.dp),
        )
    }
}
