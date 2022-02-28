package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import io.github.engswee.flashpipe.cpi.util.ManifestHandler
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

class UpdateDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateDesignTimeArtifact)

    String iFlowId
    String iFlowName
    String iFlowDir
    String packageId
    String versionHandling
    String currentiFlowVersion

    static void main(String[] args) {
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.getEnvironmentVariables()
        try {
            updateDesignTimeArtifact.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.iFlowId = getMandatoryEnvVar('IFLOW_ID')
        this.iFlowName = getMandatoryEnvVar('IFLOW_NAME')
        this.iFlowDir = getMandatoryEnvVar('IFLOW_DIR')
        this.packageId = getMandatoryEnvVar('PACKAGE_ID')
        this.versionHandling = (System.getenv('VERSION_HANDLING') ?: 'AUTO_INCREMENT')
        this.currentiFlowVersion = System.getenv('CURR_IFLOW_VER')
    }

    @Override
    void execute() {
        if (!['AUTO_INCREMENT', 'MANIFEST'].contains(this.versionHandling.toUpperCase())) {
            logger.error("🛑 Value ${this.versionHandling} for environment variable VERSION_HANDLING not in list of accepted values: AUTO_INCREMENT or MANIFEST")
            throw new ExecutionException('Invalid entry for VERSION_HANDLING')
        }
        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('IFLOW_ID', this.iFlowId)
        validateInputContainsNoSecrets('IFLOW_NAME', this.iFlowName)
        validateInputContainsNoSecrets('PACKAGE_ID', this.packageId)

        String scriptCollectionMap = System.getenv('SCRIPT_COLLECTION_MAP')
        validateInputContainsNoSecrets('SCRIPT_COLLECTION_MAP', scriptCollectionMap)
        Map collections = scriptCollectionMap?.split(',')?.toList()?.collectEntries {
            String[] pair = it.split('=')
            [(pair[0]): pair[1]]
        }

        // TODO - Move this into ManifestHandler - or it may no longer be needed
        if (this.versionHandling == 'AUTO_INCREMENT') {
            // Get current iFlow Version and bump up the number before upload
            logger.info("Current IFlow Version in Tenant - ${this.currentiFlowVersion}")
            def matcher = (this.currentiFlowVersion =~ /(\S+\.)(\d+)\s*/)
            if (matcher.size()) {
                def patchNo = matcher[0][2] as int
                this.currentiFlowVersion = "${matcher[0][1]}${patchNo + 1}"
            }
            logger.info("New IFlow Version to be updated - ${this.currentiFlowVersion}")

            // Update the manifest file with new version number
            logger.debug('Updating MANIFEST.MF')
            File manifestFile = new File("${this.iFlowDir}/META-INF/MANIFEST.MF")
            def manifestContent = manifestFile.getText('UTF-8')
            def updatedContent = manifestContent.replaceFirst(/Bundle-Version: \S+/, "Bundle-Version: ${this.currentiFlowVersion}")
            manifestFile.setText(updatedContent, 'UTF-8')
        }

        ManifestHandler manifestHandler = new ManifestHandler("${this.iFlowDir}/META-INF/MANIFEST.MF")
        manifestHandler.updateAttributes(this.iFlowId, this.iFlowName, collections.collect { it.value })
        manifestHandler.updateFile()

        // Zip iFlow directory and encode to Base 64
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(this.iFlowDir), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)
        designTimeArtifact.update(iFlowContent, this.iFlowId, this.iFlowName, this.packageId, csrfToken)
        logger.info("IFlow ${this.iFlowId} updated")

        // If runtime has the same version no, then undeploy it, otherwise it gets skipped during deployment
        if (this.versionHandling == 'MANIFEST') { // TODO - no longer need versionhandling?
            def designtimeVersion = designTimeArtifact.getVersion(this.iFlowId, 'active', false)
            RuntimeArtifact runtimeArtifact = new RuntimeArtifact(this.httpExecuter)
            def runtimeVersion = runtimeArtifact.getVersion(this.iFlowId)

            if (runtimeVersion == designtimeVersion) {
                logger.info('Undeploying existing runtime artifact with same version number due to changes in design')
                runtimeArtifact.undeploy(this.iFlowId, csrfToken)
            }
        }
    }
}