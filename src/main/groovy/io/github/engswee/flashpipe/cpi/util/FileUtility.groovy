package io.github.engswee.flashpipe.cpi.util

import org.slf4j.Logger
import org.slf4j.LoggerFactory

import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.Paths
import java.nio.file.StandardCopyOption

class FileUtility {
    static Logger logger = LoggerFactory.getLogger(FileUtility)

    static void removeCommentsFromFile(File inputFile) {
        logger.debug("Removing comments on ${inputFile}")
        String fileContent = inputFile.getText('UTF-8')
        String updatedContent = fileContent.replaceAll(/#.*\r?\n/, '')
        inputFile.setText(updatedContent, 'UTF-8')
    }

    static void copyDirectory(String sourceDirectoryLocation, String destinationDirectoryLocation) throws IOException {
        logger.debug("Copying directory from ${sourceDirectoryLocation} to ${destinationDirectoryLocation}")
        new File(destinationDirectoryLocation).mkdirs()
        Path sourceDirPath = Paths.get(sourceDirectoryLocation)
        Files.walk(sourceDirPath).forEach({ source ->
            Path destination = Paths.get(destinationDirectoryLocation, source.toString().substring(sourceDirPath.toString().length()))
            Files.copy(source, destination, StandardCopyOption.REPLACE_EXISTING)
        })
    }

    static void replaceDirectory(String sourceDirectoryLocation, String destinationDirectoryLocation) throws IOException {
        new File(destinationDirectoryLocation).deleteDir()
        copyDirectory(sourceDirectoryLocation, destinationDirectoryLocation)
    }

    static void replaceFile(String sourceFilePath, String targetFilePath) {
        new File(targetFilePath).delete()
        Files.copy(Paths.get(sourceFilePath), Paths.get(targetFilePath), StandardCopyOption.REPLACE_EXISTING)
    }
}
