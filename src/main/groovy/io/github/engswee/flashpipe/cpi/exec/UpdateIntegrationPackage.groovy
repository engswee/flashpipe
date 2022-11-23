package io.github.engswee.flashpipe.cpi.exec

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class UpdateIntegrationPackage extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateIntegrationPackage)

    String packageFilePath

    static void main(String[] args) {
        UpdateIntegrationPackage updateDesignTimeArtifact = new UpdateIntegrationPackage()
        updateDesignTimeArtifact.getEnvironmentVariables()
        try {
            updateDesignTimeArtifact.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.packageFilePath = getMandatoryEnvVar('PACKAGE_FILE')
    }

    @Override
    void execute() {
        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)

        // Get package details from JSON file
        logger.info("Getting package details from ${this.packageFilePath} file")
        Map packageContent = new JsonSlurper().parse(new FileInputStream(this.packageFilePath))
        String packageId = packageContent.d.Id

        IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
        if (!integrationPackage.exists(packageId)) {
            logger.info("Package ${packageId} does not exist. Creating package...")
            def result = integrationPackage.create(packageContent.d, csrfToken)
            logger.info("Package ${packageId} created")
            logger.debug("${result}")
        } else {
            // Update integration package
            logger.info("Updating package ${packageId}")
            integrationPackage.update(packageContent.d, csrfToken)
            logger.info("Package ${packageId} updated")
        }
    }
}