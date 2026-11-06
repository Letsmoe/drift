package dev.letsmoe.drift.importing.platform

import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import android.webkit.MimeTypeMap
import dev.letsmoe.drift.importing.ImportFileStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File
import java.io.IOException
import java.util.UUID

private const val FALLBACK_NAME = "spotify-import"

/**
 * Copies documents picked through the system file picker into the cache directory. The copy keeps
 * the document's own name, because the core reads its extension and Exportify names playlists
 * after their file.
 */
class ContentResolverImportFileStore(private val context: Context) : ImportFileStore {
    private val rootDirectory = File(context.cacheDir, "import")

    override suspend fun copyIntoWorkspace(documentReference: String): String =
        withContext(Dispatchers.IO) {
            val uri = Uri.parse(documentReference)
            val copyDirectory = File(rootDirectory, UUID.randomUUID().toString())
            copyDirectory.mkdirs()
            val target = File(copyDirectory, fileNameOf(uri))
            val input = context.contentResolver.openInputStream(uri)
                ?: throw IOException("The selected file could not be opened")
            input.use { source -> target.outputStream().use { sink -> source.copyTo(sink) } }
            target.absolutePath
        }

    override suspend fun delete(path: String) {
        withContext(Dispatchers.IO) {
            val copyDirectory = File(path).parentFile ?: return@withContext
            if (copyDirectory.parentFile == rootDirectory) {
                copyDirectory.deleteRecursively()
            }
        }
    }

    override suspend fun deleteAll() {
        withContext(Dispatchers.IO) { rootDirectory.deleteRecursively() }
    }

    private fun fileNameOf(uri: Uri): String {
        val displayName = displayNameOf(uri)
        if (displayName.isNotBlank()) {
            return displayName.replace('/', '_')
        }
        return FALLBACK_NAME + extensionOf(uri)
    }

    private fun displayNameOf(uri: Uri): String {
        val cursor = context.contentResolver.query(uri, arrayOf(OpenableColumns.DISPLAY_NAME), null, null, null)
        cursor.use {
            if (it != null && it.moveToFirst()) {
                return it.getString(0).orEmpty()
            }
        }
        return ""
    }

    private fun extensionOf(uri: Uri): String {
        val mimeType = context.contentResolver.getType(uri) ?: return ""
        val extension = MimeTypeMap.getSingleton().getExtensionFromMimeType(mimeType) ?: return ""
        return ".$extension"
    }
}
