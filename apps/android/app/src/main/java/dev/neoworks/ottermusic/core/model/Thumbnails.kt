package dev.neoworks.ottermusic.core.model

/**
 * The smallest thumbnail that is at least [targetPixels] wide, or the largest one when none is
 * big enough. Thumbnails without a width are treated as the fallback.
 */
fun List<Thumbnail>.bestFor(targetPixels: Int): Thumbnail? {
    if (isEmpty()) {
        return null
    }
    val sizedThumbnails = filter { thumbnail -> thumbnail.width > 0 }.sortedBy { thumbnail -> thumbnail.width }
    if (sizedThumbnails.isEmpty()) {
        return first()
    }
    val bigEnough = sizedThumbnails.firstOrNull { thumbnail -> thumbnail.width >= targetPixels }
    if (bigEnough != null) {
        return bigEnough
    }
    return sizedThumbnails.last()
}

fun List<Thumbnail>.largestUrl(): String? = bestFor(Int.MAX_VALUE)?.url

fun Track.toItem(kind: ItemKind = ItemKind.Song): Item =
    Item(kind = kind, id = videoId, title = title, subtitle = artistNames, thumbnails = thumbnails, track = this)
