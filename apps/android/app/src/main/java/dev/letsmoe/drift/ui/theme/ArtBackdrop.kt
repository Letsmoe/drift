package dev.letsmoe.drift.ui.theme

import androidx.compose.runtime.Immutable
import androidx.compose.runtime.State
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Paint
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.drawIntoCanvas
import androidx.compose.ui.graphics.drawscope.scale

/** The four backdrop colors, the same slots as desktop's --art-1..--art-4. */
@Immutable
data class ArtColors(val first: Color, val second: Color, val third: Color, val fourth: Color) {
    companion object {
        /** Desktop's untinted defaults, shown while nothing plays. */
        val Untinted = ArtColors(
            first = Color(0xFF1C1C1E),
            second = Color(0xFF18181B),
            third = Color(0xFF1C1C1E),
            fourth = Color(0xFF141416),
        )
    }
}

/**
 * The playing cover's colors, animated. A State rather than a value so pages read it only while
 * drawing: a color fade then redraws the backdrop without recomposing any page.
 */
val LocalArtColors = staticCompositionLocalOf<State<ArtColors>> { mutableStateOf(ArtColors.Untinted) }

private const val BACKDROP_OPACITY = 0.85f
private const val DARKENING_ALPHA = 0.4f
private val BackdropLayerPaint = Paint().apply { alpha = BACKDROP_OPACITY }

// One blob: an ellipse with radii relative to the size, centered at a relative point.
private class Blob(val centerX: Float, val centerY: Float, val radiusX: Float, val radiusY: Float)

// Desktop's ArtBackdrop.svelte, blob for blob.
private val FirstBlob = Blob(centerX = 0.1f, centerY = 0f, radiusX = 0.6f, radiusY = 0.7f)
private val SecondBlob = Blob(centerX = 0.95f, centerY = 0.05f, radiusX = 0.55f, radiusY = 0.65f)
private val ThirdBlob = Blob(centerX = 0.3f, centerY = 1.05f, radiusX = 0.6f, radiusY = 0.7f)
private val FourthBlob = Blob(centerX = 1f, centerY = 1f, radiusX = 0.5f, radiusY = 0.6f)

/** Paints the app background tinted with soft blobs of the playing cover's colors. */
fun Modifier.artBackdrop(colors: State<ArtColors>): Modifier = drawBehind {
    val current = colors.value
    drawRect(Canvas)
    drawIntoCanvas { canvas ->
        canvas.saveLayer(Rect(Offset.Zero, size), BackdropLayerPaint)
        drawBlob(FirstBlob, current.first)
        drawBlob(SecondBlob, current.second)
        drawBlob(ThirdBlob, current.third)
        drawBlob(FourthBlob, current.fourth)
        drawRect(Canvas.copy(alpha = DARKENING_ALPHA))
        canvas.restore()
    }
}

// A radial gradient is circular, so draw a circle and stretch it vertically into the ellipse.
private fun DrawScope.drawBlob(blob: Blob, color: Color) {
    val center = Offset(size.width * blob.centerX, size.height * blob.centerY)
    val radius = size.width * blob.radiusX
    val verticalStretch = size.height * blob.radiusY / radius
    val brush = Brush.radialGradient(listOf(color, color.copy(alpha = 0f)), center = center, radius = radius)
    scale(scaleX = 1f, scaleY = verticalStretch, pivot = center) {
        drawRect(brush, topLeft = center - Offset(radius, radius), size = Size(radius * 2, radius * 2))
    }
}
