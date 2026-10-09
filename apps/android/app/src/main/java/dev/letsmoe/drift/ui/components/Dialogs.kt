package dev.letsmoe.drift.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

/** Edits a local playlist's name and description; [onSave] receives trimmed values and a non-empty name. */
@Composable
fun EditPlaylistDialog(
    name: String,
    description: String,
    onSave: (name: String, description: String) -> Unit,
    onDismiss: () -> Unit,
) {
    var draftName by remember { mutableStateOf(name) }
    var draftDescription by remember { mutableStateOf(description) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Edit details") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                OutlinedTextField(
                    value = draftName,
                    onValueChange = { draftName = it },
                    singleLine = true,
                    label = { Text("Name") },
                )
                OutlinedTextField(
                    value = draftDescription,
                    onValueChange = { draftDescription = it },
                    minLines = 2,
                    maxLines = 4,
                    label = { Text("Description") },
                )
            }
        },
        confirmButton = {
            TextButton(
                onClick = { onSave(draftName.trim(), draftDescription.trim()) },
                enabled = draftName.isNotBlank(),
            ) { Text("Save") }
        },
        dismissButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}

/** A yes / no question; the confirm button is red for destructive actions. */
@Composable
fun ConfirmDialog(
    title: String,
    message: String,
    confirmLabel: String,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
    confirmColor: Color = MaterialTheme.colorScheme.error,
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(title) },
        text = { Text(message) },
        confirmButton = {
            TextButton(onClick = onConfirm) { Text(confirmLabel, color = confirmColor) }
        },
        dismissButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}
