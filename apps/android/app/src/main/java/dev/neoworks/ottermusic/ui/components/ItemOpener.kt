package dev.neoworks.ottermusic.ui.components

import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import dev.neoworks.ottermusic.core.model.Item
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.LocalNavController
import dev.neoworks.ottermusic.ui.navigation.openItem

/**
 * What tapping a shelf item does: songs and episodes play on their own (like the desktop app's
 * media cards) with queue [context], every other kind opens its page.
 */
@Composable
fun rememberItemOpener(context: String = "home"): (Item) -> Unit {
    val navController = LocalNavController.current
    val playerController = LocalAppContainer.current.playerController
    return remember(navController, playerController, context) {
        { item: Item ->
            navController.openItem(item) { playable ->
                val track = playable.track
                if (track != null) {
                    playerController.playOrToggle(track, listOf(track), 0, context, playable.kind)
                }
            }
        }
    }
}
