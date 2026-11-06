package dev.letsmoe.drift.ui.navigation

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

// Material 3's emphasized easing: a quick start that settles slowly, so a page lands softly.
private val EmphasizedEasing = CubicBezierEasing(0.05f, 0.7f, 0.1f, 1f)
private const val PAGE_DURATION_MS = 380
private const val TAB_FADE_IN_MS = 180
private const val TAB_FADE_OUT_MS = 120
// The page underneath moves a quarter of the width and dims, so the new page reads as on top.
private const val UNDERNEATH_SLIDE_DIVISOR = 4
private const val UNDERNEATH_ALPHA = 0.6f

private fun isTabRoot(destination: NavDestination): Boolean =
    destination.hasRoute<Routes.Home>() ||
        destination.hasRoute<Routes.Search>() ||
        destination.hasRoute<Routes.Library>()

private fun isTabSwitch(from: NavDestination, to: NavDestination): Boolean = isTabRoot(from) && isTabRoot(to)

private fun <T> pageTween() = tween<T>(PAGE_DURATION_MS, easing = EmphasizedEasing)

/** Tab switches cross-fade; a page that opens slides in from the right edge over the one below. */
fun enterTransitionFor(from: NavDestination, to: NavDestination): EnterTransition {
    if (isTabSwitch(from, to)) {
        return fadeIn(tween(TAB_FADE_IN_MS))
    }
    return slideInHorizontally(pageTween()) { fullWidth -> fullWidth }
}

/** The page being covered drifts left a little and dims. */
fun exitTransitionFor(from: NavDestination, to: NavDestination): ExitTransition {
    if (isTabSwitch(from, to)) {
        return fadeOut(tween(TAB_FADE_OUT_MS))
    }
    return slideOutHorizontally(pageTween()) { fullWidth -> -fullWidth / UNDERNEATH_SLIDE_DIVISOR } +
        fadeOut(pageTween(), targetAlpha = UNDERNEATH_ALPHA)
}

/** Going back, the page below drifts back into place and brightens. */
fun popEnterTransitionFor(from: NavDestination, to: NavDestination): EnterTransition {
    if (isTabSwitch(from, to)) {
        return fadeIn(tween(TAB_FADE_IN_MS))
    }
    return slideInHorizontally(pageTween()) { fullWidth -> -fullWidth / UNDERNEATH_SLIDE_DIVISOR } +
        fadeIn(pageTween(), initialAlpha = UNDERNEATH_ALPHA)
}

/** The page being left slides out to the right edge. */
fun popExitTransitionFor(from: NavDestination, to: NavDestination): ExitTransition {
    if (isTabSwitch(from, to)) {
        return fadeOut(tween(TAB_FADE_OUT_MS))
    }
    return slideOutHorizontally(pageTween()) { fullWidth -> fullWidth }
}
