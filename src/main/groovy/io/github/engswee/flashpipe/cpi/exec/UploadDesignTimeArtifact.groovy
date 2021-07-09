package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

class UploadDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UploadDesignTimeArtifact)
    
    static void main(String[] args) {
        UploadDesignTimeArtifact uploadDesignTimeArtifact = new UploadDesignTimeArtifact()
        uploadDesignTimeArtifact.execute()
    }

    @Override
    void execute() {
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        def iFlowName = getMandatoryEnvVar('IFLOW_NAME')
        def iFlowDir = getMandatoryEnvVar('IFLOW_DIR')
        def packageId = getMandatoryEnvVar('PACKAGE_ID')
        def packageName = getMandatoryEnvVar('PACKAGE_NAME')

        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('IFLOW_ID')
        validateInputContainsNoSecrets('IFLOW_NAME')
        validateInputContainsNoSecrets('PACKAGE_ID')
        validateInputContainsNoSecrets('PACKAGE_NAME')

        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)

        IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
        if (!integrationPackage.exists(packageId)) {
            logger.info("Package ${packageId} does not exist. Creating package...")
            def result = integrationPackage.create(packageId, packageName, csrfToken)
            logger.info("Package ${packageId} created")
            logger.debug("${result}")
        }

        // Zip iFlow directory and encode to Base 64
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(iFlowDir), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        def response = designTimeArtifact.upload(iFlowContent, iFlowId, iFlowName, packageId, csrfToken)
        logger.info("IFlow ${iFlowId} created")
        logger.debug("${response}")
    }
}