package dev.letsmoe.drift.downloads

import dev.letsmoe.drift.core.DriftCore
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch

/** Tells whether the current connection is unmetered, e.g. Wi-Fi. Implemented per platform. */
interface NetworkMonitor {
    val isUnmetered: StateFlow<Boolean>
}

/** Downloads wait for an unmetered connection only while the user asked for Wi-Fi only. */
fun downloadsAllowed(isUnmetered: Boolean, downloadOnlyOnWifi: Boolean): Boolean {
    if (!downloadOnlyOnWifi) {
        return true
    }
    return isUnmetered
}

/** Keeps the core's download veto in line with the network and the "Download only on Wi-Fi" setting. */
class DownloadNetworkPolicy(
    private val core: DriftCore,
    private val downloadOnlyOnWifi: StateFlow<Boolean>,
    private val networkMonitor: NetworkMonitor,
    private val scope: CoroutineScope,
) {
    /** Applies the policy now and again whenever the network or the setting changes. */
    fun start() {
        scope.launch {
            combine(networkMonitor.isUnmetered, downloadOnlyOnWifi, ::downloadsAllowed)
                .distinctUntilChanged()
                .collect { allowed -> applyAllowed(allowed) }
        }
    }

    private suspend fun applyAllowed(allowed: Boolean) {
        try {
            core.setAllowDownloads(allowed)
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (_: Exception) {
            // The core without downloads has nothing to allow or block.
        }
    }
}
