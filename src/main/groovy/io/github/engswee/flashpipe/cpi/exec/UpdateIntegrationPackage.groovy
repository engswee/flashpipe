package io.github.engswee.flashpipe.cpi.exec

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.cpi.util.Normalizer
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class UpdateIntegrationPackage extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateIntegrationPackage)

    String packageFilePath
    String normalizePackageAction
    String normalizePackageIDPrefixOrSuffix
    String normalizePackageNamePrefixOrSuffix

    static void main(String[] args) {
        UpdateIntegrationPackage updateIntegrationPackage = new UpdateIntegrationPackage()
        updateIntegrationPackage.getEnvironmentVariables()
        try {
            updateIntegrationPackage.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.packageFilePath = getMandatoryEnvVar('PACKAGE_FILE')
        this.normalizePackageAction = (System.getenv('NORMALIZE_PACKAGE_ACTION') ?: 'NONE')
        this.normalizePackageIDPrefixOrSuffix = (System.getenv('NORMALIZE_PACKAGE_ID_PREFIX_SUFFIX') ?: '')
        this.normalizePackageNamePrefixOrSuffix = (System.getenv('NORMALIZE_PACKAGE_NAME_PREFIX_SUFFIX') ?: '')
    }

    @Override
    void execute() {
        if (!['NONE', 'ADD_PREFIX', 'ADD_SUFFIX', 'DELETE_PREFIX', 'DELETE_SUFFIX'].contains(this.normalizePackageAction.toUpperCase())) {
            logger.error("🛑 Value ${this.normalizePackageAction} for environment variable NORMALIZE_PACKAGE_ACTION not in list of accepted values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX or DELETE_SUFFIX")
            throw new ExecutionException('Invalid value for NORMALIZE_PACKAGE_ACTION')
        }

        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)

        // Get package details from JSON file
        logger.info("Getting package details from ${this.packageFilePath} file")
        Map packageContent = new JsonSlurper().parse(new FileInputStream(this.packageFilePath))

        // Normalize ID & Name
        packageContent.d.Id = Normalizer.normalize(packageContent.d.Id, normalizePackageAction, normalizePackageIDPrefixOrSuffix)
        packageContent.d.Name = Normalizer.normalize(packageContent.d.Name, normalizePackageAction, normalizePackageNamePrefixOrSuffix)

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