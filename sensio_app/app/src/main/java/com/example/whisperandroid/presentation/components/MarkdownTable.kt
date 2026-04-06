package com.example.whisperandroid.presentation.components

import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisperandroid.util.MarkdownBlock

/**
 * Renders a markdown table in a TV-friendly format.
 * 
 * Features:
 * - Responsive layout with clear row separation
 * - Header row with stronger emphasis
 * - Handles empty cells safely
 * - Horizontal scroll fallback for very wide tables
 * - Adequate padding and spacing for TV viewing distance
 * 
 * @param table The table block to render
 * @param modifier Optional modifier for the table container
 */
@Composable
fun MarkdownTable(
    table: MarkdownBlock.Table,
    modifier: Modifier = Modifier
) {
    val rows = table.rows
    if (rows.isEmpty()) return
    
    val headerRow = rows.first()
    val dataRows = rows.drop(1)
    
    // Determine max columns for consistent layout
    val maxColumns = rows.maxOfOrNull { it.size } ?: 0
    if (maxColumns == 0) return
    
    Card(
        modifier = modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp)
    ) {
        Column(
            modifier = Modifier.padding(8.dp)
        ) {
            // Render header row
            TableHeaderRow(
                cells = headerRow,
                maxColumns = maxColumns
            )
            
            // Render divider
            Spacer(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(1.dp)
                    .background(MaterialTheme.colorScheme.onSurface.copy(alpha = 0.1f))
            )
            
            // Render data rows
            dataRows.forEach { row ->
                TableRow(
                    cells = row,
                    maxColumns = maxColumns,
                    isLastRow = row == dataRows.last()
                )
            }
        }
    }
}

/**
 * Renders a table header row with emphasis.
 */
@Composable
private fun TableHeaderRow(
    cells: List<String>,
    maxColumns: Int
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .horizontalScroll(rememberScrollState())
            .height(IntrinsicSize.Min),
        verticalAlignment = Alignment.CenterVertically
    ) {
        for (i in 0 until maxColumns) {
            val cellContent = if (i < cells.size) cells[i] else ""
            TableCell(
                text = cellContent,
                isHeader = true,
                modifier = Modifier.weight(1f)
            )
            
            // Add divider between cells (except after last)
            if (i < maxColumns - 1) {
                Spacer(
                    modifier = Modifier
                        .width(1.dp)
                        .height(32.dp)
                        .background(MaterialTheme.colorScheme.onSurface.copy(alpha = 0.1f))
                )
            }
        }
    }
}

/**
 * Renders a standard table row.
 */
@Composable
private fun TableRow(
    cells: List<String>,
    maxColumns: Int,
    isLastRow: Boolean
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .horizontalScroll(rememberScrollState())
            .height(IntrinsicSize.Min),
        verticalAlignment = Alignment.CenterVertically
    ) {
        for (i in 0 until maxColumns) {
            val cellContent = if (i < cells.size) cells[i] else ""
            TableCell(
                text = cellContent,
                isHeader = false,
                modifier = Modifier.weight(1f)
            )
            
            // Add divider between cells (except after last)
            if (i < maxColumns - 1) {
                Spacer(
                    modifier = Modifier
                        .width(1.dp)
                        .height(32.dp)
                        .background(MaterialTheme.colorScheme.onSurface.copy(alpha = 0.05f))
                )
            }
        }
    }
    
    // Add bottom divider for row separation (except for last row)
    if (!isLastRow) {
        Spacer(
            modifier = Modifier
                .fillMaxWidth()
                .height(1.dp)
                .background(MaterialTheme.colorScheme.onSurface.copy(alpha = 0.05f))
        )
    }
}

/**
 * Renders a single table cell with appropriate styling.
 */
@Composable
private fun TableCell(
    text: String,
    isHeader: Boolean,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier.padding(12.dp),
        contentAlignment = Alignment.CenterStart
    ) {
        Text(
            text = text.ifEmpty { "—" }, // Use em-dash for empty cells
            style = if (isHeader) {
                MaterialTheme.typography.labelLarge.copy(
                    fontWeight = FontWeight.Bold,
                    fontSize = 14.sp,
                    lineHeight = 20.sp
                )
            } else {
                MaterialTheme.typography.bodyMedium.copy(
                    fontWeight = FontWeight.Normal,
                    fontSize = 13.sp,
                    lineHeight = 18.sp
                )
            },
            color = if (isHeader) {
                MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.9f)
            } else {
                MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
            },
            maxLines = if (isHeader) 2 else 3, // Allow wrapping for long content
            softWrap = true
        )
    }
}

/**
 * Renders a fallback layout for malformed or complex tables.
 * This stacked layout is more readable than showing raw markdown pipes.
 */
@Composable
fun MalformedTableFallback(
    rawMarkdown: String,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.2f)
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            rawMarkdown
                .lines()
                .filter { it.trim().isNotEmpty() }
                .forEach { line ->
                    val trimmed = line.trim()
                    // Remove leading/trailing pipes for cleaner display
                    val cleaned = trimmed
                        .removePrefix("|")
                        .removeSuffix("|")
                        .trim()
                        .replace(Regex("\\|"), " → ")
                    
                    Text(
                        text = cleaned,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f),
                        modifier = Modifier.padding(vertical = 4.dp)
                    )
                }
        }
    }
}
