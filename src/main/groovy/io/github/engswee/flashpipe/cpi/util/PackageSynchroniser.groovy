package io.github.engswee.flashpipe.cpi.util

import groovy.json.JsonOutput
import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.cpi.exec.BPMN2Handler
import io.github.engswee.flashpipe.http.HTTPExecuter
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

class PackageSynchroniser {
    final HTTPExecuter httpExecuter
    final IntegrationPackage integrationPackage

    static Logger logger = LoggerFactory.getLogger(PackageSynchroniser)

    PackageSynchroniser(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
        this.integrationPackage = new IntegrationPackage(this.httpExecuter)
    }

    void syncInfo(String packageId, String workDir, String gitSrcDir, String normalizePackageAction, String normalizePackageIDPrefixOrSuffix, String normalizePackageNamePrefixOrSuffix) {
        logger.info("Processing details of integration package ${packageId}")
        if (this.integrationPackage.isReadOnly(packageId)) {
            logger.warn("⚠️ Skipping package ${packageId} as it is Configure-only")
            return
        }
        // Create temp directory in working dir
        new File("${workDir}/from_tenant").mkdirs()

        // Get details from tenant
        String packageContent = this.integrationPackage.getDetails(packageId)
        Map packageFromTenant = new JsonSlurper().parseText(packageContent)

        // Normalize ID
        String normalizedPackageID = Normalizer.normalize(packageId, normalizePackageAction, normalizePackageIDPrefixOrSuffix)
        // Normalize Name
        String normalizedPackageName = Normalizer.normalize(packageFromTenant.d.Name, normalizePackageAction, normalizePackageNamePrefixOrSuffix)

        logger.info("Storing package details from tenant for comparison")
        File packageFile = new File("${workDir}/from_tenant/${normalizedPackageID}.json")
        if (normalizePackageAction != 'NONE') {
            logger.info("Normalizing Package ID from '${packageId}' to '${normalizedPackageID}'")
            packageFromTenant.d.Id = normalizedPackageID
            logger.info("Normalizing Package Name from '${packageFromTenant.d.Name}' to '${normalizedPackageName}'")
            packageFromTenant.d.Name = normalizedPackageName
        }
        // Update file content with normalized values
        packageFile.bytes = JsonOutput.toJson(packageFromTenant).getBytes('UTF-8')

        // Get existing package info file if it exists and compare values
        String gitSourceFile = "${gitSrcDir}/${normalizedPackageID}.json"
        if (new File(gitSourceFile).exists()) {
            Map packageFromGit = new JsonSlurper().parse(new File(gitSourceFile))
            if (contentDiffer(packageFromGit.d, packageFromTenant.d)) {
                logger.info('🏆 Changes detected and will be updated to Git')
                FileUtility.replaceFile("${workDir}/from_tenant/${normalizedPackageID}.json", gitSourceFile)
            } else {
                logger.info('🏆 No changes detected. Update to Git not required')
            }
        } else {
            logger.info("🏆 Saving new file to Git")
            FileUtility.copyFile("${workDir}/from_tenant/${normalizedPackageID}.json", gitSrcDir)
        }
    }

    void syncArtifacts(String packageId, String workDir, String gitSrcDir, List<String> includedIds, List<String> excludedIds, String draftHandling, String dirNamingType, String scriptCollectionMap, String normalizeManifestAction, String normalizeManifestPrefixOrSuffix) {
        // Get all design time artifacts of package
        logger.info("Getting artifacts in integration package ${packageId}")
        // Verify the package is downloadable
        if (this.integrationPackage.isReadOnly(packageId)) {
            logger.warn("⚠️ Skipping package ${packageId} as it is Configure-only and cannot be downloaded")
            return
        }

        List artifacts = this.integrationPackage.getIFlowsWithDraftState(packageId)
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)

        // Create temp directories in working dir
        new File("${workDir}/download").mkdirs()
        new File("${workDir}/from_git").mkdirs()
        new File("${workDir}/from_tenant").mkdirs()

        List filteredArtifacts = filterArtifacts(artifacts, includedIds, excludedIds)

        // Process through the artifacts
        for (Map artifact : filteredArtifacts) {
            println '---------------------------------------------------------------------------------'
            logger.info("📢 Begin processing for artifact ${artifact.id}")
            // Check if artifact is in draft version
            if (artifact.isDraft) {
                switch (draftHandling.toUpperCase()) {
                    case 'SKIP':
                        logger.warn("⚠️ Integration artifact ${artifact.id} is in draft version, and will be skipped")
                        continue
                    case 'ADD':
                        logger.info("Integration artifact ${artifact.id} is in draft version, and will be added")
                        break
                    case 'ERROR':
                        logger.error("🛑 Integration artifact ${artifact.id} is in draft version, Save Version in Web UI first!")
                        throw new UtilException('Artifact in draft version')
                }
            }
            // Download IFlow
            logger.info("Downloading IFlow ${artifact.id} from tenant for comparison")
            File outputZip = new File("${workDir}/download/${artifact.id}.zip")
            outputZip.bytes = designTimeArtifact.download(artifact.id, 'active')
            logger.info("IFlow ${artifact.id} downloaded to ${outputZip}")

            String normalizedIFlowID = Normalizer.normalize(artifact.id, normalizeManifestAction, normalizeManifestPrefixOrSuffix)
            String normalizedIFlowName = Normalizer.normalize(artifact.name, normalizeManifestAction, normalizeManifestPrefixOrSuffix)
            logger.debug("Normalized IFlow ID - ${normalizedIFlowID}")
            logger.debug("Normalized IFlow Name - ${normalizedIFlowName}")

            // Unzip IFlow contents
            def directoryName = (dirNamingType.toUpperCase() == 'NAME') ? normalizedIFlowName : normalizedIFlowID
            logger.debug("Target IFlow Directory Name - ${directoryName}")
            ZipUtil.unpack(outputZip, new File("${workDir}/download/${directoryName}"))
            logger.info("Downloaded IFlow artifact unzipped to ${workDir}/download/${directoryName}")

            // Normalize MANIFEST.MF before sync to Git
            ScriptCollection scriptCollection = ScriptCollection.newInstance(scriptCollectionMap)
            Map collections = scriptCollection.getCollections()
            ManifestHandler.newInstance("${workDir}/download/${directoryName}/META-INF/MANIFEST.MF").normalizeAttributesInFile(normalizedIFlowID, normalizedIFlowName, scriptCollection.getTargetCollectionValues())

            // Normalize the script collection in IFlow BPMN2 XML before syncing to Git
            if (collections.size()) {
                BPMN2Handler bpmn2Handler = new BPMN2Handler()
                bpmn2Handler.updateFiles(collections, "${workDir}/download/${directoryName}")
            }

            // (1) If IFlow already exists in Git, then compare and update
            if (new File("${gitSrcDir}/${directoryName}").exists()) {
                logger.info("Comparing content from tenant against Git")
                // Copy to temp directory for diff comparison
                FileUtility.copyDirectory("${workDir}/download/${directoryName}/src/main/resources", "${workDir}/from_tenant/${directoryName}/src/main/resources")
                FileUtility.copyDirectory("${workDir}/download/${directoryName}/META-INF", "${workDir}/from_tenant/${directoryName}/META-INF")
                FileUtility.copyDirectory("${gitSrcDir}/${directoryName}/src/main/resources", "${workDir}/from_git/${directoryName}/src/main/resources")
                FileUtility.copyDirectory("${gitSrcDir}/${directoryName}/META-INF", "${workDir}/from_git/${directoryName}/META-INF")

                // Remove comments from parameters.prop before comparison only if it exists
                File tenantParamFile = new File("${workDir}/from_tenant/${directoryName}/src/main/resources/parameters.prop")
                File gitParamFile = new File("${workDir}/from_git/${directoryName}/src/main/resources/parameters.prop")
                if (tenantParamFile.exists() && gitParamFile.exists()) {
                    FileUtility.removeCommentsFromFile(tenantParamFile)
                    FileUtility.removeCommentsFromFile(gitParamFile)
                }

                // Execute shell command diff to compare directory contents
                ShellCommand shellCommand = new ShellCommand('bash')
                String command = "diff -I '^Origin.*' --strip-trailing-cr -qr -x '.DS_Store' '${workDir}/from_tenant/${directoryName}' '${workDir}/from_git/${directoryName}'"
                shellCommand.execute(command)
                switch (shellCommand.getExitValue()) {
                    case 0:
                        logger.info('🏆 No changes detected. Update to Git not required')
                        break
                    case 1:
                        println shellCommand.getOutputText()
                        logger.info('🏆 Changes detected and will be updated to Git')
                        // Update the changes into the Git directory
                        // (a) Replace /src/main/resources
                        FileUtility.replaceDirectory("${workDir}/download/${directoryName}/src/main/resources", "${gitSrcDir}/${directoryName}/src/main/resources")
                        // (b) Replace /META-INF/MANIFEST.MF
                        FileUtility.replaceFile("${workDir}/download/${directoryName}/META-INF/MANIFEST.MF", "${gitSrcDir}/${directoryName}/META-INF/MANIFEST.MF")
                        break
                    default:
                        logger.error("🛑 ${shellCommand.getErrorText()}")
                        throw new UtilException('Error executing shell command')
                }
            } else {
                // (2) If IFlow does not exist in Git, then add it
                if (!new File(gitSrcDir).exists()) {
                    new File(gitSrcDir).mkdirs()
                }
                logger.info("🏆 Artifact ${artifact.id} does not exist, and will be added to Git")
                FileUtility.copyDirectory("${workDir}/download/${directoryName}", "${gitSrcDir}/${directoryName}")
            }
        }
        // Clean up working directory
        new File("${workDir}/download").deleteDir()
        new File("${workDir}/from_git").deleteDir()
        new File("${workDir}/from_tenant").deleteDir()
        println '---------------------------------------------------------------------------------'
        logger.info("🏆 Completed processing of artifacts in integration package ${packageId}")
    }

    boolean contentDiffer(Map source, Map target) {
        if (source.Name.toString() != target.Name.toString()) {
            return true
        }
        if (source.Description.toString() != target.Description.toString()) {
            return true
        }
        if (source.ShortText.toString() != target.ShortText.toString()) {
            return true
        }
        if (source.Version.toString() != target.Version.toString()) {
            return true
        }
        if (source.Vendor.toString() != target.Vendor.toString()) {
            return true
        }
        if (source.Mode.toString() != target.Mode.toString()) {
            return true
        }
        if (source.SupportedPlatform.toString() != target.SupportedPlatform.toString()) {
            return true
        }
        if (source.Products.toString() != target.Products.toString()) {
            return true
        }
        if (source.Keywords.toString() != target.Keywords.toString()) {
            return true
        }
        if (source.Countries.toString() != target.Countries.toString()) {
            return true
        }
        if (source.Industries.toString() != target.Industries.toString()) {
            return true
        }
        if (source.LineOfBusiness.toString() != target.LineOfBusiness.toString()) {
            return true
        }
        return false
    }

    private List filterArtifacts(List artifacts, List includedIds, List excludedIds) {
        if (includedIds) {
            List outputList = []
            includedIds.each { iFlowId ->
                Map artifactDetails = artifacts.find { it.id == iFlowId }
                if (!artifactDetails) {
                    logger.error("🛑 IFlow ${iFlowId} in INCLUDE_IDS does not exist")
                    throw new UtilException('Invalid input in INCLUDE_IDS')
                } else {
                    outputList.add(artifactDetails)
                }
            }
            logger.info("Include only IFlow with IDs - ${includedIds.join(',')}")
            return outputList
        } else if (excludedIds) {
            List outputList = []
            // Check if the Ids are valid
            excludedIds.each { iFlowId ->
                Map artifactDetails = artifacts.find { it.id == iFlowId }
                if (!artifactDetails) {
                    logger.error("🛑 IFlow ${iFlowId} in EXCLUDE_IDS does not exist")
                    throw new UtilException('Invalid input in EXCLUDE_IDS')
                }
            }
            logger.info("Exclude IFlow with IDs - ${excludedIds.join(',')}")
            artifacts.each { artifact ->
                if (!excludedIds.contains(artifact.id)) {
                    outputList.add(artifact)
                }
            }
            return outputList
        } else {
            return artifacts
        }
    }
}