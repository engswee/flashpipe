package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.cpi.util.PackageSynchroniser
import io.github.engswee.flashpipe.cpi.util.UtilException
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class GetTenantSnapshot extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(GetTenantSnapshot)

    String workDir
    String gitSrcDir
    String commitMessage
    String draftHandling

    static void main(String[] args) {
        GetTenantSnapshot getTenantSnapshot = new GetTenantSnapshot()
        getTenantSnapshot.getEnvironmentVariables()
        try {
            getTenantSnapshot.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.gitSrcDir = getMandatoryEnvVar('GIT_SRC_DIR')
        this.workDir = getMandatoryEnvVar('WORK_DIR')
        this.commitMessage = System.getenv('COMMIT_MESSAGE')
        this.draftHandling = (System.getenv('DRAFT_HANDLING') ?: 'SKIP')
    }

    @Override
    void execute() {
        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('GIT_SRC_DIR', this.gitSrcDir)
//        validateInputContainsNoSecrets('GIT_BRANCH', 'TODO') //-TODO-: Add Git branch support
        validateInputContainsNoSecrets('COMMIT_MESSAGE', this.commitMessage)

        if (!['SKIP', 'ADD', 'ERROR'].contains(this.draftHandling.toUpperCase())) {
            logger.error("🛑 Value ${this.draftHandling} for environment variable DRAFT_HANDLING not in list of accepted values: SKIP, ADD or ERROR")
            throw new ExecutionException('Invalid value for DRAFT_HANDLING')
        }

        println '---------------------------------------------------------------------------------'
        logger.info("📢 Begin taking a snapshot of the tenant")

        // Get packages from the tenant
        IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
        List packages = integrationPackage.getPackagesList()
        if (packages.size() == 0) {
            logger.error("🛑 No packages found in the tenant")
            throw new ExecutionException('No packages found in the tenant')
        }

        logger.info("Processing ${packages.size()} packages")
        PackageSynchroniser packageSynchroniser = new PackageSynchroniser(this.httpExecuter)
        packages.eachWithIndex { it, i ->
            def index = i + 1
            def packageId = it.Id.toString()
            println '---------------------------------------------------------------------------------'
            logger.info("Processing package ${index}/${packages.size()} - ID: ${packageId}")
            try {
                packageSynchroniser.sync(packageId, "${this.workDir}/${packageId}", "${this.gitSrcDir}/${packageId}", [], [], this.draftHandling, 'ID', '', 'NONE', '')
            } catch (UtilException e) {
                logger.error("🛑 Error occurred when processing package ${packageId}")
                throw new ExecutionException(e.getMessage())
            }
        }
        println '---------------------------------------------------------------------------------'
        logger.info("🏆 Completed taking a snapshot of the tenant")
    }
}