package com.example.whisperandroid.utils

import android.content.Context
import java.io.File
import java.util.UUID
import net.lingala.zip4j.ZipFile
import net.lingala.zip4j.model.ZipParameters
import net.lingala.zip4j.model.enums.AesKeyStrength
import net.lingala.zip4j.model.enums.EncryptionMethod

class ZipEncryptor {

    suspend fun createEncryptedZip(
        context: Context,
        sourceFilePath: String,
        password: String
    ): File {
        val uuid = UUID.randomUUID().toString()
        val outputFile = File(context.cacheDir, "$uuid.zip")

        val zipFile = ZipFile(outputFile, password.toCharArray())
        val zipParameters = ZipParameters().apply {
            setEncryptFiles(true)
            encryptionMethod = EncryptionMethod.AES
            aesKeyStrength = AesKeyStrength.KEY_STRENGTH_256
        }
        zipFile.addFile(File(sourceFilePath), zipParameters)

        return outputFile
    }
}
