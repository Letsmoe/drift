package dev.letsmoe.drift.ui.navigation

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import dev.letsmoe.drift.ui.components.glassSurface
import dev.letsmoe.drift.ui.icons.DriftIcon
import dev.letsmoe.drift.ui.icons.DriftIcons

enum class Tab(val label: String, val icon: DriftIcon, val route: Any) {
    Home("Home", DriftIcons.Home, Routes.Home),
    Search("Search", DriftIcons.Search, Routes.Search),
    Library("Library", DriftIcons.Library, Routes.Library),
}

private val TabHeight = 52.dp
private val BarPadding = 6.dp

/**
 * Floating pill tab bar (the Otter web app's bottom nav): inset 12dp from the sides and the
 * bottom, translucent, with a highlight pill that slides between tabs with a springy overshoot.
 */
@Composable
fun FloatingNavBar(
    selectedTab: Tab,
    onSelect: (Tab) -> Unit,
    modifier: Modifier = Modifier,
) {
    val tabs = Tab.entries
    Box(
        modifier = modifier
            .navigationBarsPadding()
            .padding(start = 12.dp, end = 12.dp, bottom = 12.dp)
            .fillMaxWidth()
            .glassSurface(),
    ) {
        BoxWithConstraints(modifier = Modifier.padding(BarPadding)) {
            val tabWidth = maxWidth / tabs.size
            val pillOffset by animateDpAsState(
                targetValue = tabWidth * tabs.indexOf(selectedTab),
                animationSpec = spring(dampingRatio = 0.55f, stiffness = Spring.StiffnessMediumLow),
                label = "tabPill",
            )
            Box(
                modifier = Modifier
                    .offset { IntOffset(pillOffset.roundToPx(), 0) }
                    .width(tabWidth)
                    .height(TabHeight)
                    .clip(RoundedCornerShape(22.dp))
                    .background(Color.White.copy(alpha = 0.10f)),
            )
            Row {
                for (tab in tabs) {
                    TabItem(
                        tab = tab,
                        isSelected = tab == selectedTab,
                        onClick = { onSelect(tab) },
                        modifier = Modifier.width(tabWidth),
                    )
                }
            }
        }
    }
}

@Composable
private fun TabItem(tab: Tab, isSelected: Boolean, onClick: () -> Unit, modifier: Modifier = Modifier) {
    val popScale = remember { Animatable(1f) }
    LaunchedEffect(isSelected) {
        if (isSelected) {
            popScale.snapTo(0.8f)
            popScale.animateTo(1f, spring(dampingRatio = 0.4f, stiffness = Spring.StiffnessMedium))
        }
    }
    val contentColor = if (isSelected) MaterialTheme.colorScheme.onSurface else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
    Column(
        modifier = modifier
            .height(TabHeight)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(2.dp, Alignment.CenterVertically),
    ) {
        Icon(
            imageVector = tab.icon.select(isSelected),
            contentDescription = tab.label,
            tint = contentColor,
            modifier = Modifier
                .size(22.dp)
                .scale(popScale.value),
        )
        Text(
            text = tab.label,
            style = MaterialTheme.typography.labelSmall,
            fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
            color = contentColor,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
}
