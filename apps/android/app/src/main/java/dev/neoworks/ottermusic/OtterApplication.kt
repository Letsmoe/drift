package dev.neoworks.ottermusic

import android.app.Application
import android.content.Context
import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.library.ImpressionLogger
import dev.neoworks.ottermusic.library.LibraryState
import dev.neoworks.ottermusic.player.PlayerController
import dev.neoworks.ottermusic.playback.PlaybackEvents
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

/** Hand-wired dependencies; created once with the process. */
class AppContainer(context: Context) {
    val applicationScope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    val core = OtterCore(context.filesDir.absolutePath)
    val playbackEvents = PlaybackEvents()
    val playerController = PlayerController(context, core, playbackEvents, applicationScope)
    val libraryState = LibraryState(core, applicationScope)
    val impressionLogger = ImpressionLogger(core, applicationScope)
}

class OtterApplication : Application() {
    lateinit var container: AppContainer
        private set

    override fun onCreate() {
        super.onCreate()
        container = AppContainer(this)
        container.playerController.connect()
        container.libraryState.load()
    }
}
