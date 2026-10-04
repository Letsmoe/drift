package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState

/**
 * Calls [onNearEnd] when the last [bufferItems] items of the list come into view, and again
 * after the list grew while it is still near the end. Does nothing while [isEnabled] is false.
 */
@Composable
fun LazyListState.OnNearEnd(
    isEnabled: Boolean,
    bufferItems: Int = 6,
    onNearEnd: () -> Unit,
) {
    val state = this
    val itemCount by remember(state) { derivedStateOf { state.layoutInfo.totalItemsCount } }
    val isNearEnd by remember(state, bufferItems) {
        derivedStateOf {
            val lastVisibleIndex = state.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisibleIndex >= state.layoutInfo.totalItemsCount - 1 - bufferItems
        }
    }
    val currentOnNearEnd by rememberUpdatedState(onNearEnd)
    LaunchedEffect(isEnabled, isNearEnd, itemCount) {
        if (isEnabled && isNearEnd) {
            currentOnNearEnd()
        }
    }
}
