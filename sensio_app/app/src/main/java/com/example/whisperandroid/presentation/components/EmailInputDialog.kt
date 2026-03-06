package com.example.whisperandroid.presentation.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@Composable
fun EmailInputDialog(
    onDismiss: () -> Unit,
    onSend: (isMacMode: Boolean, target: String, subject: String) -> Unit
) {
    var targetInput by remember { mutableStateOf("") }
    var subject by remember { mutableStateOf("Meeting Summary") }
    var targetError by remember { mutableStateOf<String?>(null) }
    var subjectError by remember { mutableStateOf<String?>(null) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Send via Email") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = targetInput,
                    onValueChange = { newValue ->
                        targetInput = newValue
                        if (newValue.isNotBlank()) targetError = null
                    },
                    label = { Text("Recipient Email(s)") },
                    placeholder = {
                        Text("user1@a.com, user2@b.com")
                    },
                    singleLine = true,
                    isError = targetError != null,
                    supportingText = {
                        if (targetError != null) {
                            Text(text = targetError!!)
                        }
                    },
                    modifier = Modifier.fillMaxWidth()
                )

                OutlinedTextField(
                    value = subject,
                    onValueChange = { newValue ->
                        subject = newValue
                        if (newValue.isNotBlank()) subjectError = null
                    },
                    label = { Text("Subject") },
                    singleLine = true,
                    isError = subjectError != null,
                    supportingText = {
                        if (subjectError != null) {
                            Text(text = subjectError!!)
                        }
                    },
                    modifier = Modifier.fillMaxWidth()
                )
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    var isValid = true
                    if (targetInput.isBlank()) {
                        targetError = "Email is required"
                        isValid = false
                    }
                    if (subject.isBlank()) {
                        subjectError = "Subject is required"
                        isValid = false
                    }

                    if (isValid) {
                        onSend(false, targetInput, subject)
                    }
                }
            ) {
                Text("Send")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}
