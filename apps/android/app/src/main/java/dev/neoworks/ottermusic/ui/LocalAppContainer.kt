package dev.neoworks.ottermusic.ui

import androidx.compose.runtime.compositionLocalOf
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController
import dev.neoworks.ottermusic.AppContainer

/** The hand-wired dependencies; provided once at the root. Use `LocalAppContainer.current.playerController` etc. */
val LocalAppContainer = compositionLocalOf<AppContainer> { error("AppContainer not provided") }

/** The app's NavController, so shared components can navigate without callbacks threaded through. */
val LocalNavController = compositionLocalOf<NavController> { error("NavController not provided") }

/**
 * Height of the floating mini player + tab bar stack (including the system navigation bar).
 * Scrolling screens add it as bottom content padding so content can scroll under the bars.
 */
val LocalBottomOverlayPadding = compositionLocalOf { 0.dp }

