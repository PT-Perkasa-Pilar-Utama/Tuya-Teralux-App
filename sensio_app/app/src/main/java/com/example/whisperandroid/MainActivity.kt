package com.example.whisperandroid

import android.os.Build
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.domain.usecase.MeetingProcessState
import com.example.whisperandroid.navigation.AppRoutes
import com.example.whisperandroid.presentation.bootstrap.AppBootstrapViewModel
import com.example.whisperandroid.presentation.dashboard.DashboardScreen
import com.example.whisperandroid.presentation.meeting.components.MeetingSuccessContent
import com.example.whisperandroid.presentation.register.RegisterScreen
import com.example.whisperandroid.ui.theme.SensioTheme
import com.example.whisperandroid.util.FeatureAvailabilityGuard
import com.example.whisperandroid.util.normalizeMeetingSummaryMarkdown
import dev.jeziellago.compose.markdowntext.MarkdownText

class MainActivity : ComponentActivity() {
    companion object {
        lateinit var appContext: android.content.Context
            private set
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        appContext = applicationContext

        // Ensure NetworkModule is initialized
        NetworkModule.ensureInitialized(this)

        if (com.example.whisperandroid.util.DeviceUtils.isTerminal()) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
        } else if (com.example.whisperandroid.util.DeviceUtils.isPhone(this)) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        }

        enableEdgeToEdge()
        setContent {
            SensioTheme {
                MainScreen()
            }
        }
    }
}

@Composable
fun MainScreen(
    bootstrapViewModel: AppBootstrapViewModel = viewModel {
        AppBootstrapViewModel(
            NetworkModule.authenticateUseCase,
            NetworkModule.terminalRepository,
            NetworkModule.tokenManager
        )
    }
) {
    val debugHarness = rememberDebugFinalSummaryHarness()
    if (debugHarness != null) {
        DebugFinalSummaryHarnessScreen(harness = debugHarness)
        return
    }

    val context = LocalContext.current
    val navController = rememberNavController()

    val bootstrapState by bootstrapViewModel.uiState.collectAsState()

    LaunchedEffect(Unit) {
        bootstrapViewModel.bootstrap()
    }

    // Permission Handling
    val launcher = androidx.activity.compose.rememberLauncherForActivityResult(
        contract = androidx.activity.result.contract.ActivityResultContracts.RequestMultiplePermissions()
    ) { _ -> }

    LaunchedEffect(Unit) {
        val permissions = mutableListOf(android.Manifest.permission.RECORD_AUDIO)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            permissions.add(android.Manifest.permission.READ_MEDIA_AUDIO)
            permissions.add(android.Manifest.permission.POST_NOTIFICATIONS)
        } else {
            permissions.add(android.Manifest.permission.READ_EXTERNAL_STORAGE)
            permissions.add(android.Manifest.permission.WRITE_EXTERNAL_STORAGE)
        }
        launcher.launch(permissions.toTypedArray())
    }

    val backgroundModeEnabled by NetworkModule.backgroundAssistantModeStore.isEnabled.collectAsState()
    val coordinator = remember { NetworkModule.backgroundAssistantCoordinator }

    // Stop local listeners if background mode is enabled
    // This allows the Service to own the runtime
    LaunchedEffect(backgroundModeEnabled) {
        if (backgroundModeEnabled) {
            val currentRoute = navController.currentBackStackEntry?.destination?.route
            if (currentRoute == AppRoutes.Meeting.route || currentRoute == AppRoutes.Assistant.route) {
                navController.navigate(AppRoutes.Dashboard.route) {
                    popUpTo(AppRoutes.Dashboard.route) { inclusive = false }
                }
            }
        }
    }

    // App Navigation & Overlay
    val token = remember { NetworkModule.tokenManager.getAccessToken() }
    val startDestination = if (token != null) AppRoutes.Dashboard.route else AppRoutes.Register.route

    Box(modifier = Modifier.fillMaxSize()) {
        NavHost(
            navController = navController,
            startDestination = startDestination
        ) {
            composable(AppRoutes.Register.route) {
                RegisterScreen(onNavigateToDashboard = {
                    navController.navigate(AppRoutes.Dashboard.route) {
                        popUpTo(AppRoutes.Register.route) { inclusive = true }
                    }
                })
            }

            composable(AppRoutes.Dashboard.route) {
                DashboardScreen(
                    onNavigateToRegister = {
                        navController.navigate(AppRoutes.Register.route) {
                            popUpTo(AppRoutes.Dashboard.route) { inclusive = true }
                        }
                    },
                    onNavigateToUpload = { },
                    onNavigateToStreaming = {
                        if (!bootstrapState.isBootstrapped) {
                            Toast.makeText(context, "Syncing devices, please wait...", Toast.LENGTH_SHORT).show()
                        } else if (FeatureAvailabilityGuard.canOpenInteractiveScreens(backgroundModeEnabled)) {
                            navController.navigate(AppRoutes.Meeting.route)
                        }
                    },
                    onNavigateToEdge = {
                        if (!bootstrapState.isBootstrapped) {
                            Toast.makeText(context, "Syncing devices, please wait...", Toast.LENGTH_SHORT).show()
                        } else if (FeatureAvailabilityGuard.canOpenInteractiveScreens(backgroundModeEnabled)) {
                            navController.navigate(AppRoutes.Assistant.route)
                        }
                    },
                    bootstrapViewModel = bootstrapViewModel
                )
            }

            composable(AppRoutes.Meeting.route) {
                LaunchedEffect(backgroundModeEnabled) {
                    if (backgroundModeEnabled) {
                        Toast.makeText(context, "Feature disabled in Background Mode", Toast.LENGTH_SHORT).show()
                        navController.navigate(AppRoutes.Dashboard.route) {
                            popUpTo(AppRoutes.Meeting.route) { inclusive = true }
                        }
                    }
                }

                if (!backgroundModeEnabled) {
                    com.example.whisperandroid.presentation.meeting.MeetingTranscriberScreen(
                        onNavigateBack = { navController.popBackStack() }
                    )
                }
            }

            composable(AppRoutes.Assistant.route) {
                LaunchedEffect(backgroundModeEnabled) {
                    if (backgroundModeEnabled) {
                        Toast.makeText(context, "Feature disabled in Background Mode", Toast.LENGTH_SHORT).show()
                        navController.navigate(AppRoutes.Dashboard.route) {
                            popUpTo(AppRoutes.Assistant.route) { inclusive = true }
                        }
                    }
                }

                if (!backgroundModeEnabled) {
                    com.example.whisperandroid.presentation.assistant.AiAssistantScreen(
                        onNavigateBack = { navController.popBackStack() }
                    )
                }
            }
        }

        // Foreground-Only Background Assistant Modal Overlay
        com.example.whisperandroid.presentation.assistant.BackgroundAssistantModalHost(
            coordinator = coordinator
        )
    }
}

@Composable
private fun DebugFinalSummaryHarnessScreen(harness: DebugFinalSummaryHarness) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp)
    ) {
        Text(
            text = "Final Summary QA Harness",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold
        )
        Text(
            text = harness.label,
            style = MaterialTheme.typography.bodyMedium,
            modifier = Modifier.padding(top = 4.dp, bottom = 16.dp)
        )

        DebugHarnessSection(
            title = "MeetingSuccessContent"
        ) {
            MeetingSuccessContentBody(
                summary = harness.summary,
                modifier = Modifier.fillMaxWidth()
            )
        }

        HorizontalDivider(modifier = Modifier.padding(vertical = 24.dp))

        DebugHarnessSection(
            title = "Summary Preview parity target"
        ) {
            MarkdownText(
                markdown = normalizeMeetingSummaryMarkdown(harness.summary),
                style = MaterialTheme.typography.bodyLarge,
                modifier = Modifier.fillMaxWidth()
            )
        }
    }
}

@Composable
private fun DebugHarnessSection(
    title: String,
    content: @Composable () -> Unit
) {
    Text(
        text = title,
        style = MaterialTheme.typography.titleMedium,
        fontWeight = FontWeight.SemiBold,
        modifier = Modifier.padding(bottom = 12.dp)
    )
    Box(modifier = Modifier.fillMaxWidth()) {
        content()
    }
}

@Composable
private fun MeetingSuccessContentBody(
    summary: String,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier) {
        Text(
            text = "Meeting Summary",
            style = MaterialTheme.typography.titleLarge,
            color = MaterialTheme.colorScheme.primary,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 2.dp)
        )
        MarkdownText(
            markdown = normalizeMeetingSummaryMarkdown(summary),
            style = MaterialTheme.typography.bodyLarge,
            modifier = Modifier.fillMaxWidth()
        )
    }
}

@Composable
private fun rememberDebugFinalSummaryHarness(): DebugFinalSummaryHarness? {
    if (!BuildConfig.DEBUG) return null

    val context = LocalContext.current
    val activity = context as? ComponentActivity ?: return null
    val mode = activity.intent?.getStringExtra(DEBUG_FINAL_SUMMARY_MODE)?.trim().orEmpty()
    if (mode.isEmpty()) return null

    return when (mode.lowercase()) {
        DEBUG_FINAL_SUMMARY_MODE_KNOWN_BROKEN -> DebugFinalSummaryHarness(
            mode = DEBUG_FINAL_SUMMARY_MODE_KNOWN_BROKEN,
            label = "Known-broken real summary sample",
            summary = DEBUG_KNOWN_BROKEN_REAL_SUMMARY
        )

        DEBUG_FINAL_SUMMARY_MODE_MALFORMED -> DebugFinalSummaryHarness(
            mode = DEBUG_FINAL_SUMMARY_MODE_MALFORMED,
            label = "Malformed-but-common markdown sample",
            summary = DEBUG_MALFORMED_BUT_COMMON_SUMMARY
        )

        else -> null
    }
}

private data class DebugFinalSummaryHarness(
    val mode: String,
    val label: String,
    val summary: String
)

private const val DEBUG_FINAL_SUMMARY_MODE = "debug_final_summary_mode"
private const val DEBUG_FINAL_SUMMARY_MODE_KNOWN_BROKEN = "known-broken"
private const val DEBUG_FINAL_SUMMARY_MODE_MALFORMED = "malformed"

private val DEBUG_MALFORMED_BUT_COMMON_SUMMARY = """
    # Meeting Summary

    ## Attendees
    - John Doe
    - Jane Smith

    ## Notes

    This is a summary with   extra spaces.

    Another line with trailing spaces.

    | Col A | Col B |
    | --- | --- |
    | Val 1 | Val 2 |

    1. First item
    2. Second item


    Some content after extra blank lines.

    Final paragraph here.
""".trimIndent()

private val DEBUG_KNOWN_BROKEN_REAL_SUMMARY = """
    # Ringkasan Pertemuan Tim

    ## Poin Penting
    - Pertumbuhan pasar mencapai 15% di kuartal ini.
    - Alokasi anggaran baru sudah disetujui.
    - Perlu fokus pada kemitraan strategis bulan depan — terutama dengan partner utama.

    ### Action Items
    1. Kirim dokumen anggaran ke tim finance — deadline Jumat
    2. Jadwalkan meeting dengan partner eksternal — tanggal belum ditentukan
    3. Review proposal dari vendor — perlu persetujuan manajemen

    ## Catatan Tambahan

    Diskusi berjalan dengan lancar. Semua pihak setuju dengan proposal yang diajukan.
    Beberapa poin penting yang perlu diperhatikan — masalah teknis yang masih pending
    dan sumber daya yang terbatas.

    | Item | Status | Priority |
    | --- | --- | --- |
    | Budget allocation | Approved | High |
    | Partnership proposal | Under review | Medium |
    | Technical setup | In progress | High |
""".trimIndent()
