package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

class UploadDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UploadDesignTimeArtifact)

    String iFlowId
    String iFlowName
    String iFlowDir
    String packageId
    String packageName

    static void main(String[] args) {
        UploadDesignTimeArtifact uploadDesignTimeArtifact = new UploadDesignTimeArtifact()
        uploadDesignTimeArtifact.getEnvironmentVariables()
        uploadDesignTimeArtifact.execute()
    }

    @Override
    void getEnvironmentVariables() {
        this.iFlowId = getMandatoryEnvVar('IFLOW_ID')
        this.iFlowName = getMandatoryEnvVar('IFLOW_NAME')
        this.iFlowDir = getMandatoryEnvVar('IFLOW_DIR')
        this.packageId = getMandatoryEnvVar('PACKAGE_ID')
        this.packageName = getMandatoryEnvVar('PACKAGE_NAME')
    }

    @Override
    void execute() {
        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('IFLOW_ID', this.iFlowId)
        validateInputContainsNoSecrets('IFLOW_NAME', this.iFlowName)
        validateInputContainsNoSecrets('PACKAGE_ID', this.packageId)
        validateInputContainsNoSecrets('PACKAGE_NAME', this.packageName)

        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)

        IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
        if (!integrationPackage.exists(this.packageId)) {
            logger.info("Package ${this.packageId} does not exist. Creating package...")
            def result = integrationPackage.create(this.packageId, this.packageName, csrfToken)
            logger.info("Package ${this.packageId} created")
            logger.debug("${result}")
        }

        // Zip iFlow directory and encode to Base 64
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(this.iFlowDir), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        def response = designTimeArtifact.upload(iFlowContent, this.iFlowId, this.iFlowName, this.packageId, csrfToken)
        logger.info("IFlow ${this.iFlowId} created")
        logger.debug("${response}")
    }
}