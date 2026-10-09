package dev.letsmoe.drift.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class RepeatMode(val wireName: String) {
    @SerialName("off")
    Off("off"),

    @SerialName("all")
    All("all"),

    @SerialName("one")
    One("one"),
}

@Serializable
data class QueueEntry(
    val id: Long,
    val track: Track,
    val kind: ItemKind = ItemKind.Song,
    /** Added by autoplay rather than the user. */
    val autoplay: Boolean = false,
)

@Serializable
data class QueueState(
    /** Play order: history, current entry, upcoming. */
    val entries: List<QueueEntry> = emptyList(),
    /** Current entry, or -1 when the queue is empty. */
    val index: Int = -1,
    val shuffle: Boolean = false,
    val repeat: RepeatMode = RepeatMode.Off,
    val autoplay: Boolean = true,
    /** Where playback started, e.g. "album:MPREb_x" or "radio". */
    val context: String = "",
) {
    val currentEntry: QueueEntry?
        get() = entries.getOrNull(index)

    val upcomingEntries: List<QueueEntry>
        get() {
            if (index < 0) {
                return emptyList()
            }
            return entries.drop(index + 1)
        }

    companion object {
        val Empty = QueueState()
    }
}

/** Result of queueNext / queueAdvance; [moved] is false when the queue had nowhere to go. */
@Serializable
data class QueueStep(
    val state: QueueState,
    val moved: Boolean = false,
)
