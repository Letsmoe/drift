package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.ui.theme.Surface

val FloatingBarShape = RoundedCornerShape(28.dp)

/**
 * The floating "glass" look of the tab bar, mini player and action sheets: a translucent surface,
 * a hairline border and a soft shadow. Compose cannot blur what is behind a view cheaply, so the
 * translucency stands in for the frosted blur.
 */
fun Modifier.glassSurface(shape: Shape = FloatingBarShape, alpha: Float = 0.96f): Modifier =
    this
        .shadow(elevation = 16.dp, shape = shape, ambientColor = Color.Black, spotColor = Color.Black)
        .clip(shape)
        .background(Surface.copy(alpha = alpha))
        .border(width = 1.dp, color = Color.White.copy(alpha = 0.10f), shape = shape)
