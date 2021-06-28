package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.Paths
import java.nio.file.StandardCopyOption

class DownloadIntegrationPackageContent extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(DownloadIntegrationPackageContent)

    static void main(String[] args) {
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.execute()
    }

    @Override
    void execute() {
        def packageId = getMandatoryEnvVar('PACKAGE_ID')
        def workDir = getMandatoryEnvVar('WORK_DIR')
        def gitSrcDir = getMandatoryEnvVar('GIT_SRC_DIR')
        String dirNamingType = (System.getenv('DIR_NAMING_TYPE') ?: 'ID')
        if (!['ID', 'NAME'].contains(dirNamingType.toUpperCase())) {
            println "[ERROR] 🛑 Value ${dirNamingType} for environment variable DIR_NAMING_TYPE not in list of accepted values: ID or NAME"
            System.exit(1)
        }
        String draftHandling = (System.getenv('DRAFT_HANDLING') ?: 'SKIP')
        if (!['SKIP', 'ADD', 'ERROR'].contains(draftHandling.toUpperCase())) {
            println "[ERROR] 🛑 Value ${draftHandling} for environment variable DRAFT_HANDLING not in list of accepted values: SKIP, ADD or ERROR"
            System.exit(1)
        }
        List includedIds = System.getenv('INCLUDE_IDS')?.split(',')?.toList()*.trim()
        List excludeIds = System.getenv('EXCLUDE_IDS')?.split(',')?.toList()*.trim()
        if (includedIds && excludeIds) {
            println '[ERROR] 🛑 INCLUDE_IDS and EXCLUDE_IDS are mutually exclusive - use only one of them'
            System.exit(1)
        }

        // Get all design time artifacts of package
        println "[INFO] Getting artifacts in integration package ${packageId}"
        IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
        List artifacts = integrationPackage.getIFlowsWithDraftState(packageId)
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)

        // Create temp directories in working dir
        new File("${workDir}/download").mkdir()
        new File("${workDir}/from_git").mkdir()
        new File("${workDir}/from_tenant").mkdir()

        List filteredArtifacts = filterArtifacts(artifacts, includedIds, excludeIds)

        // Process through the artifacts
        for (Map artifact : filteredArtifacts) {
            println '---------------------------------------------------------------------------------'
            println "[INFO] Begin processing for artifact ${artifact.id}"
            // Check if artifact is in draft version
            if (artifact.isDraft) {
                switch (draftHandling.toUpperCase()) {
                    case 'SKIP':
                        println "[WARNING] ⚠️ Integration artifact ${artifact.id} is in draft version, and will be skipped"
                        continue
                    case 'ADD':
                        println "[INFO] Integration artifact ${artifact.id} is in draft version, and will be added"
                        break
                    case 'ERROR':
                        println "[ERROR] 🛑 Integration artifact ${artifact.id} is in draft version, Save Version in Web UI first!"
                        System.exit(1)
                }
            }
            // Download IFlow
            println "[INFO] Downloading IFlow ${artifact.id} from tenant for comparison"
            File outputZip = new File("${workDir}/download/${artifact.id}.zip")
            outputZip.bytes = designTimeArtifact.download(artifact.id, 'active')
            println "[INFO] IFlow ${artifact.id} downloaded to ${outputZip}"

            // Unzip IFlow contents
            def directoryName = (dirNamingType.toUpperCase() == 'NAME') ? artifact.name : artifact.id
            ZipUtil.unpack(outputZip, new File("${workDir}/download/${directoryName}"))
            println "[INFO] Downloaded IFlow artifact unzipped to ${workDir}/download/${directoryName}"

            // (1) If IFlow already exists in Git, then compare and update
            if (new File("${gitSrcDir}/${directoryName}").exists()) {
                println "[INFO] Comparing content from tenant against Git"
                // Copy to temp directory for diff comparison
                copyDirectory("${workDir}/download/${directoryName}/src/main/resources", "${workDir}/from_tenant/${directoryName}")
                copyDirectory("${gitSrcDir}/${directoryName}/src/main/resources", "${workDir}/from_git/${directoryName}")

                // Remove comments from parameters.prop before comparison only if it exists
                File tenantParamFile = new File("${workDir}/from_tenant/${directoryName}/parameters.prop")
                File gitParamFile = new File("${workDir}/from_git/${directoryName}/parameters.prop")
                if (tenantParamFile.exists() && gitParamFile.exists()) {
                    removeCommentsFromFile(tenantParamFile)
                    removeCommentsFromFile(gitParamFile)
                }
                // Execute shell command diff to compare directory contents
                String command = "diff --strip-trailing-cr -qr -x '.DS_Store' ${workDir}/from_tenant/${directoryName} ${workDir}/from_git/${directoryName}"
                println "[INFO] Executing shell command: ${command}"
                ProcessBuilder processBuilder = new ProcessBuilder()
                processBuilder.command('bash', '-c', command)
                Process process = processBuilder.start()
                process.waitFor()
                switch (process.exitValue()) {
                    case 0:
                        println '[INFO] 🏆 No changes detected. Update to Git not required'
                        break
                    case 1:
                        println process.getText()
                        println '[INFO] 🏆 Changes detected and will be updated to Git'
                        // Update the changes into the Git directory
                        // (a) Replace /src/main/resources
                        new File("${gitSrcDir}/${directoryName}/src/main/resources").deleteDir()
                        copyDirectory("${workDir}/download/${directoryName}/src/main/resources", "${gitSrcDir}/${directoryName}/src/main/resources")
                        // (b) Replace /META-INF/MANIFEST.MF
                        new File("${gitSrcDir}/${directoryName}/META-INF/MANIFEST.MF").delete()
                        Files.copy(Paths.get("${workDir}/download/${directoryName}/META-INF/MANIFEST.MF"), Paths.get("${gitSrcDir}/${directoryName}/META-INF/MANIFEST.MF"), StandardCopyOption.REPLACE_EXISTING)
                        break
                    default:
                        println "[ERROR] 🛑 ${process.err.text}"
                        System.exit(1)
                }
            } else {
                // (2) If IFlow does not exist in Git, then add it
                if (!new File(gitSrcDir).exists()) {
                    new File(gitSrcDir).mkdir()
                }
                println "[INFO] 🏆 Artifact ${artifact.id} does not exist, and will be added to Git"
                copyDirectory("${workDir}/download/${directoryName}", "${gitSrcDir}/${directoryName}")
            }

        }
        // Clean up working directory
        new File("${workDir}/download").deleteDir()
        new File("${workDir}/from_git").deleteDir()
        new File("${workDir}/from_tenant").deleteDir()
        println '---------------------------------------------------------------------------------'
        println "[INFO] 🏆 Completed processing of integration package ${packageId}"
    }

    List filterArtifacts(List artifacts, List includeIds, List excludeIds) {
        if (includeIds) {
            List outputList = []
            includeIds.each { iFlowId ->
                Map artifactDetails = artifacts.find { it.id == iFlowId }
                if (!artifactDetails) {
                    println "[ERROR] 🛑 IFlow ${iFlowId} in INCLUDE_IDS does not exist"
                    System.exit(1)
                } else {
                    outputList.add(artifactDetails)
                }
            }
            println "[INFO] Include only IFlow with IDs - ${System.getenv('INCLUDE_IDS')}"
            return outputList
        } else if (excludeIds) {
            List outputList = []
            // Check if the Ids are valid
            excludeIds.each { iFlowId ->
                Map artifactDetails = artifacts.find { it.id == iFlowId }
                if (!artifactDetails) {
                    println "[ERROR] 🛑 IFlow ${iFlowId} in EXCLUDE_IDS does not exist"
                    System.exit(1)
                }
            }
            println "[INFO] Exclude IFlow with IDs - ${System.getenv('EXCLUDE_IDS')}"
            artifacts.each { artifact ->
                if (!excludeIds.contains(artifact.id)) {
                    outputList.add(artifact)
                }
            }
            return outputList
        } else {
            return artifacts
        }
    }

    void removeCommentsFromFile(File inputFile) {
        logger.debug("[INFO] Removing comments on ${inputFile}")
        String fileContent = inputFile.getText('UTF-8')
        String updatedContent = fileContent.replaceAll(/#.*\r?\n/, '')
        inputFile.setText(updatedContent, 'UTF-8')
    }

    void copyDirectory(String sourceDirectoryLocation, String destinationDirectoryLocation) throws IOException {
        logger.debug("[INFO] Copying directory from ${sourceDirectoryLocation} to ${destinationDirectoryLocation}")
        Files.walk(Paths.get(sourceDirectoryLocation)).forEach({ source ->
            Path destination = Paths.get(destinationDirectoryLocation, source.toString().substring(sourceDirectoryLocation.length()))
            Files.copy(source, destination, StandardCopyOption.REPLACE_EXISTING)
        })
    }
}