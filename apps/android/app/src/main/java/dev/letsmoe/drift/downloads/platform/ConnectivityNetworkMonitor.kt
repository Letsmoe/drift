package dev.letsmoe.drift.downloads.platform

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import dev.letsmoe.drift.downloads.NetworkMonitor
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/** [NetworkMonitor] on Android's [ConnectivityManager]: the default network counts when it is not metered. */
class ConnectivityNetworkMonitor(context: Context) : NetworkMonitor {
    private val connectivityManager = context.getSystemService(ConnectivityManager::class.java)
    private val mutableIsUnmetered = MutableStateFlow(activeNetworkIsUnmetered())

    override val isUnmetered: StateFlow<Boolean> = mutableIsUnmetered.asStateFlow()

    init {
        connectivityManager.registerDefaultNetworkCallback(
            object : ConnectivityManager.NetworkCallback() {
                override fun onCapabilitiesChanged(network: Network, capabilities: NetworkCapabilities) {
                    mutableIsUnmetered.value = isNotMetered(capabilities)
                }

                override fun onLost(network: Network) {
                    mutableIsUnmetered.value = false
                }
            },
        )
    }

    private fun activeNetworkIsUnmetered(): Boolean {
        val activeNetwork = connectivityManager.activeNetwork
        if (activeNetwork == null) {
            return false
        }
        val capabilities = connectivityManager.getNetworkCapabilities(activeNetwork)
        if (capabilities == null) {
            return false
        }
        return isNotMetered(capabilities)
    }

    private fun isNotMetered(capabilities: NetworkCapabilities): Boolean =
        capabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_NOT_METERED)
}
