package dev.letsmoe.drift.core.model

import kotlinx.serialization.Serializable

@Serializable
data class Reason(
    val kind: String = "",
    val entityId: String = "",
    val entityName: String = "",
    val text: String = "",
)

@Serializable
data class RecommendedItem(
    val track: Track,
    val source: String = "",
    val reason: Reason = Reason(),
    val score: Double = 0.0,
)

@Serializable
data class RecommendedRow(
    val key: String = "",
    val title: String = "",
    val reason: Reason = Reason(),
    val items: List<RecommendedItem> = emptyList(),
)
