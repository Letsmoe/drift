package dev.neoworks.ottermusic.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import dev.neoworks.ottermusic.ui.icons.OtterIcons
import dev.neoworks.ottermusic.ui.theme.OtterTheme

/** A section title; [onAction] adds a trailing "See all" style link. */
@Composable
fun SectionHeader(
    title: String,
    modifier: Modifier = Modifier,
    actionLabel: String? = null,
    onAction: (() -> Unit)? = null,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleLarge,
            maxLines = 1,
            modifier = Modifier.weight(1f),
        )
        if (actionLabel != null && onAction != null) {
            Text(
                text = actionLabel,
                style = MaterialTheme.typography.labelLarge,
                color = OtterTheme.colors.textMuted,
                modifier = Modifier
                    .clickable(onClick = onAction)
                    .padding(8.dp),
            )
        }
    }
}

@Composable
fun LoadingState(modifier: Modifier = Modifier) {
    Box(modifier = modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
    }
}

@Composable
fun ErrorState(message: String, onRetry: () -> Unit, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(32.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Icon(OtterIcons.Warning.outline, contentDescription = null, tint = MaterialTheme.colorScheme.error)
        Text(
            text = message,
            style = MaterialTheme.typography.bodyMedium,
            color = OtterTheme.colors.textMuted,
            textAlign = TextAlign.Center,
            modifier = Modifier.padding(vertical = 16.dp),
        )
        OutlinedButton(onClick = onRetry) {
            Text("Try again")
        }
    }
}

/** Apple Music style: a filled Play pill and a tonal Shuffle pill side by side. */
@Composable
fun PlayShuffleButtons(
    onPlay: () -> Unit,
    onShuffle: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Button(
            onClick = onPlay,
            modifier = Modifier.weight(1f),
            shape = RoundedCornerShape(50),
            contentPadding = ButtonDefaults.ContentPadding,
        ) {
            Icon(OtterIcons.Play.filled, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
            Text("Play", style = MaterialTheme.typography.labelLarge)
        }
        FilledTonalButton(
            onClick = onShuffle,
            modifier = Modifier.weight(1f),
            shape = RoundedCornerShape(50),
        ) {
            Icon(OtterIcons.Shuffle.outline, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
            Text("Shuffle", style = MaterialTheme.typography.labelLarge)
        }
    }
}
