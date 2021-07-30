package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.cpi.util.PackageSynchroniser
import io.github.engswee.flashpipe.cpi.util.UtilException
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class GetTenantSnapshot extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(GetTenantSnapshot)

    static void main(String[] args) {
        GetTenantSnapshot GetTenantSnapshot = new GetTenantSnapshot()
        GetTenantSnapshot.execute()
    }

    @Override
    void execute() {
        def gitSrcDir = getMandatoryEnvVar('GIT_SRC_DIR')
        def workDir = getMandatoryEnvVar('WORK_DIR')

        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('GIT_SRC_DIR')
        validateInputContainsNoSecrets('GIT_BRANCH') //-TODO-: Add Git branch support
        validateInputContainsNoSecrets('COMMIT_MESSAGE')

        println '---------------------------------------------------------------------------------'
        logger.info("📢 Begin taking a snapshot of the tenant")

        // Get packages from the tentant
        IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
        List packages = integrationPackage.getPackagesList()
        if (packages.size() == 0) {
            logger.error("🛑 No packages found in the tenant")
            System.exit(1)
        }

        logger.info("Processing ${packages.size()} packages")
        PackageSynchroniser packageSynchroniser = new PackageSynchroniser(this.httpExecuter)
        packages.eachWithIndex { it, i ->
            def index = i + 1
            def packageId = it.Id.toString()
            println '---------------------------------------------------------------------------------'
            logger.info("Processing package ${index}/${packages.size()} - ID: ${packageId}")
            try {
                packageSynchroniser.sync(packageId, "${workDir}/${packageId}", "${gitSrcDir}/${packageId}", [], [], 'SKIP', 'ID')
            } catch (UtilException ignored) {
                logger.error("🛑 Error occurred when processing package ${packageId}")
                System.exit(1)
            }
        }
        println '---------------------------------------------------------------------------------'
        logger.info("🏆 Completed taking a snapshot of the tenant")
    }
}