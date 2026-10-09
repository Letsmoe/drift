package dev.letsmoe.drift.ui.screens.common

import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/** One page of a continued list: its entries and the token for the next page (empty on the last). */
data class ContinuationPage<Entry>(val entries: List<Entry>, val continuation: String)

/**
 * A list whose later pages are fetched with a continuation token as the user scrolls. A failed
 * extra page ends the list: the first page is already on screen.
 */
class ContinuationPager<Entry>(
    private val scope: CoroutineScope,
    private val fetchPage: suspend (token: String) -> ContinuationPage<Entry>,
) {
    private val mutableEntries = MutableStateFlow<List<Entry>>(emptyList())
    val entries: StateFlow<List<Entry>> = mutableEntries.asStateFlow()

    private val mutableHasMore = MutableStateFlow(false)
    val hasMore: StateFlow<Boolean> = mutableHasMore.asStateFlow()

    private val mutableIsLoadingMore = MutableStateFlow(false)
    val isLoadingMore: StateFlow<Boolean> = mutableIsLoadingMore.asStateFlow()

    private var continuation = ""
    private var generation = 0

    /** Starts over with the first page; a page still in flight for the old list is dropped. */
    fun reset(firstEntries: List<Entry>, firstContinuation: String) {
        generation += 1
        continuation = firstContinuation
        mutableEntries.value = firstEntries
        mutableHasMore.value = firstContinuation.isNotEmpty()
        mutableIsLoadingMore.value = false
    }

    fun loadMore() {
        if (continuation.isEmpty() || mutableIsLoadingMore.value) {
            return
        }
        val startedGeneration = generation
        mutableIsLoadingMore.value = true
        scope.launch {
            try {
                val page = fetchPage(continuation)
                if (startedGeneration == generation) {
                    appendPage(page)
                }
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (error: Exception) {
                if (startedGeneration == generation) {
                    endList()
                }
            }
        }
    }

    private fun appendPage(page: ContinuationPage<Entry>) {
        mutableEntries.value = mutableEntries.value + page.entries
        continuation = page.continuation
        mutableHasMore.value = page.continuation.isNotEmpty()
        mutableIsLoadingMore.value = false
    }

    private fun endList() {
        continuation = ""
        mutableHasMore.value = false
        mutableIsLoadingMore.value = false
    }
}
