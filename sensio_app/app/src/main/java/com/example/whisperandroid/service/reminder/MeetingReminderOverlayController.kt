package com.example.whisperandroid.service.reminder

import android.content.Context
import android.graphics.PixelFormat
import android.os.Build
import android.util.Log
import android.view.Gravity
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
import com.example.whisperandroid.domain.model.reminder.MeetingReminderUiModel
import com.example.whisperandroid.presentation.reminder.MeetingReminderOverlayHost

/**
 * Controller for managing the meeting reminder overlay window.
 *
 * Handles showing and hiding the overlay using WindowManager.
 */
class MeetingReminderOverlayController(
    private val context: Context,
    private val arbiter: OverlayArbiter
) {
    private val windowManager: WindowManager = context.getSystemService(Context.WINDOW_SERVICE) as WindowManager
    private var overlayView: ComposeView? = null
    private var lifecycleOwner: OverlayLifecycleOwner? = null
    private var isShowing = false
    private val tag = "ReminderOverlayCtrl"

    /**
     * Show the reminder overlay.
     *
     * @param uiModel The UI model to display
     */
    fun show(uiModel: MeetingReminderUiModel) {
        if (isShowing) {
            Log.w(tag, "Overlay already showing, skipping")
            return
        }

        if (!hasOverlayPermission()) {
            Log.w(tag, "Missing overlay permission, cannot show overlay")
            return
        }

        try {
            val composeView = ComposeView(context).apply {
                setContent {
                    MeetingReminderOverlayHost(
                        uiModel = uiModel,
                        onClose = { hide() }
                    )
                }
            }

            // We need a dummy lifecycle owner for Compose in a Service
            val lo = OverlayLifecycleOwner()
            this.lifecycleOwner = lo

            composeView.setViewTreeLifecycleOwner(lo)
            composeView.setViewTreeViewModelStoreOwner(lo)
            composeView.setViewTreeSavedStateRegistryOwner(lo)

            lo.activate()

            val layoutParams = WindowManager.LayoutParams().apply {
                type = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                    WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY
                } else {
                    @Suppress("DEPRECATION")
                    WindowManager.LayoutParams.TYPE_PHONE
                }
                flags = WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN or
                    WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS or
                    WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL or
                    WindowManager.LayoutParams.FLAG_WATCH_OUTSIDE_TOUCH
                format = PixelFormat.TRANSLUCENT
                width = WindowManager.LayoutParams.WRAP_CONTENT
                height = WindowManager.LayoutParams.WRAP_CONTENT
                gravity = Gravity.CENTER
            }

            windowManager.addView(composeView, layoutParams)
            overlayView = composeView
            isShowing = true
            arbiter.markReminderOverlayActive(true)
            Log.i(tag, "Overlay shown")
        } catch (e: Exception) {
            Log.e(tag, "Failed to show overlay: ${e.message}")
            cleanupInternal()
        }
    }

    private fun cleanupInternal() {
        overlayView?.let { v ->
            try {
                windowManager.removeView(v)
            } catch (e: Exception) {
                Log.e(tag, "Failed to remove view during cleanup: ${e.message}")
            }
        }
        overlayView = null
        lifecycleOwner?.destroy()
        lifecycleOwner = null
        isShowing = false
        arbiter.markReminderOverlayActive(false)
    }

    /**
     * Hide the reminder overlay.
     */
    fun hide() {
        if (!isShowing) {
            return
        }
        cleanupInternal()
        Log.i(tag, "Overlay hidden")
    }

    /**
     * Check if overlay permission is granted.
     */
    private fun hasOverlayPermission(): Boolean {
        return android.provider.Settings.canDrawOverlays(context)
    }

    /**
     * Check if overlay is currently showing.
     */
    fun isShowing(): Boolean = isShowing

    /**
     * Clean up resources.
     */
    fun destroy() {
        hide()
    }

    /**
     * Lifecycle owner for overlay ComposeView in a Service context.
     */
    private class OverlayLifecycleOwner : LifecycleOwner, ViewModelStoreOwner, SavedStateRegistryOwner {
        private val lifecycleRegistry = LifecycleRegistry(this)
        private val viewModelStoreInstance = ViewModelStore()
        private val savedStateRegistryController = SavedStateRegistryController.create(this)

        override val lifecycle: Lifecycle get() = lifecycleRegistry
        override val viewModelStore: ViewModelStore get() = viewModelStoreInstance
        override val savedStateRegistry: SavedStateRegistry get() = savedStateRegistryController.savedStateRegistry

        fun activate() {
            if (lifecycleRegistry.currentState == Lifecycle.State.INITIALIZED) {
                savedStateRegistryController.performAttach()
                savedStateRegistryController.performRestore(null)
            }
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_CREATE)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_START)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_RESUME)
        }

        fun destroy() {
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_PAUSE)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_STOP)
            lifecycleRegistry.handleLifecycleEvent(Lifecycle.Event.ON_DESTROY)
            viewModelStoreInstance.clear()
        }
    }
}
