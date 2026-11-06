package dev.letsmoe.drift.ui

import dev.letsmoe.drift.core.model.Track

private const val SECONDS_PER_MINUTE = 60
private const val MINUTES_PER_HOUR = 60

/** "3:07" or "1:02:09"; empty when the duration is unknown. */
fun formatDuration(totalSeconds: Int): String {
    if (totalSeconds <= 0) {
        return ""
    }
    val hours = totalSeconds / (SECONDS_PER_MINUTE * MINUTES_PER_HOUR)
    val minutes = totalSeconds % (SECONDS_PER_MINUTE * MINUTES_PER_HOUR) / SECONDS_PER_MINUTE
    val seconds = totalSeconds % SECONDS_PER_MINUTE
    if (hours > 0) {
        return "%d:%02d:%02d".format(hours, minutes, seconds)
    }
    return "%d:%02d".format(minutes, seconds)
}

/** "42 min" or "2 hr 5 min" for the summed length of [tracks]. */
fun formatTotalDuration(tracks: List<Track>): String {
    val totalSeconds = tracks.sumOf { track -> track.durationSeconds }
    val totalMinutes = Math.round(totalSeconds / SECONDS_PER_MINUTE.toDouble()).toInt()
    if (totalMinutes < MINUTES_PER_HOUR) {
        return "$totalMinutes min"
    }
    return "${totalMinutes / MINUTES_PER_HOUR} hr ${totalMinutes % MINUTES_PER_HOUR} min"
}

fun formatSongCount(count: Int): String {
    if (count == 1) {
        return "1 song"
    }
    return "$count songs"
}

private const val BYTES_PER_KILOBYTE = 1024L
private const val BYTES_PER_MEGABYTE = BYTES_PER_KILOBYTE * 1024
private const val BYTES_PER_GIGABYTE = BYTES_PER_MEGABYTE * 1024

/** "512 B", "340 KB", "12.5 MB" or "1.25 GB". */
fun formatBytes(bytes: Long): String {
    if (bytes < BYTES_PER_KILOBYTE) {
        return "$bytes B"
    }
    if (bytes < BYTES_PER_MEGABYTE) {
        return "${bytes / BYTES_PER_KILOBYTE} KB"
    }
    if (bytes < BYTES_PER_GIGABYTE) {
        return "%.1f MB".format(bytes.toDouble() / BYTES_PER_MEGABYTE)
    }
    return "%.2f GB".format(bytes.toDouble() / BYTES_PER_GIGABYTE)
}

fun describeError(error: Throwable): String = error.message ?: error.javaClass.simpleName
