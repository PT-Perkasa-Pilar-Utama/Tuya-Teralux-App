package com.example.whisperandroid.service

import android.content.Context
import android.graphics.PixelFormat
import android.view.Gravity
import android.view.View
import android.view.WindowManager
import androidx.compose.ui.platform.ComposeView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.LifecycleRegistry
import androidx.lifecycle.ViewModelStore
import androidx.lifecycle.ViewModelStoreOwner
import androidx.lifecycle.setViewTreeLifecycleOwner
import androidx.lifecycle.setViewTreeViewModelStoreOwner
import androidx.savedstate.SavedStateRegistry
import androidx.savedstate.SavedStateRegistryController
import androidx.savedstate.SavedStateRegistryOwner
import androidx.savedstate.setViewTreeSavedStateRegistryOwner
import com.example.whisperandroid.presentation.assistant.BackgroundAssistantCoordinator
import com.example.whisperandroid.presentation.assistant.BackgroundAssistantModalHost
import com.example.whisperandroid.ui.theme.SensioTheme
import com.example.whisperandroid.util.AppLog

class BackgroundAssistantOverlayController(
    private val context: Context,
    private val coordinator: BackgroundAssistantCoordinator,
    private val onError: () -> Unit
) {
    private val TAG = "Overlay"
    private val windowManager = context.getSystemService(Context.WINDOW_SERVICE) as WindowManager
    private var overlayView: View? = null
    private var lifecycleOwner: OverlayLifecycleOwner? = null

    fun show() {
        if (overlayView != null) return
        AppLog.d(TAG, "show() requested")

        try {
            val layoutParams = WindowManager.LayoutParams().apply {
                type = WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY
                format = PixelFormat.TRANSLUCENT
                flags = WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN or
                    WindowManager.LayoutParams.FLAG_WATCH_OUTSIDE_TOUCH or
                    WindowManager.LayoutParams.FLAG_DRAWS_SYSTEM_BAR_BACKGROUNDS or
                    WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS

                width = WindowManager.LayoutParams.MATCH_PARENT
                height = WindowManager.LayoutParams.MATCH_PARENT
                gravity = Gravity.BOTTOM
            }

            val composeView = ComposeView(context).apply {
                setContent {
                    SensioTheme {
                        BackgroundAssistantModalHost(coordinator = coordinator)
                    }
                }
            }

            // We need a dummy lifecycle owner for Compose in a Service
            val lo = OverlayLifecycleOwner()
            this.lifecycleOwner = lo

            composeView.setViewTreeLifecycleOwner(lo)
            composeView.setViewTreeViewModelStoreOwner(lo)
            composeView.setViewTreeSavedStateRegistryOwner(lo)

            lo.activate()

            windowManager.addView(composeView, layoutParams)
            overlayView = composeView
            AppLog.i(TAG, "Overlay view added successfully")
        } catch (e: Exception) {
            AppLog.e(TAG, "Failed to show overlay window", e)
            cleanupInternal()
            onError()
        }
    }

    private fun cleanupInternal() {
        overlayView?.let { v ->
            try {
                windowManager.removeView(v)
            } catch (e: Exception) {
                AppLog.e(TAG, "Failed to remove view during cleanup", e)
            }
        }
        overlayView = null
        lifecycleOwner?.destroy()
        lifecycleOwner = null
    }

    private inner class OverlayLifecycleOwner : LifecycleOwner, ViewModelStoreOwner, SavedStateRegistryOwner {
        private val lifecycleRegistry = LifecycleRegistry(this)
        private val vmStore = ViewModelStore()
        private val ssrController = SavedStateRegistryController.create(this)

        override val lifecycle: Lifecycle get() = lifecycleRegistry
        override val viewModelStore: ViewModelStore get() = vmStore
        override val savedStateRegistry: SavedStateRegistry get() = ssrController.savedStateRegistry

        fun activate() {
            ssrController.performAttach()
            ssrController.performRestore(null)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_CREATE)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_START)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_RESUME)
        }

        fun destroy() {
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_PAUSE)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_STOP)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_DESTROY)
            vmStore.clear()
        }
    }
    fun hide() {
        android.util.Log.d("SensioOverlay", "hide() requested")
        cleanupInternal()
    }

    fun destroy() {
        hide()
    }
}
