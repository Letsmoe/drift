package dev.letsmoe.drift.ui.theme

import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.LocalContentColor
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Shapes
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.Immutable
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

/** Design tokens that Material 3's color scheme has no slot for. */
@Immutable
data class DriftColors(
    val surfaceRaised: Color = SurfaceRaised,
    val surfaceHover: Color = SurfaceHover,
    val border: Color = Border,
    val borderStrong: Color = BorderStrong,
    val textMuted: Color = TextMuted,
    val textDim: Color = TextDim,
    val accent: Color = Accent,
)

private val LocalDriftColors = staticCompositionLocalOf { DriftColors() }

/** Access to the extra tokens: `DriftTheme.colors.textDim`. */
object DriftTheme {
    val colors: DriftColors
        @Composable
        @ReadOnlyComposable
        get() = LocalDriftColors.current
}

private val DriftColorScheme = darkColorScheme(
    primary = Accent,
    onPrimary = OnAccent,
    primaryContainer = Accent,
    onPrimaryContainer = OnAccent,
    secondary = SurfaceHover,
    onSecondary = TextPrimary,
    secondaryContainer = SurfaceHover,
    onSecondaryContainer = TextPrimary,
    tertiary = Accent,
    background = Canvas,
    onBackground = TextPrimary,
    surface = Surface,
    onSurface = TextPrimary,
    surfaceVariant = SurfaceRaised,
    onSurfaceVariant = TextMuted,
    surfaceContainerLowest = Canvas,
    surfaceContainerLow = Surface,
    surfaceContainer = Surface,
    surfaceContainerHigh = SurfaceRaised,
    surfaceContainerHighest = SurfaceHover,
    outline = BorderStrong,
    outlineVariant = Border,
    error = Danger,
    onError = OnAccent,
)

private val DriftShapes = Shapes(
    extraSmall = RoundedCornerShape(6.dp),
    small = RoundedCornerShape(10.dp),
    medium = RoundedCornerShape(14.dp),
    large = RoundedCornerShape(20.dp),
    extraLarge = RoundedCornerShape(28.dp),
)

@Composable
fun DriftAppTheme(content: @Composable () -> Unit) {
    CompositionLocalProvider(LocalDriftColors provides DriftColors()) {
        MaterialTheme(
            colorScheme = DriftColorScheme,
            typography = DriftTypography,
            shapes = DriftShapes,
        ) {
            // Without a Surface ancestor, text defaults to black on the dark canvas.
            CompositionLocalProvider(LocalContentColor provides DriftColorScheme.onSurface, content = content)
        }
    }
}
