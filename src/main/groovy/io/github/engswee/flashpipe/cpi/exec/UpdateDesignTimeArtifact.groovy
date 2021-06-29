package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

class UpdateDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateDesignTimeArtifact)
    
    static void main(String[] args) {
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.execute()
    }

    @Override
    void execute() {
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        def iFlowName = getMandatoryEnvVar('IFLOW_NAME')
        def iFlowDir = getMandatoryEnvVar('IFLOW_DIR')
        def currentiFlowVersion = getMandatoryEnvVar('CURR_IFLOW_VER')
        def packageId = getMandatoryEnvVar('PACKAGE_ID')

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

        // Zip iFlow directory and encode to Base 64
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(iFlowDir), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        CSRFToken csrfToken = this.oauthTokenHost ? null : new CSRFToken(this.httpExecuter)
        designTimeArtifact.update(iFlowContent, iFlowId, iFlowName, packageId, csrfToken)
        logger.info("IFlow ${iFlowId} updated")
    }
}