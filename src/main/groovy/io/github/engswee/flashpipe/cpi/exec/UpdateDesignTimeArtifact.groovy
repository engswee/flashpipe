package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import org.zeroturnaround.zip.ZipUtil

class UpdateDesignTimeArtifact extends APIExecuter {

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
        println "[INFO] Current IFlow Version in Tenant - ${currentiFlowVersion}"
        def matcher = (currentiFlowVersion =~ /(\S+\.)(\d+)\s*/)
        if (matcher.size()) {
            def patchNo = matcher[0][2] as int
            currentiFlowVersion = "${matcher[0][1]}${patchNo + 1}"
        }
        println "[INFO] New IFlow Version to be updated - ${currentiFlowVersion}"

        // Update the manifest file with new version number
        println "[INFO] Updating MANIFEST.MF"
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
    }
}