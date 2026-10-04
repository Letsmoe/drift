package dev.neoworks.ottermusic.ui.screens.likedsongs

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier

// Placeholder: replaced by the real screen.
@Composable
fun LikedSongsScreen(modifier: Modifier = Modifier) {
    Box(modifier = modifier.fillMaxSize().statusBarsPadding(), contentAlignment = Alignment.Center) {
        Text("Liked Songs", style = MaterialTheme.typography.titleLarge)
    }
}
