package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.util.PackageSynchroniser
import io.github.engswee.flashpipe.cpi.util.UtilException
import org.slf4j.Logger
import org.slf4j.LoggerFactory

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

        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('GIT_SRC_DIR')
        validateInputContainsNoSecrets('COMMIT_MESSAGE')

        String dirNamingType = (System.getenv('DIR_NAMING_TYPE') ?: 'ID')
        if (!['ID', 'NAME'].contains(dirNamingType.toUpperCase())) {
            logger.error("🛑 Value ${dirNamingType} for environment variable DIR_NAMING_TYPE not in list of accepted values: ID or NAME")
            System.exit(1)
        }
        String draftHandling = (System.getenv('DRAFT_HANDLING') ?: 'SKIP')
        if (!['SKIP', 'ADD', 'ERROR'].contains(draftHandling.toUpperCase())) {
            logger.error("🛑 Value ${draftHandling} for environment variable DRAFT_HANDLING not in list of accepted values: SKIP, ADD or ERROR")
            System.exit(1)
        }
        List includedIds = System.getenv('INCLUDE_IDS') ? System.getenv('INCLUDE_IDS').split(',').toList()*.trim() : []
        List excludeIds = System.getenv('EXCLUDE_IDS') ? System.getenv('EXCLUDE_IDS').split(',').toList()*.trim() : []
        if (includedIds && excludeIds) {
            logger.error('🛑 INCLUDE_IDS and EXCLUDE_IDS are mutually exclusive - use only one of them')
            System.exit(1)
        }

        try {
            new PackageSynchroniser(this.httpExecuter).sync(packageId, workDir, gitSrcDir, includedIds, excludeIds, draftHandling, dirNamingType)
        } catch (UtilException ignored) {
            logger.error("🛑 Error occurred when processing package ${packageId}")
            System.exit(1)
        }
    }
}