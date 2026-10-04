package dev.neoworks.ottermusic.ui

import android.content.Context
import android.widget.Toast
import kotlinx.coroutines.CancellationException

/** Runs [block] and shows a toast with the error message when it fails. */
suspend fun showToastOnFailure(context: Context, block: suspend () -> Unit) {
    try {
        block()
    } catch (cancellation: CancellationException) {
        throw cancellation
    } catch (error: Exception) {
        Toast.makeText(context, describeError(error), Toast.LENGTH_SHORT).show()
    }
}
