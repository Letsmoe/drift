package dev.letsmoe.drift.core

import kotlinx.serialization.json.Json

/** The one Json configuration for everything exchanged with the Go core. */
val DriftJson = Json {
    ignoreUnknownKeys = true
    coerceInputValues = true
    explicitNulls = false
    encodeDefaults = true
}
