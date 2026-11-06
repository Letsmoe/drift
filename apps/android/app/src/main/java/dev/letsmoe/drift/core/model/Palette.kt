package dev.letsmoe.drift.core.model

import kotlinx.serialization.Serializable

/** The colors of one cover image as "#rrggbb"; all are empty when the image has no visible pixels. */
@Serializable
data class CoverColors(
    val dominant: String = "",
    val vibrant: String = "",
    val accent: String = "",
    /** Up to four clearly different colors, most common first. */
    val colors: List<String> = emptyList(),
)
