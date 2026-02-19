package com.example.whisper_android.presentation.meeting

import android.app.DownloadManager
import android.net.Uri
import android.os.Environment
import android.widget.Toast

fun downloadPdf(
    context: android.content.Context,
    url: String,
    title: String,
) {
    try {
        // Ensure standard HTTP/HTTPS URL for DownloadManager
        val absoluteUrl =
            if (url.startsWith("/")) {
                val base =
                    com.example.whisper_android.data.di.NetworkModule.BASE_URL
                        .removeSuffix("/")
                "$base$url"
            } else {
                url
            }

        val request =
            DownloadManager
                .Request(Uri.parse(absoluteUrl))
                .setTitle(title)
                .setDescription("Downloading meeting summary PDF...")
                .setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
                .setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, "${title.replace(" ", "_")}.pdf")
                .setAllowedOverMetered(true)
                .setAllowedOverRoaming(true)

        val downloadManager = context.getSystemService(android.content.Context.DOWNLOAD_SERVICE) as DownloadManager
        downloadManager.enqueue(request)
        Toast.makeText(context, "Download started...", Toast.LENGTH_SHORT).show()

        // Auto-open PDF in browser/viewer
        val intent = android.content.Intent(android.content.Intent.ACTION_VIEW, Uri.parse(absoluteUrl))
        context.startActivity(intent)
    } catch (e: Exception) {
        Toast.makeText(context, "Download failed: ${e.message}", Toast.LENGTH_LONG).show()
    }
}
