package dev.neoworks.ottermusic.ui.components

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.ui.theme.OtterTheme

private const val SKELETON_PULSE_MS = 900

/** A pulsing placeholder block; size it with [modifier]. */
@Composable
fun Skeleton(modifier: Modifier = Modifier, shape: Shape = RoundedCornerShape(8.dp)) {
    val transition = rememberInfiniteTransition(label = "skeleton")
    val pulse by transition.animateFloat(
        initialValue = 0.45f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(tween(SKELETON_PULSE_MS), RepeatMode.Reverse),
        label = "skeletonPulse",
    )
    Box(
        modifier = modifier
            .alpha(pulse)
            .clip(shape)
            .background(OtterTheme.colors.surfaceRaised),
    )
}
