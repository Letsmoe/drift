package dev.letsmoe.drift.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/** Where a download is in its life; [None] means the download no longer exists. */
@Serializable
enum class DownloadState {
    @SerialName("")
    None,

    @SerialName("queued")
    Queued,

    @SerialName("downloading")
    Downloading,

    @SerialName("done")
    Done,

    @SerialName("failed")
    Failed,

    /** Reported while the whole queue is paused or blocked. */
    @SerialName("paused")
    Paused,
}

/** What a downloaded collection is made of; [wireName] is the core's collection kind. */
@Serializable
enum class CollectionKind(val wireName: String) {
    @SerialName("album")
    Album("album"),

    @SerialName("playlist")
    Playlist("playlist"),

    @SerialName("local")
    LocalPlaylist("local"),

    @SerialName("likes")
    Likes("likes"),
}

/** The id of the single [CollectionKind.Likes] collection. */
const val LIKED_COLLECTION_ID = "liked"

@Serializable
data class DownloadStatus(
    val videoId: String = "",
    val state: DownloadState = DownloadState.None,
    val bytesDone: Long = 0,
    val bytesTotal: Long = 0,
    val error: String = "",
) {
    /** Download progress from 0 to 1; 0 while the size is unknown. */
    val fraction: Float
        get() {
            if (bytesTotal <= 0) {
                return 0f
            }
            return (bytesDone.toFloat() / bytesTotal.toFloat()).coerceIn(0f, 1f)
        }
}

/** A queued, running, failed or finished download with its track, so lists render offline. */
@Serializable
data class DownloadItem(
    val track: Track,
    val kind: ItemKind = ItemKind.Song,
    val status: DownloadStatus = DownloadStatus(),
    /** True when the user downloaded the track itself, not only as part of a collection. */
    val explicit: Boolean = false,
    val sizeBytes: Long = 0,
    val downloadedAt: String = "",
)

@Serializable
data class DownloadCollection(
    val kind: CollectionKind,
    val id: String,
    val title: String = "",
    /** New tracks are downloaded and removed ones deleted as the collection changes. */
    val keep: Boolean = false,
    val trackCount: Int = 0,
    val doneCount: Int = 0,
    val failedCount: Int = 0,
    val syncedAt: String = "",
) {
    val isComplete: Boolean
        get() = trackCount > 0 && doneCount >= trackCount

    val fraction: Float
        get() {
            if (trackCount <= 0) {
                return 0f
            }
            return (doneCount.toFloat() / trackCount.toFloat()).coerceIn(0f, 1f)
        }
}

@Serializable
data class DownloadSummary(
    val queued: Int = 0,
    val downloading: Int = 0,
    val failed: Int = 0,
    val done: Int = 0,
    /** True while the user paused the queue. */
    val paused: Boolean = false,
    /** True while the app disallowed downloads, e.g. Wi-Fi only on mobile data. */
    val blocked: Boolean = false,
)

/** What the core reports when one download changes. */
@Serializable
data class DownloadUpdate(
    val status: DownloadStatus = DownloadStatus(),
    val summary: DownloadSummary = DownloadSummary(),
)

/** Storage taken by finished downloads. */
@Serializable
data class StorageUsage(
    val bytes: Long = 0,
    val count: Int = 0,
)
