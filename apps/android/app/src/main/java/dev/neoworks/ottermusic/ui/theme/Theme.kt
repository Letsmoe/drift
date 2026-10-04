package dev.neoworks.ottermusic.ui.theme

import androidx.compose.foundation.shape.RoundedCornerShape
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
data class OtterColors(
    val surfaceRaised: Color = SurfaceRaised,
    val surfaceHover: Color = SurfaceHover,
    val border: Color = Border,
    val borderStrong: Color = BorderStrong,
    val textMuted: Color = TextMuted,
    val textDim: Color = TextDim,
    val accent: Color = Accent,
)

private val LocalOtterColors = staticCompositionLocalOf { OtterColors() }

/** Access to the extra tokens: `OtterTheme.colors.textDim`. */
object OtterTheme {
    val colors: OtterColors
        @Composable
        @ReadOnlyComposable
        get() = LocalOtterColors.current
}

private val OtterColorScheme = darkColorScheme(
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

private val OtterShapes = Shapes(
    extraSmall = RoundedCornerShape(6.dp),
    small = RoundedCornerShape(10.dp),
    medium = RoundedCornerShape(14.dp),
    large = RoundedCornerShape(20.dp),
    extraLarge = RoundedCornerShape(28.dp),
)

@Composable
fun OtterMusicTheme(content: @Composable () -> Unit) {
    CompositionLocalProvider(LocalOtterColors provides OtterColors()) {
        MaterialTheme(
            colorScheme = OtterColorScheme,
            typography = OtterTypography,
            shapes = OtterShapes,
            content = content,
        )
    }
}
