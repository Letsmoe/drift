package dev.letsmoe.drift.ui.screens.settings

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.settings.CrossfadeSeconds
import dev.letsmoe.drift.settings.SettingsStore
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalBottomOverlayPadding
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.components.DetailTopBar
import dev.letsmoe.drift.ui.components.SettingsSection
import dev.letsmoe.drift.ui.components.SettingsSliderRow
import dev.letsmoe.drift.ui.components.SettingsSwitchRow

/** App preferences. A feature adds its settings as rows in a section here. */
@Composable
fun SettingsScreen(modifier: Modifier = Modifier) {
    val settings = LocalAppContainer.current.settings
    val navController = LocalNavController.current

    Column(modifier = modifier.fillMaxSize()) {
        DetailTopBar(title = "Settings", onBack = { navController.popBackStack() })
        Column(
            modifier = Modifier
                .verticalScroll(rememberScrollState())
                .padding(bottom = LocalBottomOverlayPadding.current),
        ) {
            PlaybackSection(settings)
            DownloadsSection(settings)
        }
    }
}

@Composable
private fun PlaybackSection(settings: SettingsStore) {
    val crossfadeEnabled by settings.crossfadeEnabled.collectAsStateWithLifecycle()
    val crossfadeSeconds by settings.crossfadeSeconds.collectAsStateWithLifecycle()

    SettingsSection(title = "Playback") {
        SettingsSwitchRow(
            title = "Smooth transitions",
            description = "The next song fades in while the current one fades out.",
            isChecked = crossfadeEnabled,
            onCheckedChange = settings::setCrossfadeEnabled,
        )
        if (crossfadeEnabled) {
            SettingsSliderRow(
                title = "Crossfade length",
                value = crossfadeSeconds,
                valueRange = CrossfadeSeconds.MIN..CrossfadeSeconds.MAX,
                valueLabel = "$crossfadeSeconds s",
                onValueChange = settings::setCrossfadeSeconds,
            )
        }
    }
}

@Composable
private fun DownloadsSection(settings: SettingsStore) {
    val downloadOnlyOnWifi by settings.downloadOnlyOnWifi.collectAsStateWithLifecycle()

    SettingsSection(title = "Downloads") {
        SettingsSwitchRow(
            title = "Download only on Wi-Fi",
            description = "Downloads wait until you are on Wi-Fi.",
            isChecked = downloadOnlyOnWifi,
            onCheckedChange = settings::setDownloadOnlyOnWifi,
        )
    }
}
