package dev.letsmoe.drift.ui.player

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.SwipeToDismissBox
import androidx.compose.material3.SwipeToDismissBoxValue
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.material3.rememberSwipeToDismissBoxState
import androidx.compose.runtime.Composable
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalHapticFeedback
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.core.model.QueueEntry
import dev.letsmoe.drift.core.model.QueueState
import dev.letsmoe.drift.core.model.RepeatMode
import dev.letsmoe.drift.player.PlayerController
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.components.Cover
import dev.letsmoe.drift.ui.components.PlayingIndicator
import dev.letsmoe.drift.ui.icons.DriftIcons
import dev.letsmoe.drift.ui.theme.DriftTheme
import sh.calvin.reorderable.ReorderableItem
import sh.calvin.reorderable.rememberReorderableLazyListState

/**
 * The play queue: the current entry in accent, then "Next up" (reorderable by the drag handle,
 * swipe to remove, tap to jump) and the autoplay entries; shuffle / repeat / autoplay toggles at
 * the bottom. Reordering applies locally while dragging and is sent to the core on drop.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun QueueSheet(onDismiss: () -> Unit) {
    val player = LocalAppContainer.current.playerController
    val queue by player.queue.collectAsStateWithLifecycle()
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        containerColor = MaterialTheme.colorScheme.surface,
    ) {
        Column(modifier = Modifier.fillMaxHeight(0.92f).navigationBarsPadding()) {
            QueueHeader(onClear = {
                player.clearQueue()
                onDismiss()
            })
            QueueList(queue = queue, player = player, modifier = Modifier.weight(1f))
            QueueFooter(queue = queue, player = player)
        }
    }
}

@Composable
private fun QueueHeader(onClear: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 20.dp, end = 8.dp, bottom = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text("Queue", style = MaterialTheme.typography.titleLarge)
        TextButton(onClick = onClear) { Text("Clear queue") }
    }
}

@Composable
private fun QueueList(queue: QueueState, player: PlayerController, modifier: Modifier = Modifier) {
    val lazyListState = rememberLazyListState()
    val haptics = LocalHapticFeedback.current
    val upcoming = remember { mutableStateListOf<QueueEntry>() }
    // Follow the core's queue unless a drag has reordered the local copy.
    LaunchedEffect(queue.entries, queue.index) {
        upcoming.clear()
        upcoming.addAll(queue.upcomingEntries)
    }
    val reorderableState = rememberReorderableLazyListState(lazyListState) { from, to ->
        moveLocally(upcoming, from.key, to.key)
    }
    val nowPlaying = queue.currentEntry
    val userEntries = upcoming.filter { entry -> !entry.autoplay }
    val autoplayEntries = upcoming.filter { entry -> entry.autoplay }

    LazyColumn(state = lazyListState, modifier = modifier) {
        if (nowPlaying != null) {
            item(key = "header-now") { QueueSectionTitle("Now playing") }
            item(key = "now-${nowPlaying.id}") { NowPlayingRow(nowPlaying, player) }
        }
        if (userEntries.isNotEmpty()) {
            item(key = "header-next") { QueueSectionTitle("Next up") }
        }
        items(userEntries, key = { entry -> entry.id }) { entry ->
            ReorderableItem(reorderableState, key = entry.id) { isDragging ->
                UpcomingRow(
                    entry = entry,
                    isDragging = isDragging,
                    player = player,
                    dragHandle = Modifier.draggableHandle(
                        onDragStarted = { haptics.performHapticFeedback(HapticFeedbackType.GestureThresholdActivate) },
                        onDragStopped = {
                            haptics.performHapticFeedback(HapticFeedbackType.GestureEnd)
                            commitMove(entry.id, upcoming, queue.index, player)
                        },
                    ),
                )
            }
        }
        if (autoplayEntries.isNotEmpty()) {
            item(key = "header-autoplay") { QueueSectionTitle("Autoplay") }
        }
        items(autoplayEntries, key = { entry -> entry.id }) { entry ->
            ReorderableItem(reorderableState, key = entry.id) { isDragging ->
                UpcomingRow(
                    entry = entry,
                    isDragging = isDragging,
                    player = player,
                    dragHandle = Modifier.draggableHandle(
                        onDragStarted = { haptics.performHapticFeedback(HapticFeedbackType.GestureThresholdActivate) },
                        onDragStopped = {
                            haptics.performHapticFeedback(HapticFeedbackType.GestureEnd)
                            commitMove(entry.id, upcoming, queue.index, player)
                        },
                    ),
                )
            }
        }
    }
}

// Keys of non-entry items (section headers) are Strings and never match.
private fun moveLocally(upcoming: MutableList<QueueEntry>, fromKey: Any, toKey: Any) {
    val fromPosition = upcoming.indexOfFirst { entry -> entry.id == fromKey }
    val toPosition = upcoming.indexOfFirst { entry -> entry.id == toKey }
    if (fromPosition < 0 || toPosition < 0) {
        return
    }
    upcoming.add(toPosition, upcoming.removeAt(fromPosition))
}

private fun commitMove(entryId: Long, upcoming: List<QueueEntry>, currentIndex: Int, player: PlayerController) {
    val upcomingPosition = upcoming.indexOfFirst { entry -> entry.id == entryId }
    if (upcomingPosition < 0) {
        return
    }
    player.move(entryId, currentIndex + 1 + upcomingPosition)
}

@Composable
private fun QueueSectionTitle(title: String) {
    Text(
        text = title,
        style = MaterialTheme.typography.titleMedium,
        modifier = Modifier.padding(start = 20.dp, end = 20.dp, top = 16.dp, bottom = 4.dp),
    )
}

@Composable
private fun EntryText(entry: QueueEntry, titleColor: Color, modifier: Modifier = Modifier) {
    Column(modifier = modifier) {
        Text(entry.track.title, style = MaterialTheme.typography.bodyLarge, color = titleColor, maxLines = 1, overflow = TextOverflow.Ellipsis)
        Text(
            text = entry.track.artistNames,
            style = MaterialTheme.typography.bodySmall,
            color = DriftTheme.colors.textDim,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

@Composable
private fun NowPlayingRow(entry: QueueEntry, player: PlayerController) {
    val isPlaying by player.isPlaying.collectAsStateWithLifecycle()
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp)
            .background(MaterialTheme.colorScheme.primary.copy(alpha = 0.12f), RoundedCornerShape(12.dp))
            .padding(horizontal = 8.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Cover(entry.track.thumbnails, size = 44.dp)
        EntryText(entry, MaterialTheme.colorScheme.primary, Modifier.weight(1f))
        PlayingIndicator(isPlaying = isPlaying, color = MaterialTheme.colorScheme.primary, modifier = Modifier.padding(end = 8.dp))
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun UpcomingRow(entry: QueueEntry, isDragging: Boolean, player: PlayerController, dragHandle: Modifier) {
    val dismissState = rememberSwipeToDismissBoxState()
    val haptics = LocalHapticFeedback.current
    LaunchedEffect(dismissState.currentValue) {
        if (dismissState.currentValue == SwipeToDismissBoxValue.EndToStart) {
            haptics.performHapticFeedback(HapticFeedbackType.ContextClick)
            player.remove(entry.id)
        }
    }
    SwipeToDismissBox(
        state = dismissState,
        enableDismissFromStartToEnd = false,
        backgroundContent = {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.error.copy(alpha = 0.85f))
                    .padding(horizontal = 24.dp),
                contentAlignment = Alignment.CenterEnd,
            ) {
                Icon(DriftIcons.Delete.outline, contentDescription = "Remove", tint = Color.White)
            }
        },
    ) {
        var rowColor = MaterialTheme.colorScheme.surface
        if (isDragging) {
            rowColor = DriftTheme.colors.surfaceHover
        }
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(rowColor)
                .clickable { player.jump(entry.id) }
                .padding(start = 20.dp, end = 4.dp, top = 8.dp, bottom = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Cover(entry.track.thumbnails, size = 44.dp)
            EntryText(entry, MaterialTheme.colorScheme.onSurface, Modifier.weight(1f))
            IconButton(onClick = {}, modifier = dragHandle) {
                Icon(DriftIcons.DragHandle.outline, contentDescription = "Reorder", tint = DriftTheme.colors.textDim)
            }
        }
    }
}

@Composable
private fun QueueFooter(queue: QueueState, player: PlayerController) {
    val accent = MaterialTheme.colorScheme.primary
    val idle = MaterialTheme.colorScheme.onSurface
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(DriftTheme.colors.surfaceRaised)
            .padding(horizontal = 24.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        IconButton(onClick = player::toggleShuffle) {
            Icon(DriftIcons.Shuffle.outline, contentDescription = "Shuffle", tint = if (queue.shuffle) accent else idle)
        }
        IconButton(onClick = player::cycleRepeat) {
            val icon = if (queue.repeat == RepeatMode.One) DriftIcons.RepeatOnce else DriftIcons.Repeat
            val tint = if (queue.repeat == RepeatMode.Off) idle else accent
            Icon(icon.outline, contentDescription = "Repeat", tint = tint)
        }
        IconButton(onClick = { player.setAutoplay(!queue.autoplay) }) {
            Icon(DriftIcons.Infinity.outline, contentDescription = "Autoplay", tint = if (queue.autoplay) accent else idle)
        }
    }
}
