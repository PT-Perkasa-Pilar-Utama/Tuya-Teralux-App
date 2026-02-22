package com.example.whisper_android.presentation.components

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
    onSend: (String, String) -> Unit
) {
    var email by remember { mutableStateOf("") }
    var subject by remember { mutableStateOf("Meeting Summary") }
    var emailError by remember { mutableStateOf<String?>(null) }
    var subjectError by remember { mutableStateOf<String?>(null) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Send via Email") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = email,
                    onValueChange = {
                        email = it
                        if (it.isNotBlank()) emailError = null
                    },
                    label = { Text("Recipient Email") },
                    singleLine = true,
                    isError = emailError != null,
                    supportingText = { if (emailError != null) Text(emailError!!) },
                    modifier = Modifier.fillMaxWidth()
                )

                OutlinedTextField(
                    value = subject,
                    onValueChange = {
                        subject = it
                        if (it.isNotBlank()) subjectError = null
                    },
                    label = { Text("Subject") },
                    singleLine = true,
                    isError = subjectError != null,
                    supportingText = { if (subjectError != null) Text(subjectError!!) },
                    modifier = Modifier.fillMaxWidth()
                )
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    var isValid = true
                    if (email.isBlank()) {
                        emailError = "Email is required"
                        isValid = false
                    }
                    if (subject.isBlank()) {
                        subjectError = "Subject is required"
                        isValid = false
                    }

                    if (isValid) {
                        onSend(email, subject)
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
