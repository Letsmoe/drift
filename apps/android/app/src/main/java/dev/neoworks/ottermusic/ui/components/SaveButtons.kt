package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.showToastOnFailure
import dev.neoworks.ottermusic.ui.theme.OtterTheme
import kotlinx.coroutines.launch

/** Observes whether [item] is in the library and toggles it; returns the state and the toggle action. */
@Composable
private fun rememberSavedToggle(item: Item): Pair<Boolean, () -> Unit> {
    val library = LocalAppContainer.current.libraryState
    val savedItems by library.savedItems.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val isSaved = savedItems.any { saved -> saved.item.kind == item.kind && saved.item.id == item.id }
    val toggle: () -> Unit = {
        scope.launch { showToastOnFailure(context) { library.toggleSaved(item) } }
    }
    return isSaved to toggle
}

/** Round add / saved check button for albums and playlists. */
@Composable
fun SaveButton(item: Item, modifier: Modifier = Modifier) {
    val (isSaved, toggle) = rememberSavedToggle(item)
    IconButton(onClick = toggle, modifier = modifier) {
        if (isSaved) {
            Icon(
                OtterIcons.Check.outline,
                contentDescription = "Remove ${item.title} from your library",
                tint = MaterialTheme.colorScheme.primary,
            )
            return@IconButton
        }
        Icon(
            OtterIcons.Add.outline,
            contentDescription = "Save ${item.title} to your library",
            tint = OtterTheme.colors.textMuted,
        )
    }
}

/** Follow / Following pill for artists and podcasts. */
@Composable
fun FollowButton(item: Item, modifier: Modifier = Modifier) {
    val (isSaved, toggle) = rememberSavedToggle(item)
    val label = if (isSaved) "Following" else "Follow"
    OutlinedButton(
        onClick = toggle,
        modifier = modifier,
        shape = RoundedCornerShape(50),
    ) {
        Text(text = label, style = MaterialTheme.typography.labelLarge, modifier = Modifier.padding(horizontal = 4.dp))
    }
}
