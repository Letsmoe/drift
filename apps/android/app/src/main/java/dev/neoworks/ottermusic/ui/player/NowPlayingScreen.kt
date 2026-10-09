package dev.neoworks.ottermusic.ui.player

import androidx.compose.animation.AnimatedVisibility
import androidx.activity.compose.PredictiveBackHandler
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionScope
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.spring
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalWindowInfo
import androidx.compose.ui.platform.LocalHapticFeedback
import kotlin.coroutines.cancellation.CancellationException
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.ExperimentalFoundationApi
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
import kotlinx.coroutines.launch

private const val DISMISS_DISTANCE_FRACTION = 0.25f
private const val DISMISS_VELOCITY = 1800f
private const val BACK_SCALE_REDUCTION = 0.1f
private val PlayerCornerRadius = 28.dp

/**
 * Slides the Now Playing screen up from the bottom while [isOpen]; its cover is shared with the
 * mini player. System back (including the predictive gesture) shrinks it away.
 */
@OptIn(ExperimentalSharedTransitionApi::class)
@Composable
fun NowPlayingOverlay(isOpen: Boolean, sharedTransitionScope: SharedTransitionScope, onClose: () -> Unit) {
    val backProgress = remember { Animatable(0f) }
    LaunchedEffect(isOpen) {
        if (isOpen) {
            backProgress.snapTo(0f)
        }
    }
    PredictiveBackHandler(enabled = isOpen) { progressEvents ->
        try {
            progressEvents.collect { event -> backProgress.snapTo(event.progress) }
            onClose()
        } catch (cancellation: CancellationException) {
            backProgress.animateTo(0f)
            throw cancellation
        }
    }
    AnimatedVisibility(
        visible = isOpen,
        enter = slideInVertically { fullHeight -> fullHeight },
        exit = slideOutVertically { fullHeight -> fullHeight },
    ) {
        val sharedCover = rememberSharedCoverModifier(sharedTransitionScope, this)
        NowPlayingScreen(
            onClose = onClose,
            sharedCover = sharedCover,
            modifier = Modifier.graphicsLayer {
                val scale = 1f - BACK_SCALE_REDUCTION * backProgress.value
                scaleX = scale
                scaleY = scale
            },
        )
    }
}

/**
 * Full-screen player: gradient from the cover color, big cover, title, seek bar, transport
 * controls, autoplay and queue. Dragging down follows the finger and dismisses on distance or fling.
 */
@Suppress("ModifierParameter")
@Composable
fun NowPlayingScreen(onClose: () -> Unit, sharedCover: Modifier, modifier: Modifier = Modifier) {
    val player = LocalAppContainer.current.playerController
    val queue by player.queue.collectAsStateWithLifecycle()
    val entry = queue.currentEntry
    if (entry == null) {
        return
    }
    val dismissOffset = remember { Animatable(0f) }
    val scope = rememberCoroutineScope()
    val screenHeightPixels = LocalWindowInfo.current.containerSize.height.toFloat()
    val coverColor by rememberCoverColor(entry.track.thumbnails.bestFor(128)?.url)
    var showQueue by remember { mutableStateOf(false) }
    var showMore by remember { mutableStateOf(false) }

    Box(
        modifier = modifier
            .graphicsLayer {
                translationY = dismissOffset.value
                shape = RoundedCornerShape(topStart = PlayerCornerRadius, topEnd = PlayerCornerRadius)
                clip = true
            }
            .fillMaxSize()
            .drawBehind {
                drawRect(Brush.verticalGradient(listOf(coverColor, Canvas), endY = size.height * 0.75f))
                drawRect(Canvas.copy(alpha = 0.35f))
            }
            .draggable(
                orientation = Orientation.Vertical,
                state = rememberDraggableState { delta ->
                    scope.launch { dismissOffset.snapTo((dismissOffset.value + delta).coerceAtLeast(0f)) }
                },
                onDragStopped = { velocity ->
                    scope.launch { settleDismissDrag(dismissOffset, velocity, screenHeightPixels, onClose) }
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
            NowPlayingCover(entry = entry, sharedCover = sharedCover)
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

// Closing keeps the dragged offset so the exit slide continues from where the finger left off.
private suspend fun settleDismissDrag(
    dismissOffset: Animatable<Float, *>,
    velocity: Float,
    screenHeightPixels: Float,
    onClose: () -> Unit,
) {
    val isFarEnough = dismissOffset.value > screenHeightPixels * DISMISS_DISTANCE_FRACTION
    if (isFarEnough || velocity > DISMISS_VELOCITY) {
        onClose()
        return
    }
    dismissOffset.animateTo(0f, spring(dampingRatio = Spring.DampingRatioLowBouncy, stiffness = Spring.StiffnessMedium), velocity)
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

@Suppress("ModifierParameter")
@Composable
private fun NowPlayingCover(entry: QueueEntry, sharedCover: Modifier) {
    val player = LocalAppContainer.current.playerController
    val isPlaying by player.isPlaying.collectAsStateWithLifecycle()
    // Apple Music feel: the cover shrinks while paused.
    val coverScale by animateFloatAsState(
        targetValue = if (isPlaying) 1f else 0.86f,
        animationSpec = spring(dampingRatio = 0.7f, stiffness = Spring.StiffnessLow),
        label = "coverScale",
    )
    CoverFill(
        thumbnails = entry.track.thumbnails,
        requestSize = 400.dp,
        shape = RoundedCornerShape(14.dp),
        modifier = sharedCover
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
    val haptics = LocalHapticFeedback.current

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
        LikeButton(
            isLiked = isLiked,
            onToggle = {
                haptics.performHapticFeedback(if (isLiked) HapticFeedbackType.ToggleOff else HapticFeedbackType.ToggleOn)
                container.libraryState.toggleLike(entry.track, entry.kind)
            },
        )
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

/** Heart that pops with a springy overshoot when it turns on. */
@Composable
private fun LikeButton(isLiked: Boolean, onToggle: () -> Unit) {
    val popScale = remember { Animatable(1f) }
    LaunchedEffect(isLiked) {
        if (isLiked) {
            popScale.snapTo(0.7f)
            popScale.animateTo(1f, spring(dampingRatio = Spring.DampingRatioMediumBouncy, stiffness = Spring.StiffnessMedium))
        }
    }
    IconButton(onClick = onToggle) {
        Icon(
            imageVector = OtterIcons.Heart.select(isLiked),
            contentDescription = if (isLiked) "Unlike" else "Like",
            tint = if (isLiked) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface,
            modifier = Modifier.graphicsLayer {
                scaleX = popScale.value
                scaleY = popScale.value
            },
        )
    }
}
