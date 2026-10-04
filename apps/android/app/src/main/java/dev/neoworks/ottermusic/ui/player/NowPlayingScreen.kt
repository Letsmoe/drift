package dev.neoworks.ottermusic.ui.player

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.Orientation
import androidx.compose.foundation.gestures.draggable
import androidx.compose.foundation.gestures.rememberDraggableState
import androidx.compose.foundation.basicMarquee
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.neoworks.ottermusic.core.model.Artist
import dev.neoworks.ottermusic.core.model.QueueEntry
import dev.neoworks.ottermusic.core.model.bestFor
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.LocalNavController
import dev.neoworks.ottermusic.ui.components.CoverFill
import dev.neoworks.ottermusic.ui.components.TrackActionSheet
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.navigation.openAlbum
import dev.neoworks.ottermusic.ui.navigation.openArtist
import dev.neoworks.ottermusic.ui.theme.Canvas
import dev.neoworks.ottermusic.ui.theme.OtterTheme
import kotlin.math.roundToInt
import kotlinx.coroutines.launch

private const val DISMISS_DRAG_PIXELS = 360f

/** Slides the Now Playing screen up from the bottom while [isOpen]. */
@Composable
fun NowPlayingOverlay(isOpen: Boolean, onClose: () -> Unit) {
    AnimatedVisibility(
        visible = isOpen,
        enter = slideInVertically { fullHeight -> fullHeight },
        exit = slideOutVertically { fullHeight -> fullHeight },
    ) {
        NowPlayingScreen(onClose = onClose)
    }
}

/**
 * Full-screen player: gradient from the cover color, big cover, title, seek bar, transport
 * controls, autoplay and queue. Dragging down dismisses it.
 */
@Composable
fun NowPlayingScreen(onClose: () -> Unit) {
    val player = LocalAppContainer.current.playerController
    val queue by player.queue.collectAsStateWithLifecycle()
    val entry = queue.currentEntry
    if (entry == null) {
        return
    }
    val dismissOffset = remember { Animatable(0f) }
    val scope = rememberCoroutineScope()
    val coverColor by rememberCoverColor(entry.track.thumbnails.bestFor(128)?.url)
    var showQueue by remember { mutableStateOf(false) }
    var showMore by remember { mutableStateOf(false) }

    Box(
        modifier = Modifier
            .offset { IntOffset(0, dismissOffset.value.roundToInt()) }
            .fillMaxSize()
            .background(Brush.verticalGradient(listOf(coverColor, Canvas), endY = 1800f))
            .background(Canvas.copy(alpha = 0.35f))
            .draggable(
                orientation = Orientation.Vertical,
                state = rememberDraggableState { delta ->
                    scope.launch { dismissOffset.snapTo((dismissOffset.value + delta).coerceAtLeast(0f)) }
                },
                onDragStopped = {
                    scope.launch { settleDismissDrag(dismissOffset, onClose) }
                },
            ),
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .statusBarsPadding()
                .navigationBarsPadding()
                .padding(horizontal = 24.dp),
        ) {
            NowPlayingTopBar(
                playContext = queue.context,
                onClose = onClose,
                onMore = { showMore = true },
            )
            Spacer(Modifier.weight(1f))
            NowPlayingCover(entry = entry)
            Spacer(Modifier.weight(1f))
            TrackInfoRow(entry = entry, onClose = onClose)
            Spacer(Modifier.height(8.dp))
            SeekBar()
            Spacer(Modifier.height(8.dp))
            TransportControls()
            Spacer(Modifier.height(12.dp))
            NowPlayingBottomRow(onOpenQueue = { showQueue = true })
            Spacer(Modifier.height(8.dp))
        }
    }

    if (showQueue) {
        QueueSheet(onDismiss = { showQueue = false })
    }
    if (showMore) {
        MoreSheet(entry = entry, onDismiss = { showMore = false }, onClose = onClose)
    }
}

private suspend fun settleDismissDrag(dismissOffset: Animatable<Float, *>, onClose: () -> Unit) {
    if (dismissOffset.value > DISMISS_DRAG_PIXELS) {
        onClose()
        dismissOffset.snapTo(0f)
        return
    }
    dismissOffset.animateTo(0f)
}

@Composable
private fun MoreSheet(entry: QueueEntry, onDismiss: () -> Unit, onClose: () -> Unit) {
    val navController = LocalNavController.current
    TrackActionSheet(
        track = entry.track,
        kind = entry.kind,
        onDismiss = onDismiss,
        onOpenAlbum = { album ->
            onClose()
            navController.openAlbum(album.id)
        },
        onOpenArtist = { artist ->
            onClose()
            navController.openArtist(artist.id)
        },
    )
}

@Composable
private fun NowPlayingTopBar(playContext: String, onClose: () -> Unit, onMore: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(56.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        IconButton(onClick = onClose) {
            Icon(OtterIcons.ChevronDown.outline, contentDescription = "Close player")
        }
        Column(modifier = Modifier.weight(1f), horizontalAlignment = Alignment.CenterHorizontally) {
            Text(
                text = "PLAYING FROM",
                style = MaterialTheme.typography.labelSmall,
                color = OtterTheme.colors.textMuted,
                letterSpacing = 1.2.sp,
            )
            Text(
                text = contextLabel(playContext),
                style = MaterialTheme.typography.labelLarge,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        IconButton(onClick = onMore) {
            Icon(OtterIcons.More.outline, contentDescription = "More")
        }
    }
}

@Composable
private fun NowPlayingCover(entry: QueueEntry) {
    val player = LocalAppContainer.current.playerController
    val isPlaying by player.isPlaying.collectAsStateWithLifecycle()
    // Apple Music feel: the cover shrinks while paused.
    val coverScale by animateFloatAsState(if (isPlaying) 1f else 0.86f, label = "coverScale")
    CoverFill(
        thumbnails = entry.track.thumbnails,
        requestSize = 400.dp,
        shape = RoundedCornerShape(14.dp),
        modifier = Modifier
            .fillMaxWidth()
            .aspectRatio(1f)
            .graphicsLayer {
                scaleX = coverScale
                scaleY = coverScale
            }
            .shadow(28.dp, RoundedCornerShape(14.dp), ambientColor = Color.Black, spotColor = Color.Black),
    )
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun TrackInfoRow(entry: QueueEntry, onClose: () -> Unit) {
    val container = LocalAppContainer.current
    val navController = LocalNavController.current
    val likedVideoIds by container.libraryState.likedVideoIds.collectAsStateWithLifecycle()
    LaunchedEffect(entry.track.videoId) { container.libraryState.ensureRatings(listOf(entry.track.videoId)) }
    val isLiked = entry.track.videoId in likedVideoIds

    Row(verticalAlignment = Alignment.CenterVertically) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = entry.track.title,
                style = MaterialTheme.typography.headlineSmall,
                maxLines = 1,
                modifier = Modifier.basicMarquee(),
            )
            ArtistLinks(
                artists = entry.track.artists,
                onOpenArtist = { artist ->
                    onClose()
                    navController.openArtist(artist.id)
                },
            )
        }
        IconButton(onClick = { container.libraryState.toggleLike(entry.track, entry.kind) }) {
            Icon(
                imageVector = OtterIcons.Heart.select(isLiked),
                contentDescription = "Like",
                tint = if (isLiked) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
            )
        }
    }
}

@Composable
private fun ArtistLinks(artists: List<Artist>, onOpenArtist: (Artist) -> Unit) {
    Row {
        artists.forEachIndexed { artistIndex, artist ->
            val separator = if (artistIndex < artists.lastIndex) ", " else ""
            val canOpen = artist.id.isNotEmpty()
            Text(
                text = artist.name + separator,
                style = MaterialTheme.typography.bodyLarge,
                color = OtterTheme.colors.textMuted,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.clickable(enabled = canOpen) { onOpenArtist(artist) },
            )
        }
    }
}
