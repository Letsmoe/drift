package dev.letsmoe.drift.ui.player

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.State
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.graphics.Color
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.CoverColors
import dev.letsmoe.drift.core.model.bestFor
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.theme.ArtColors
import kotlinx.coroutines.CancellationException

// The colors come from a small thumbnail; a bigger one finds the same colors slower.
private const val SAMPLED_THUMBNAIL_PIXELS = 60
private const val COLOR_FADE_MS = 1200

/** The playing track's cover colors for the backdrop, fading over 1.2 s when the song changes. */
@Composable
fun rememberNowPlayingArtColors(): State<ArtColors> {
    val container = LocalAppContainer.current
    val currentTrack by container.playerController.currentTrack.collectAsStateWithLifecycle()
    val coverUrl = currentTrack?.thumbnails?.bestFor(SAMPLED_THUMBNAIL_PIXELS)?.url
    var target by remember { mutableStateOf(ArtColors.Untinted) }
    LaunchedEffect(coverUrl) {
        target = loadArtColors(container.core, coverUrl)
    }
    val first by animateArtColor(target.first)
    val second by animateArtColor(target.second)
    val third by animateArtColor(target.third)
    val fourth by animateArtColor(target.fourth)
    return remember { derivedStateOf { ArtColors(first, second, third, fourth) } }
}

@Composable
private fun animateArtColor(color: Color): State<Color> =
    animateColorAsState(color, animationSpec = tween(COLOR_FADE_MS, easing = FastOutSlowInEasing), label = "artColor")

// A cover that can't be read leaves the backdrop untinted, like desktop.
private suspend fun loadArtColors(core: DriftCore, coverUrl: String?): ArtColors {
    if (coverUrl == null) {
        return ArtColors.Untinted
    }
    try {
        return toArtColors(core.coverColors(coverUrl))
    } catch (cancellation: CancellationException) {
        throw cancellation
    } catch (failure: Exception) {
        return ArtColors.Untinted
    }
}

// Covers with fewer distinct colors repeat them; the second slot is always the vibrant color.
private fun toArtColors(cover: CoverColors): ArtColors {
    if (cover.vibrant.isEmpty()) {
        return ArtColors.Untinted
    }
    val vibrant = parseHexColor(cover.vibrant)
    if (cover.colors.isEmpty()) {
        return ArtColors(vibrant, vibrant, vibrant, vibrant)
    }
    val colors = cover.colors.map(::parseHexColor)
    return ArtColors(
        first = colors[0],
        second = vibrant,
        third = colors[2 % colors.size],
        fourth = colors[3 % colors.size],
    )
}

private fun parseHexColor(hex: String): Color = Color(0xFF000000 or hex.removePrefix("#").toLong(radix = 16))
