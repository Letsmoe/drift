package dev.letsmoe.drift.importing.platform

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.os.IBinder
import androidx.core.app.NotificationCompat
import androidx.core.app.ServiceCompat
import androidx.core.content.ContextCompat
import dev.letsmoe.drift.DriftApplication
import dev.letsmoe.drift.MainActivity
import dev.letsmoe.drift.importing.ImportController
import dev.letsmoe.drift.importing.ImportKeepAlive
import dev.letsmoe.drift.importing.ImportState
import dev.letsmoe.drift.importing.ImportStep
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.launch

private const val CHANNEL_ID = "import"
private const val NOTIFICATION_ID = 2
private const val ACTION_CANCEL = "dev.letsmoe.drift.importing.CANCEL"
private const val PERCENT = 100

/** Starts and stops [ImportService] for the import controller. */
class ServiceImportKeepAlive(private val context: Context) : ImportKeepAlive {
    override fun start() {
        ContextCompat.startForegroundService(context, Intent(context, ImportService::class.java))
    }

    override fun stop() {
        context.stopService(Intent(context, ImportService::class.java))
    }
}

/**
 * Foreground service (type dataSync) that keeps the process alive while an import runs and shows
 * its progress with a cancel action. It ends itself when the import does.
 */
class ImportService : Service() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    private var observing: Job? = null

    private val controller: ImportController
        get() = (application as DriftApplication).container.importController

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        createChannel()
        showInForeground(controller.state.value)
        if (intent?.action == ACTION_CANCEL) {
            controller.cancel()
        }
        if (observing == null) {
            observing = scope.launch { followImport() }
        }
        return START_NOT_STICKY
    }

    // Android 15 limits dataSync services to a few hours; stop instead of being killed.
    override fun onTimeout(startId: Int, fgsType: Int) {
        stopSelf()
    }

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }

    private suspend fun followImport() {
        controller.state.collect { state ->
            if (state.step != ImportStep.Running) {
                ServiceCompat.stopForeground(this, ServiceCompat.STOP_FOREGROUND_REMOVE)
                stopSelf()
                return@collect
            }
            getSystemService(NotificationManager::class.java).notify(NOTIFICATION_ID, buildNotification(state))
        }
    }

    private fun showInForeground(state: ImportState) {
        ServiceCompat.startForeground(
            this,
            NOTIFICATION_ID,
            buildNotification(state),
            ServiceInfo.FOREGROUND_SERVICE_TYPE_DATA_SYNC,
        )
    }

    private fun createChannel() {
        val channel = NotificationChannel(CHANNEL_ID, "Library import", NotificationManager.IMPORTANCE_LOW)
        getSystemService(NotificationManager::class.java).createNotificationChannel(channel)
    }

    private fun buildNotification(state: ImportState): Notification {
        val builder = NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(android.R.drawable.stat_sys_download)
            .setContentTitle("Importing your library")
            .setContentText(progressText(state))
            .setOnlyAlertOnce(true)
            .setOngoing(true)
            .setContentIntent(openAppIntent())
            .addAction(0, "Cancel", cancelIntent())
        val progress = state.progress
        if (progress == null || progress.total == 0) {
            builder.setProgress(0, 0, true)
        } else {
            builder.setProgress(PERCENT, (state.progressFraction * PERCENT).toInt(), false)
        }
        return builder.build()
    }

    private fun progressText(state: ImportState): String {
        val progress = state.progress ?: return state.phaseLabel
        if (progress.total == 0) {
            return state.phaseLabel
        }
        return "${state.phaseLabel}: ${progress.done} of ${progress.total}"
    }

    private fun openAppIntent(): PendingIntent =
        PendingIntent.getActivity(
            this,
            0,
            Intent(this, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_SINGLE_TOP),
            PendingIntent.FLAG_IMMUTABLE,
        )

    private fun cancelIntent(): PendingIntent =
        PendingIntent.getService(
            this,
            0,
            Intent(this, ImportService::class.java).setAction(ACTION_CANCEL),
            PendingIntent.FLAG_IMMUTABLE,
        )
}
