package dev.letsmoe.drift.ui.screens.importing

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.letsmoe.drift.importing.ImportStep
import dev.letsmoe.drift.ui.LocalAppContainer
import dev.letsmoe.drift.ui.LocalBottomOverlayPadding
import dev.letsmoe.drift.ui.LocalNavController
import dev.letsmoe.drift.ui.components.DetailTopBar
import dev.letsmoe.drift.ui.navigation.openLocalPlaylist

private val ExportMimeTypes = arrayOf(
    "application/zip",
    "application/x-zip-compressed",
    "application/json",
    "text/csv",
    "text/comma-separated-values",
    "application/vnd.ms-excel",
    "application/octet-stream",
)

/** The Spotify import flow: choose a source, pick what to bring over, watch it run, read the summary. */
@Composable
fun ImportScreen() {
    val controller = LocalAppContainer.current.importController
    val navController = LocalNavController.current
    val state by controller.state.collectAsStateWithLifecycle()
    val filePicker = rememberLauncherForActivityResult(ActivityResultContracts.OpenDocument()) { uri ->
        if (uri != null) {
            controller.previewFile(uri.toString())
        }
    }
    val startAfterNotificationChoice = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) {
        controller.start()
    }
    val context = LocalContext.current

    fun leave() {
        controller.dismissFinished()
        navController.popBackStack()
    }

    fun startImport() {
        if (isNotificationPermissionMissing(context)) {
            startAfterNotificationChoice.launch(Manifest.permission.POST_NOTIFICATIONS)
            return
        }
        controller.start()
    }

    BackHandler {
        if (state.step == ImportStep.Preview) {
            controller.restart()
            return@BackHandler
        }
        leave()
    }

    Column(modifier = Modifier.fillMaxSize()) {
        DetailTopBar(title = "Import from Spotify", onBack = ::leave)
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(start = 16.dp, end = 16.dp, bottom = LocalBottomOverlayPadding.current + 16.dp),
        ) {
            when (state.step) {
                ImportStep.Source -> sourceStep(
                    errorMessage = state.errorMessage,
                    onPreviewLink = controller::previewLink,
                    onChooseFile = { filePicker.launch(ExportMimeTypes) },
                )
                ImportStep.Loading -> loadingStep()
                ImportStep.Preview -> previewStep(state, controller, onBack = controller::restart, onImport = ::startImport)
                ImportStep.Running -> runningStep(state, onCancel = controller::cancel)
                ImportStep.Done -> doneStep(
                    state = state,
                    onOpenPlaylist = { playlistId ->
                        controller.dismissFinished()
                        navController.openLocalPlaylist(playlistId)
                    },
                    onDone = ::leave,
                )
            }
        }
    }
}

// The foreground service runs without it, but its progress notification stays hidden.
private fun isNotificationPermissionMissing(context: android.content.Context): Boolean {
    if (Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) {
        return false
    }
    val granted = ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) ==
        PackageManager.PERMISSION_GRANTED
    return !granted
}

