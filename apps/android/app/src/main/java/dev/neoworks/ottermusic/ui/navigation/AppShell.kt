package dev.neoworks.ottermusic.ui.navigation

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.getValue
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavDestination.Companion.hasRoute
import androidx.navigation.NavDestination.Companion.hierarchy
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.toRoute
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.LocalBottomOverlayPadding
import dev.neoworks.ottermusic.ui.LocalNavController
import dev.neoworks.ottermusic.ui.player.MiniPlayer
import dev.neoworks.ottermusic.ui.player.NowPlayingOverlay
import dev.neoworks.ottermusic.ui.screens.album.AlbumScreen
import dev.neoworks.ottermusic.ui.screens.artist.ArtistScreen
import dev.neoworks.ottermusic.ui.screens.home.HomeScreen
import dev.neoworks.ottermusic.ui.screens.library.LibraryScreen
import dev.neoworks.ottermusic.ui.screens.likedsongs.LikedSongsScreen
import dev.neoworks.ottermusic.ui.screens.localplaylist.LocalPlaylistScreen
import dev.neoworks.ottermusic.ui.screens.playlist.PlaylistScreen
import dev.neoworks.ottermusic.ui.screens.podcast.PodcastScreen
import dev.neoworks.ottermusic.ui.screens.search.SearchScreen
import dev.neoworks.ottermusic.ui.theme.Canvas

/**
 * The app frame: the NavHost under a floating stack of mini player + tab bar, with the Now
 * Playing screen as an overlay above everything. Screens draw edge to edge; they handle the
 * status bar themselves (top bars use statusBarsPadding) and pad their scrolling content with
 * [LocalBottomOverlayPadding].
 */
@Composable
fun AppShell() {
    val navController = rememberNavController()
    val playerController = LocalAppContainer.current.playerController
    val currentTrack by playerController.currentTrack.collectAsStateWithLifecycle()
    var isNowPlayingOpen by rememberSaveable { mutableStateOf(false) }
    var overlayHeight by remember { mutableStateOf(0.dp) }
    val density = LocalDensity.current

    CompositionLocalProvider(
        LocalNavController provides navController,
        LocalBottomOverlayPadding provides overlayHeight,
    ) {
        Box(modifier = Modifier.fillMaxSize().background(Canvas)) {
            AppNavHost(navController)
            Column(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .onSizeChanged { size -> overlayHeight = with(density) { size.height.toDp() } },
            ) {
                MiniPlayer(onOpen = { isNowPlayingOpen = true })
                FloatingNavBar(
                    selectedTab = selectedTabOf(navController),
                    onSelect = { tab -> navController.switchToTab(tab.route) },
                )
            }
            NowPlayingOverlay(
                isOpen = isNowPlayingOpen && currentTrack != null,
                onClose = { isNowPlayingOpen = false },
            )
        }
    }
    BackHandler(enabled = isNowPlayingOpen) { isNowPlayingOpen = false }
}

// The tab that owns the current destination; detail pages keep the tab they were opened from.
@Composable
private fun selectedTabOf(navController: androidx.navigation.NavController): Tab {
    val backStackEntry by navController.currentBackStackEntryAsState()
    var lastTab by remember { mutableStateOf(Tab.Home) }
    val destination = backStackEntry?.destination
    if (destination != null) {
        val owningTab = tabOwning(destination)
        if (owningTab != null) {
            lastTab = owningTab
        }
    }
    return lastTab
}

private fun tabOwning(destination: androidx.navigation.NavDestination): Tab? {
    val routesByTab = listOf(
        Tab.Home to listOf(Routes.Home::class),
        Tab.Search to listOf(Routes.Search::class),
        Tab.Library to listOf(Routes.Library::class, Routes.LikedSongs::class, Routes.LocalPlaylist::class),
    )
    for ((tab, routes) in routesByTab) {
        val owns = destination.hierarchy.any { node -> routes.any { route -> node.hasRoute(route) } }
        if (owns) {
            return tab
        }
    }
    return null
}

@Composable
private fun AppNavHost(navController: androidx.navigation.NavHostController) {
    NavHost(navController = navController, startDestination = Routes.Home) {
        composable<Routes.Home> { HomeScreen() }
        composable<Routes.Search> { SearchScreen() }
        composable<Routes.Library> { LibraryScreen() }
        composable<Routes.Album> { entry -> AlbumScreen(albumId = entry.toRoute<Routes.Album>().id) }
        composable<Routes.Playlist> { entry -> PlaylistScreen(playlistId = entry.toRoute<Routes.Playlist>().id) }
        composable<Routes.Artist> { entry -> ArtistScreen(artistId = entry.toRoute<Routes.Artist>().id) }
        composable<Routes.Podcast> { entry -> PodcastScreen(podcastId = entry.toRoute<Routes.Podcast>().id) }
        composable<Routes.LikedSongs> { LikedSongsScreen() }
        composable<Routes.LocalPlaylist> { entry ->
            LocalPlaylistScreen(playlistId = entry.toRoute<Routes.LocalPlaylist>().id)
        }
    }
}
