package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

class UpdateDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateDesignTimeArtifact)

    static void main(String[] args) {
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.getEnvironmentVariables()
        updateDesignTimeArtifact.execute()
    }

    @Override
    void getEnvironmentVariables() {
    }

    @Override
    void execute() {
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        def iFlowName = getMandatoryEnvVar('IFLOW_NAME')
        def iFlowDir = getMandatoryEnvVar('IFLOW_DIR')
        
        def packageId = getMandatoryEnvVar('PACKAGE_ID')
        String versionHandling = (System.getenv('VERSION_HANDLING') ?: 'AUTO_INCREMENT')
        if (!['AUTO_INCREMENT', 'MANIFEST'].contains(versionHandling.toUpperCase())) {
            logger.error("🛑 Value ${versionHandling} for environment variable VERSION_HANDLING not in list of accepted values: AUTO_INCREMENT or MANIFEST")
            System.exit(1)
        }
        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('IFLOW_ID')
        validateInputContainsNoSecrets('IFLOW_NAME')
        validateInputContainsNoSecrets('PACKAGE_ID')
        validateInputContainsNoSecrets('PACKAGE_NAME')

        if (versionHandling == 'AUTO_INCREMENT') {
            def currentiFlowVersion = System.getenv('CURR_IFLOW_VER')
            // Get current iFlow Version and bump up the number before upload
            logger.info("Current IFlow Version in Tenant - ${currentiFlowVersion}")
            def matcher = (currentiFlowVersion =~ /(\S+\.)(\d+)\s*/)
            if (matcher.size()) {
                def patchNo = matcher[0][2] as int
                currentiFlowVersion = "${matcher[0][1]}${patchNo + 1}"
            }
            logger.info("New IFlow Version to be updated - ${currentiFlowVersion}")

            // Update the manifest file with new version number
            logger.debug('Updating MANIFEST.MF')
            File manifestFile = new File("${iFlowDir}/META-INF/MANIFEST.MF")
            def manifestContent = manifestFile.getText('UTF-8')
            def updatedContent = manifestContent.replaceFirst(/Bundle-Version: \S+/, "Bundle-Version: ${currentiFlowVersion}")
            manifestFile.setText(updatedContent, 'UTF-8')
        }

        // Zip iFlow directory and encode to Base 64
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(iFlowDir), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)
        designTimeArtifact.update(iFlowContent, iFlowId, iFlowName, packageId, csrfToken)
        logger.info("IFlow ${iFlowId} updated")

        // If runtime has the same version no, then undeploy it, otherwise it gets skipped during deployment
        if (versionHandling == 'MANIFEST') {
            def designtimeVersion = designTimeArtifact.getVersion(iFlowId, 'active', false)
            RuntimeArtifact runtimeArtifact = new RuntimeArtifact(this.httpExecuter)
            def runtimeVersion = runtimeArtifact.getVersion(iFlowId)

            if (runtimeVersion == designtimeVersion) {
                logger.info('Undeploying existing runtime artifact with same version number due to changes in design')
                runtimeArtifact.undeploy(iFlowId, csrfToken)
            }
        }
    }
}