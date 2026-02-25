package com.example.whisper_android.presentation.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

@Composable
fun EmailInputDialog(
    onDismiss: () -> Unit,
    onSend: (isMacMode: Boolean, target: String, subject: String) -> Unit
) {
    var isMacMode by remember { mutableStateOf(false) }
    var targetInput by remember { mutableStateOf("") }
    var subject by remember { mutableStateOf("Meeting Summary") }
    var targetError by remember { mutableStateOf<String?>(null) }
    var subjectError by remember { mutableStateOf<String?>(null) }
    
    val context = androidx.compose.ui.platform.LocalContext.current
    val deviceMacAddress = remember { com.example.whisper_android.util.DeviceUtils.getDeviceId(context) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Send via Email") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                
                // Mode Selector
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(48.dp)
                        .clip(RoundedCornerShape(8.dp))
                        .background(MaterialTheme.colorScheme.surfaceVariant),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxHeight()
                            .padding(4.dp)
                            .clip(RoundedCornerShape(6.dp))
                            .background(if (!isMacMode) MaterialTheme.colorScheme.primary else Color.Transparent)
                            .clickable {
                                isMacMode = false
                                targetInput = ""
                                targetError = null
                            },
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = "Custom Email",
                            color = if (!isMacMode) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant,
                            fontWeight = if (!isMacMode) FontWeight.Bold else FontWeight.Normal,
                            style = MaterialTheme.typography.labelLarge
                        )
                    }

                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxHeight()
                            .padding(4.dp)
                            .clip(RoundedCornerShape(6.dp))
                            .background(if (isMacMode) MaterialTheme.colorScheme.primary else Color.Transparent)
                            .clickable {
                                isMacMode = true
                                targetInput = deviceMacAddress
                                targetError = null
                            },
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = "Room MAC",
                            color = if (isMacMode) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant,
                            fontWeight = if (isMacMode) FontWeight.Bold else FontWeight.Normal,
                            style = MaterialTheme.typography.labelLarge
                        )
                    }
                }

                OutlinedTextField(
                    value = targetInput,
                    onValueChange = {
                        if (!isMacMode) {
                            targetInput = it
                            if (it.isNotBlank()) targetError = null
                        }
                    },
                    readOnly = isMacMode,
                    label = { Text(if (isMacMode) "MAC Address" else "Recipient Email(s)") },
                    placeholder = { Text(if (isMacMode) "e.g., AA:BB:CC:DD:EE:FF" else "user1@a.com, user2@b.com") },
                    singleLine = true,
                    isError = targetError != null,
                    supportingText = { if (targetError != null) Text(targetError!!) },
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
                    if (targetInput.isBlank()) {
                        targetError = if (isMacMode) "MAC Address is required" else "Email is required"
                        isValid = false
                    }
                    if (subject.isBlank()) {
                        subjectError = "Subject is required"
                        isValid = false
                    }

                    if (isValid) {
                        onSend(isMacMode, targetInput, subject)
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
