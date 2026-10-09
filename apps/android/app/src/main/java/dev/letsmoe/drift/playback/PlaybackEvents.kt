package dev.letsmoe.drift.playback

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/** Playback facts the service shares with the UI, which cannot be read from the MediaController. */
class PlaybackEvents {
    private val mutableErrorMessage = MutableStateFlow("")

    /** Empty when playback is fine. */
    val errorMessage: StateFlow<String> = mutableErrorMessage.asStateFlow()

    fun reportError(message: String) {
        mutableErrorMessage.value = message
    }

    fun clearError() {
        mutableErrorMessage.value = ""
    }
}
