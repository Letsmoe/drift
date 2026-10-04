package dev.neoworks.ottermusic.ui.screens.search

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.ui.LocalBottomOverlayPadding
import dev.neoworks.ottermusic.ui.appViewModel
import dev.neoworks.ottermusic.ui.components.ErrorState
import dev.neoworks.ottermusic.ui.components.LoadingState
import dev.neoworks.ottermusic.ui.components.SearchPill
import dev.neoworks.ottermusic.ui.theme.OtterTheme

@Composable
fun SearchScreen(modifier: Modifier = Modifier) {
    val viewModel = appViewModel { container -> SearchViewModel(container.core) }
    val query by viewModel.query.collectAsStateWithLifecycle()
    val submittedQuery by viewModel.submittedQuery.collectAsStateWithLifecycle()
    val state by viewModel.state.collectAsStateWithLifecycle()
    val focusManager = LocalFocusManager.current

    // Back first returns to the recent searches; only then does it leave the tab.
    BackHandler(enabled = query.isNotEmpty()) {
        focusManager.clearFocus()
        viewModel.clear()
    }

    Column(modifier = modifier.fillMaxSize().statusBarsPadding()) {
        Text(
            text = "Search",
            style = MaterialTheme.typography.headlineLarge,
            modifier = Modifier.padding(start = 16.dp, end = 16.dp, top = 12.dp, bottom = 12.dp),
        )
        SearchPill(
            value = query,
            onValueChange = viewModel::onQueryChange,
            onSearch = { text ->
                focusManager.clearFocus()
                viewModel.submit(text)
            },
            modifier = Modifier.padding(horizontal = 16.dp),
        )
        SearchBody(viewModel = viewModel, query = query, submittedQuery = submittedQuery, state = state)
    }
}

@Composable
private fun SearchBody(viewModel: SearchViewModel, query: String, submittedQuery: String, state: SearchState) {
    val suggestions by viewModel.suggestions.collectAsStateWithLifecycle()
    val recentQueries by viewModel.recentQueries.collectAsStateWithLifecycle()
    val filterTitle by viewModel.filterTitle.collectAsStateWithLifecycle()
    val focusManager = LocalFocusManager.current
    val isEditing = query.isNotBlank() && query.trim() != submittedQuery

    val submit: (String) -> Unit = { text ->
        focusManager.clearFocus()
        viewModel.submit(text)
    }

    if (query.isNotBlank() && !isEditing) {
        ResultsBody(state = state, viewModel = viewModel, query = submittedQuery, filterTitle = filterTitle)
        return
    }
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(top = 8.dp, bottom = LocalBottomOverlayPadding.current),
    ) {
        if (isEditing) {
            suggestionRows(suggestions, onSearch = submit, onFillIntoField = viewModel::fillIntoField)
            return@LazyColumn
        }
        recentSearches(recentQueries, onSearch = submit, onClearAll = viewModel::clearRecentQueries)
    }
}

@Composable
private fun ResultsBody(state: SearchState, viewModel: SearchViewModel, query: String, filterTitle: String) {
    when (state) {
        SearchState.Idle, SearchState.Loading -> LoadingState()
        is SearchState.Failed -> ErrorState(message = state.message, onRetry = viewModel::retry)
        is SearchState.Done -> {
            if (state.shelves.isEmpty()) {
                NoResults(query)
                return
            }
            val content = SearchResultsContent(state.shelves, query, filterTitle, viewModel::selectFilter)
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(top = 4.dp, bottom = LocalBottomOverlayPadding.current),
            ) {
                searchResults(content)
            }
        }
    }
}

@Composable
private fun NoResults(query: String) {
    Box(modifier = Modifier.fillMaxWidth().height(240.dp), contentAlignment = Alignment.Center) {
        Text(
            text = "No results found for \"$query\".",
            style = MaterialTheme.typography.bodyMedium,
            color = OtterTheme.colors.textDim,
            textAlign = TextAlign.Center,
            modifier = Modifier.fillMaxWidth().padding(32.dp),
        )
    }
}
