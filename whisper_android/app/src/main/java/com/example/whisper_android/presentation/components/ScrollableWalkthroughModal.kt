package com.example.whisper_android.presentation.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

/**
 * A reusable scrollable walkthrough modal with an integrated vertical scrollbar.
 */
@Composable
fun ScrollableWalkthroughModal(
    title: String,
    showDialog: Boolean,
    onDismiss: () -> Unit,
    content: @Composable ColumnScope.() -> Unit,
) {
    if (showDialog) {
        AlertDialog(
            onDismissRequest = onDismiss,
            title = {
                Text(
                    text = "$title Walkthrough",
                    style = MaterialTheme.typography.headlineSmall.copy(fontWeight = FontWeight.Bold),
                )
            },
            text = {
                val scrollState = rememberScrollState()
                Box(modifier = Modifier.heightIn(max = 450.dp)) {
                    Column(
                        modifier =
                            Modifier
                                .fillMaxWidth()
                                .verticalScroll(scrollState)
                                .padding(end = 12.dp),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                        content = content,
                    )
                    VerticalScrollbar(
                        modifier =
                            Modifier
                                .align(Alignment.CenterEnd)
                                .padding(vertical = 4.dp),
                        scrollState = scrollState,
                    )
                }
            },
            confirmButton = {
                TextButton(onClick = onDismiss) {
                    Text("Got it", fontWeight = FontWeight.Bold)
                }
            },
            shape = RoundedCornerShape(28.dp),
        )
    }
}
