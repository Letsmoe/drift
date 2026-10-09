package dev.neoworks.ottermusic

import android.app.Application
import android.content.Context
import coil3.ImageLoader
import coil3.PlatformContext
import coil3.SingletonImageLoader
import coil3.memory.MemoryCache
import coil3.request.crossfade
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.library.ImpressionLogger
import dev.neoworks.ottermusic.library.LibraryState
import dev.neoworks.ottermusic.player.PlayerController
import dev.neoworks.ottermusic.playback.PlaybackEvents
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

private const val IMAGE_MEMORY_CACHE_PERCENT = 0.25

/** Hand-wired dependencies; created once with the process. */
class AppContainer(context: Context) {
    val applicationScope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    val core = OtterCore(context.filesDir.absolutePath)
    val playbackEvents = PlaybackEvents()
    val playerController = PlayerController(context, core, playbackEvents, applicationScope)
    val libraryState = LibraryState(core, applicationScope)
    val impressionLogger = ImpressionLogger(core, applicationScope)
}

class OtterApplication : Application(), SingletonImageLoader.Factory {
    lateinit var container: AppContainer
        private set

    // Covers are small and reused across screens: keep a generous memory cache and fade them in.
    override fun newImageLoader(context: PlatformContext): ImageLoader =
        ImageLoader.Builder(context)
            .crossfade(150)
            .memoryCache { MemoryCache.Builder().maxSizePercent(context, IMAGE_MEMORY_CACHE_PERCENT).build() }
            .build()

    override fun onCreate() {
        super.onCreate()
        container = AppContainer(this)
        container.playerController.connect()
        container.libraryState.load()
    }
}
