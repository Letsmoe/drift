package dev.letsmoe.drift.ui.screens.detail

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.ui.LocalBottomOverlayPadding
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.components.DetailTopBar
import dev.letsmoe.drift.ui.components.hasScrolledPast
import dev.letsmoe.drift.ui.screens.common.LoadState
import dev.letsmoe.drift.ui.screens.common.LoadStateContent
import dev.letsmoe.drift.ui.theme.Canvas

/** Bottom content padding so a detail list scrolls under the floating mini player and tab bar. */
@Composable
fun detailListPadding(): PaddingValues = PaddingValues(bottom = LocalBottomOverlayPadding.current)

/**
 * Frame of a detail page: the page [content] (a LazyColumn driven by [listState]) with a
 * [DetailTopBar] floating above it. The bar is transparent over the header and turns solid with
 * the title once the list scrolled past [titleScrollThreshold].
 */
@Composable
fun DetailScaffold(
    title: String,
    listState: LazyListState,
    modifier: Modifier = Modifier,
    titleScrollThreshold: Dp = 280.dp,
    topBarActions: @Composable RowScope.() -> Unit = {},
    content: @Composable () -> Unit,
) {
    val navController = LocalNavController.current
    val thresholdPixels = with(LocalDensity.current) { titleScrollThreshold.roundToPx() }
    val isScrolled = listState.hasScrolledPast(thresholdPixels)
    val barBackgroundAlpha by animateFloatAsState(
        targetValue = if (isScrolled) 1f else 0f,
        label = "detailTopBarBackground",
    )
    Box(modifier = modifier.fillMaxSize()) {
        content()
        DetailTopBar(
            title = title,
            onBack = { navController.popBackStack() },
            showTitle = isScrolled,
            modifier = Modifier.background(Canvas.copy(alpha = barBackgroundAlpha)),
            actions = topBarActions,
        )
    }
}

/**
 * [LoadStateContent] for detail pages: while loading or after a failure the page still shows a
 * back button, so the user is never stuck.
 */
@Composable
fun <Data> DetailLoadState(
    state: LoadState<Data>,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
    content: @Composable (Data) -> Unit,
) {
    val navController = LocalNavController.current
    if (state is LoadState.Content) {
        content(state.data)
        return
    }
    Box(modifier = modifier.fillMaxSize()) {
        LoadStateContent(state = state, onRetry = onRetry, content = {})
        DetailTopBar(title = "", onBack = { navController.popBackStack() }, showTitle = false)
    }
}
