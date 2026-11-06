package dev.letsmoe.drift.importing

/** Turns a file the user picked into a plain path the core can read. Implemented per platform. */
interface ImportFileStore {
    /** Copies the picked document into app-private storage under its own name and returns the path. */
    suspend fun copyIntoWorkspace(documentReference: String): String

    /** Deletes a copy made by [copyIntoWorkspace]. */
    suspend fun delete(path: String)

    /** Deletes every copy, e.g. those a killed process left behind. */
    suspend fun deleteAll()
}

/** Keeps the process alive, and tells the user, while an import runs. Implemented per platform. */
interface ImportKeepAlive {
    fun start()

    fun stop()
}
