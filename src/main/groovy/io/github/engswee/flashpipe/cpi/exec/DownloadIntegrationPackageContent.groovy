package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.util.PackageSynchroniser
import io.github.engswee.flashpipe.cpi.util.StringUtility
import io.github.engswee.flashpipe.cpi.util.UtilException
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class DownloadIntegrationPackageContent extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(DownloadIntegrationPackageContent)

    String packageId
    String workDir
    String gitSrcDir
    String commitMessage
    String dirNamingType
    String draftHandling
    List includedIds
    List excludedIds
    String scriptCollectionMap
    String normalizeManifestAction
    String normalizeManifestPrefixOrSuffix

    static void main(String[] args) {
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.getEnvironmentVariables()
        try {
            downloadIntegrationPackageContent.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.packageId = getMandatoryEnvVar('PACKAGE_ID')
        this.workDir = getMandatoryEnvVar('WORK_DIR')
        this.gitSrcDir = getMandatoryEnvVar('GIT_SRC_DIR')
        this.commitMessage = System.getenv('COMMIT_MESSAGE')
        this.dirNamingType = (System.getenv('DIR_NAMING_TYPE') ?: 'ID')
        this.draftHandling = (System.getenv('DRAFT_HANDLING') ?: 'SKIP')
        this.includedIds = StringUtility.extractDelimitedValues(System.getenv('INCLUDE_IDS'), ',')
        this.excludedIds = StringUtility.extractDelimitedValues(System.getenv('EXCLUDE_IDS'), ',')
        this.scriptCollectionMap = System.getenv('SCRIPT_COLLECTION_MAP')
        this.normalizeManifestAction = (System.getenv('NORMALIZE_MANIFEST_ACTION') ?: 'NONE')
        this.normalizeManifestPrefixOrSuffix = (System.getenv('NORMALIZE_MANIFEST_PREFIX_SUFFIX') ?: '')
    }

    @Override
    void execute() {
        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('GIT_SRC_DIR', this.gitSrcDir)
        validateInputContainsNoSecrets('COMMIT_MESSAGE', this.commitMessage)
        validateInputContainsNoSecrets('SCRIPT_COLLECTION_MAP', this.scriptCollectionMap)

        if (!['ID', 'NAME'].contains(this.dirNamingType.toUpperCase())) {
            logger.error("🛑 Value ${this.dirNamingType} for environment variable DIR_NAMING_TYPE not in list of accepted values: ID or NAME")
            throw new ExecutionException('Invalid value for DIR_NAMING_TYPE')
        }

        if (!['SKIP', 'ADD', 'ERROR'].contains(this.draftHandling.toUpperCase())) {
            logger.error("🛑 Value ${this.draftHandling} for environment variable DRAFT_HANDLING not in list of accepted values: SKIP, ADD or ERROR")
            throw new ExecutionException('Invalid value for DRAFT_HANDLING')
        }

        if (!['ADD_PREFIX', 'ADD_SUFFIX', 'DELETE_PREFIX', 'DELETE_SUFFIX'].contains(this.normalizeManifestAction.toUpperCase())) {
            logger.error("🛑 Value ${this.normalizeManifestAction} for environment variable NORMALIZE_MANIFEST_ACTION not in list of accepted values: NONE, ADD_PREFIX, ADD_SUFFIX, DELETE_PREFIX or DELETE_SUFFIX")
            throw new ExecutionException('Invalid value for NORMALIZE_MANIFEST_ACTION')
        }

        if (this.includedIds && this.excludedIds) {
            logger.error('🛑 INCLUDE_IDS and EXCLUDE_IDS are mutually exclusive - use only one of them')
            throw new ExecutionException('INCLUDE_IDS and EXCLUDE_IDS are mutually exclusive')
        }

        Map collections = this.scriptCollectionMap?.split(',')?.toList()?.collectEntries {
            String[] pair = it.split('=')
            [(pair[0]): pair[1]]
        }

        try {
            new PackageSynchroniser(this.httpExecuter).sync(this.packageId, this.workDir, this.gitSrcDir, this.includedIds, this.excludedIds, this.draftHandling, this.dirNamingType, collections, this.normalizeManifestAction, this.normalizeManifestPrefixOrSuffix)
        } catch (UtilException e) {
            logger.error("🛑 Error occurred when processing package ${this.packageId}")
            throw new ExecutionException(e.getMessage())
        }
    }
}