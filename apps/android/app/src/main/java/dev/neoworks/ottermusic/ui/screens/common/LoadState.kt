package dev.neoworks.ottermusic.ui.screens.common

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import dev.neoworks.ottermusic.ui.components.ErrorState
import dev.neoworks.ottermusic.ui.components.LoadingState

/** What a screen that loads one resource shows. */
sealed interface LoadState<out Data> {
    data object Loading : LoadState<Nothing>

    data class Error(val message: String) : LoadState<Nothing>

    data class Content<Data>(val data: Data) : LoadState<Data>
}

/** Renders the loading and error states itself and hands loaded data to [content]. */
@Composable
fun <Data> LoadStateContent(
    state: LoadState<Data>,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
    content: @Composable (Data) -> Unit,
) {
    when (state) {
        LoadState.Loading -> LoadingState(modifier)
        is LoadState.Error -> ErrorState(message = state.message, onRetry = onRetry, modifier = modifier)
        is LoadState.Content -> content(state.data)
    }
}
