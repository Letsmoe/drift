package dev.neoworks.ottermusic

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.CompositionLocalProvider
import androidx.core.content.ContextCompat
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.lifecycleScope
import androidx.lifecycle.repeatOnLifecycle
import dev.neoworks.ottermusic.ui.LocalAppContainer
import dev.neoworks.ottermusic.ui.navigation.AppShell
import dev.neoworks.ottermusic.ui.theme.OtterMusicTheme
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {
    private var hasAskedForNotifications = false

    private val notificationPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { }

    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)
        preferHighestRefreshRate()
        val container = (application as OtterApplication).container
        setContent {
            CompositionLocalProvider(LocalAppContainer provides container) {
                OtterMusicTheme {
                    AppShell()
                }
            }
        }
        askForNotificationsOnFirstPlay(container)
    }

    // Ask for the display mode with the highest refresh rate at the current resolution (e.g. 120 Hz).
    private fun preferHighestRefreshRate() {
        val currentDisplay = ContextCompat.getDisplayOrDefault(this)
        val currentMode = currentDisplay.mode
        val fastestMode = currentDisplay.supportedModes
            .filter { mode ->
                mode.physicalWidth == currentMode.physicalWidth &&
                    mode.physicalHeight == currentMode.physicalHeight
            }
            .maxByOrNull { mode -> mode.refreshRate }
            ?: return
        val attributes = window.attributes
        attributes.preferredDisplayModeId = fastestMode.modeId
        window.attributes = attributes
    }

    // The media notification needs the permission on Android 13+; ask when playback first starts.
    private fun askForNotificationsOnFirstPlay(container: AppContainer) {
        lifecycleScope.launch {
            repeatOnLifecycle(Lifecycle.State.STARTED) {
                container.playerController.playRequests.collect { requestNotificationPermission() }
            }
        }
    }

    private fun requestNotificationPermission() {
        if (hasAskedForNotifications || Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) {
            return
        }
        hasAskedForNotifications = true
        val granted = ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS) ==
            PackageManager.PERMISSION_GRANTED
        if (!granted) {
            notificationPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
        }
    }
}
