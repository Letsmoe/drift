package dev.neoworks.ottermusic.ui.components

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

private val barPhaseDurationsMs = listOf(520, 680, 440)

/** Three equalizer bars; they dance while [isPlaying] and sit low when paused. */
@Composable
fun PlayingIndicator(
    isPlaying: Boolean,
    color: Color,
    modifier: Modifier = Modifier,
    size: Dp = 18.dp,
) {
    val transition = rememberInfiniteTransition(label = "equalizer")
    val barHeights = barPhaseDurationsMs.map { durationMs ->
        val animated by transition.animateFloat(
            initialValue = 0.25f,
            targetValue = 1f,
            animationSpec = infiniteRepeatable(tween(durationMs, easing = LinearEasing), RepeatMode.Reverse),
            label = "bar",
        )
        if (isPlaying) animated else 0.3f
    }
    Canvas(modifier = modifier.size(size)) {
        val barWidth = this.size.width / 5f
        barHeights.forEachIndexed { barIndex, fraction ->
            val barHeight = this.size.height * fraction
            drawRoundRect(
                color = color,
                topLeft = Offset(x = barIndex * barWidth * 2f, y = this.size.height - barHeight),
                size = Size(barWidth, barHeight),
                cornerRadius = CornerRadius(barWidth / 2f),
            )
        }
    }
}
