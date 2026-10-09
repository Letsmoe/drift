package dev.neoworks.ottermusic.ui.components

import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.LocalIndication
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.interaction.InteractionSource
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.interaction.collectIsPressedAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.composed
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalHapticFeedback

private const val DEFAULT_PRESSED_SCALE = 0.96f

/** Shrinks the content while [interactionSource] reports a press; read in the graphics layer, so it never recomposes. */
fun Modifier.pressScale(
    interactionSource: InteractionSource,
    pressedScale: Float = DEFAULT_PRESSED_SCALE,
): Modifier = composed {
    val isPressed by interactionSource.collectIsPressedAsState()
    val scale = animateFloatAsState(
        targetValue = if (isPressed) pressedScale else 1f,
        animationSpec = spring(dampingRatio = Spring.DampingRatioNoBouncy, stiffness = Spring.StiffnessMedium),
        label = "pressScale",
    )
    graphicsLayer {
        scaleX = scale.value
        scaleY = scale.value
    }
}

/**
 * A click target that shrinks slightly while pressed. [onLongClick] adds a long-press with haptic
 * feedback. Put shape clipping and backgrounds after this modifier so the ripple stays on top.
 */
@OptIn(ExperimentalFoundationApi::class)
fun Modifier.pressScaleClickable(
    onClick: () -> Unit,
    onLongClick: (() -> Unit)? = null,
    pressedScale: Float = DEFAULT_PRESSED_SCALE,
): Modifier = composed {
    val interactionSource = remember { MutableInteractionSource() }
    val haptics = LocalHapticFeedback.current
    pressScale(interactionSource, pressedScale).combinedClickable(
        interactionSource = interactionSource,
        indication = LocalIndication.current,
        onClick = onClick,
        onLongClick = onLongClick?.let { longClick ->
            {
                haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                longClick()
            }
        },
    )
}
