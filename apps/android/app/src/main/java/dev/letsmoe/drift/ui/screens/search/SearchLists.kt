package dev.letsmoe.drift.ui.screens.search

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.ui.components.SectionHeader
import dev.letsmoe.drift.ui.icons.DriftIcon
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme

/** Recent searches for an empty field: tap to search again, "Clear all" forgets them. */
fun LazyListScope.recentSearches(
    recentQueries: List<String>,
    onSearch: (String) -> Unit,
    onClearAll: () -> Unit,
) {
    if (recentQueries.isEmpty()) {
        item(key = "recent-empty") {
            Text(
                text = "Search for songs, albums, artists and more.",
                style = MaterialTheme.typography.bodyMedium,
                color = DriftTheme.colors.textDim,
                textAlign = TextAlign.Center,
                modifier = Modifier.fillMaxWidth().padding(horizontal = 32.dp, vertical = 64.dp),
            )
        }
        return
    }
    item(key = "recent-header") {
        SectionHeader(title = "Recent searches", actionLabel = "Clear all", onAction = onClearAll)
    }
    items(recentQueries.size, key = { index -> "recent-$index" }) { index ->
        QueryRow(
            icon = DriftIcons.Clock,
            text = recentQueries[index],
            onClick = { onSearch(recentQueries[index]) },
        )
    }
}

/** Debounced suggestions while typing: tap to search, the arrow copies one into the field. */
fun LazyListScope.suggestionRows(
    suggestions: List<String>,
    onSearch: (String) -> Unit,
    onFillIntoField: (String) -> Unit,
) {
    items(suggestions.size, key = { index -> "suggestion-$index" }) { index ->
        QueryRow(
            icon = DriftIcons.Search,
            text = suggestions[index],
            onClick = { onSearch(suggestions[index]) },
            onFillIntoField = { onFillIntoField(suggestions[index]) },
        )
    }
}

@Composable
private fun QueryRow(
    icon: DriftIcon,
    text: String,
    onClick: () -> Unit,
    onFillIntoField: (() -> Unit)? = null,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(52.dp)
            .clickable(onClick = onClick)
            .padding(start = 16.dp, end = 4.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Icon(icon.outline, contentDescription = null, tint = DriftTheme.colors.textMuted, modifier = Modifier.size(20.dp))
        Text(
            text = text,
            style = MaterialTheme.typography.bodyLarge,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f),
        )
        if (onFillIntoField != null) {
            IconButton(onClick = onFillIntoField) {
                Icon(DriftIcons.FillIntoField.outline, contentDescription = "Use in search field", tint = DriftTheme.colors.textMuted)
            }
        }
    }
}
