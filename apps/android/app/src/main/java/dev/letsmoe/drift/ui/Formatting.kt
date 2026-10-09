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

fun describeError(error: Throwable): String = error.message ?: error.javaClass.simpleName
