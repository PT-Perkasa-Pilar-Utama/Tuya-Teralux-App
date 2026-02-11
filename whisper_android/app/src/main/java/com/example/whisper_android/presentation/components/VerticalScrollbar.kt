package com.example.whisper_android.presentation.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.ScrollState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

/**
 * A reusable vertical scrollbar for [LazyListState].
 */
@Composable
fun VerticalScrollbar(
    modifier: Modifier = Modifier,
    lazyListState: LazyListState,
    width: Dp = 4.dp,
    color: Color = MaterialTheme.colorScheme.primary.copy(alpha = 0.5f)
) {
    val showScrollbar by remember {
        derivedStateOf {
            lazyListState.layoutInfo.visibleItemsInfo.size < lazyListState.layoutInfo.totalItemsCount
        }
    }

    if (showScrollbar) {
        val alpha by animateFloatAsState(
            targetValue = if (lazyListState.isScrollInProgress) 1f else 0.4f,
            animationSpec = tween(durationMillis = 500),
            label = "ScrollbarAlpha"
        )

        val visibleItemsInfo = lazyListState.layoutInfo.visibleItemsInfo
        val totalItemsCount = lazyListState.layoutInfo.totalItemsCount
        val viewportHeight = lazyListState.layoutInfo.viewportEndOffset - lazyListState.layoutInfo.viewportStartOffset

        BoxWithConstraints(modifier = modifier.fillMaxHeight().width(width)) {
            val maxHeight = maxHeight
            if (visibleItemsInfo.isEmpty()) return@BoxWithConstraints

            // Estimate total height and current scroll position in pixels
            val averageItemHeight = visibleItemsInfo.map { it.size }.average().toFloat()
            val estimatedTotalHeight = averageItemHeight * totalItemsCount
            val currentScrollPosition = (lazyListState.firstVisibleItemIndex * averageItemHeight) + lazyListState.firstVisibleItemScrollOffset

            // Calculate ratios
            val thumbHeightRatio = (viewportHeight.toFloat() / estimatedTotalHeight).coerceIn(0.1f, 1f)
            val thumbOffsetRatio = (currentScrollPosition / estimatedTotalHeight).coerceIn(0f, 1f - thumbHeightRatio)

            val thumbHeight = maxHeight * thumbHeightRatio
            val thumbOffset = maxHeight * thumbOffsetRatio

            Box(
                modifier = Modifier
                    .offset(y = thumbOffset)
                    .height(thumbHeight)
                    .fillMaxWidth()
                    .alpha(alpha)
                    .clip(CircleShape)
                    .background(color)
            )
        }
    }
}

/**
 * A reusable vertical scrollbar for [ScrollState] (Regular Column).
 */
@Composable
fun VerticalScrollbar(
    modifier: Modifier = Modifier,
    scrollState: ScrollState,
    width: Dp = 4.dp,
    color: Color = MaterialTheme.colorScheme.primary.copy(alpha = 0.5f)
) {
    val showScrollbar by remember {
        derivedStateOf {
            scrollState.maxValue > 0
        }
    }

    if (showScrollbar) {
        val alpha by animateFloatAsState(
            targetValue = if (scrollState.isScrollInProgress) 1f else 0.4f,
            animationSpec = tween(durationMillis = 500),
            label = "ScrollbarAlpha"
        )

        BoxWithConstraints(modifier = modifier.fillMaxHeight().width(width)) {
            val maxHeight = maxHeight
            
            // Calculate thumb height and offset
            val totalContentHeight = scrollState.maxValue + maxHeight.value
            val thumbHeight = (maxHeight.value / totalContentHeight * maxHeight.value).dp.coerceAtLeast(40.dp)
            val thumbOffset = (scrollState.value.toFloat() / scrollState.maxValue * (maxHeight.value - thumbHeight.value)).dp

            Box(
                modifier = Modifier
                    .offset(y = thumbOffset)
                    .height(thumbHeight)
                    .fillMaxWidth()
                    .alpha(alpha)
                    .clip(CircleShape)
                    .background(color)
            )
        }
    }
}
