package dev.letsmoe.drift.ui.player

import androidx.compose.animation.AnimatedVisibilityScope
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionScope
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier

private const val SHARED_COVER_KEY = "player-cover"

/** Links the mini player cover and the Now Playing cover into one element that flies between them. */
@Suppress("ModifierFactoryExtensionFunction")
@OptIn(ExperimentalSharedTransitionApi::class)
@Composable
fun rememberSharedCoverModifier(
    sharedTransitionScope: SharedTransitionScope,
    animatedVisibilityScope: AnimatedVisibilityScope,
): Modifier = with(sharedTransitionScope) {
    Modifier.sharedElement(
        sharedContentState = rememberSharedContentState(key = SHARED_COVER_KEY),
        animatedVisibilityScope = animatedVisibilityScope,
    )
}
