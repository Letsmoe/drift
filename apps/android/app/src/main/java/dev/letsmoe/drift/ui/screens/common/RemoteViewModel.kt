package dev.letsmoe.drift.ui.screens.common

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dev.letsmoe.drift.ui.describeError
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/** A ViewModel that loads one resource with [fetch] and exposes it as a [LoadState]. */
abstract class RemoteViewModel<Data>(private val fetch: suspend () -> Data) : ViewModel() {
    private val mutableState = MutableStateFlow<LoadState<Data>>(LoadState.Loading)
    val state: StateFlow<LoadState<Data>> = mutableState.asStateFlow()

    private var loadJob: Job? = null

    init {
        load()
    }

    fun retry() {
        mutableState.value = LoadState.Loading
        load()
    }

    /** Called on the main thread after each successful load, before the content is published. */
    protected open fun onLoaded(data: Data) {}

    /** Fetches again without showing the loading state; a failure keeps what is on screen. */
    fun refreshQuietly() {
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            try {
                val data = fetch()
                onLoaded(data)
                mutableState.value = LoadState.Content(data)
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (error: Exception) {
                Log.w("RemoteViewModel", "quiet refresh failed", error)
            }
        }
    }

    private fun load() {
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            try {
                val data = fetch()
                onLoaded(data)
                mutableState.value = LoadState.Content(data)
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (error: Exception) {
                mutableState.value = LoadState.Error(describeError(error))
            }
        }
    }
}
