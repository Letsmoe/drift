package dev.neoworks.ottermusic.ui.navigation

import androidx.compose.animation.EnterTransition
import androidx.compose.animation.ExitTransition
import androidx.compose.animation.core.CubicBezierEasing
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.navigation.NavDestination
import androidx.navigation.NavDestination.Companion.hasRoute

private val DecelerateEasing = CubicBezierEasing(0.2f, 0f, 0f, 1f)
private val AccelerateEasing = CubicBezierEasing(0.3f, 0f, 1f, 1f)
private const val ENTER_DURATION_MS = 320
private const val EXIT_DURATION_MS = 240
private const val SLIDE_DIVISOR = 6

private fun isTabRoot(destination: NavDestination): Boolean =
    destination.hasRoute<Routes.Home>() ||
        destination.hasRoute<Routes.Search>() ||
        destination.hasRoute<Routes.Library>()

/** Tab switches fade through; drilling into a page slides it in a short way while fading. */
fun enterTransitionFor(from: NavDestination, to: NavDestination): EnterTransition {
    if (isTabRoot(from) && isTabRoot(to)) {
        return fadeIn(tween(durationMillis = 200, delayMillis = 40))
    }
    return slideInHorizontally(tween(ENTER_DURATION_MS, easing = DecelerateEasing)) { fullWidth -> fullWidth / SLIDE_DIVISOR } +
        fadeIn(tween(ENTER_DURATION_MS, easing = DecelerateEasing))
}

/** The page being left slides out to the right while fading; the page below stays still. */
fun popExitTransitionFor(from: NavDestination, to: NavDestination): ExitTransition {
    if (isTabRoot(from) && isTabRoot(to)) {
        return fadeOut(tween(120))
    }
    return slideOutHorizontally(tween(EXIT_DURATION_MS, easing = AccelerateEasing)) { fullWidth -> fullWidth / SLIDE_DIVISOR } +
        fadeOut(tween(EXIT_DURATION_MS, easing = AccelerateEasing))
}
