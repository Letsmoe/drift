package dev.letsmoe.drift.ui.player

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.tween
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.State
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.palette.graphics.Palette
import coil3.SingletonImageLoader
import coil3.request.ImageRequest
import coil3.request.SuccessResult
import coil3.request.allowHardware
import coil3.toBitmap
import dev.letsmoe.drift.ui.theme.Canvas
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

private const val MIN_SATURATION = 0.3f
private const val MAX_SATURATION = 0.7f
private const val MIN_LIGHTNESS = 0.18f
private const val MAX_LIGHTNESS = 0.3f

/**
 * A darkened version of the cover's vibrant (or dominant) color for the Now Playing gradient,
 * animated when the cover changes. Falls back to the app background while loading.
 */
@Composable
fun rememberCoverColor(coverUrl: String?): State<Color> {
    val context = LocalContext.current
    var extractedColor by remember { mutableStateOf(Canvas) }
    LaunchedEffect(coverUrl) {
        if (coverUrl == null) {
            extractedColor = Canvas
            return@LaunchedEffect
        }
        val color = extractCoverColor(context, coverUrl)
        if (color != null) {
            extractedColor = color
        }
    }
    return animateColorAsState(extractedColor, animationSpec = tween(600), label = "coverColor")
}

private suspend fun extractCoverColor(context: android.content.Context, coverUrl: String): Color? {
    val request = ImageRequest.Builder(context)
        .data(coverUrl)
        .allowHardware(false)
        .size(128)
        .build()
    val result = SingletonImageLoader.get(context).execute(request)
    if (result !is SuccessResult) {
        return null
    }
    return withContext(Dispatchers.Default) {
        val palette = Palette.from(result.image.toBitmap()).generate()
        val swatch = pickSwatch(palette)
        if (swatch == null) {
            return@withContext null
        }
        toGradientColor(swatch)
    }
}

private fun pickSwatch(palette: Palette): Palette.Swatch? =
    palette.vibrantSwatch
        ?: palette.lightVibrantSwatch
        ?: palette.mutedSwatch
        ?: palette.lightMutedSwatch
        ?: palette.dominantSwatch

// Dark but never muddy: lightness is pinned to a narrow band and weak colors get some saturation.
private fun toGradientColor(swatch: Palette.Swatch): Color {
    val hsl = swatch.hsl
    val saturation = hsl[1].coerceIn(MIN_SATURATION, MAX_SATURATION)
    val lightness = hsl[2].coerceIn(MIN_LIGHTNESS, MAX_LIGHTNESS)
    return Color.hsl(hsl[0], saturation, lightness)
}

