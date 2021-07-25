package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.cpi.exec.DownloadIntegrationPackageContent
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.Paths
import java.nio.file.StandardCopyOption

class GetTenantSnapshot extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(GetTenantSnapshot)

    static void main(String[] args) {
        GetTenantSnapshot GetTenantSnapshot = new GetTenantSnapshot()
        GetTenantSnapshot.execute()
    }

    @Override
    void execute() {
        def gitSrcDir = getMandatoryEnvVar('GIT_SRC_DIR')
        def host = getMandatoryEnvVar('HOST_TMN')
        def workDir = getMandatoryEnvVar('WORK_DIR')
        
        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('GIT_SRC_DIR')
        validateInputContainsNoSecrets('GIT_BRANCH') //-TODO-: Add Git branch support
        validateInputContainsNoSecrets('COMMIT_MESSAGE')
        
        println '---------------------------------------------------------------------------------'
        logger.info("📢 Begin taking a snapshot of the tenant")
 
        // Get packages from the tentant        
        logger.info("Getting the list of packages of the tenant - ${host}")
        IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
        List packages = integrationPackage.getPackagesList()
        if (packages.size() == 0) {
            logger.error("🛑 No packages found in the tenant")
            System.exit(1)
        }
        
        logger.info("Processing ${packages.size()} packages")
        def result = new StringBuffer()
        def error = new StringBuffer()
        packages.eachWithIndex{it,i->
            def index = i + 1
            def packageId = it.Id.toString()
            println '---------------------------------------------------------------------------------'
            logger.info("Processing package ${index}/${packages.size()} - ID: ${packageId}")            

            //Set arguments for calling the script that will download the package
            def cmd =  [
                "sync_to_git_repository.sh",
                "GIT_SRC_DIR=${gitSrcDir}/packages",
                "PACKAGE_ID=${packageId}",
                "WORK_DIR=${workDir}/${packageId}",
                "DO_NOT_COMMIT=1"
            ]
            cmd << inheritedEnvVarAsArgument('HOST_TMN')
            cmd << inheritedEnvVarAsArgument('BASIC_USERID')
            cmd << inheritedEnvVarAsArgument('BASIC_PASSWORD')
            cmd << inheritedEnvVarAsArgument('HOST_OAUTH')
            cmd << inheritedEnvVarAsArgument('HOST_OAUTH_PATH')
            cmd << inheritedEnvVarAsArgument('OAUTH_CLIENTID')
            cmd << inheritedEnvVarAsArgument('OAUTH_CLIENTSECRET')
            cmd << inheritedEnvVarAsArgument('LOG4J_FILE')
            cmd << inheritedEnvVarAsArgument('DRAFT_HANDLING')
            cmd << inheritedEnvVarAsArgument('DIR_NAMING_TYPE')            
            cmd = cmd.findAll() // remove empty elements

            //Execute the script
            //Download script comes with it's own logger
            def proc = cmd.execute()
            proc.waitForProcessOutput ( result, error )
            if (error.length() > 0) {
                logger.error("🛑 Error while downloading package ${index}/${packages.size()} - ID: ${it.Id}")
                System.out << error.toString()       
                System.out << result.toString()
                System.exit(1)
            }
            System.out << result.toString()
            result.setLength(0)
            error.setLength(0)
        }
        println '---------------------------------------------------------------------------------'
        logger.info("🏆 Completed taking a snapshot of the tenant")
    }
    String inheritedEnvVarAsArgument(String envVarName) {
        def envVar = System.getenv(envVarName)
        if (envVar == null) { return null }
        return envVarName + "=" + envVar
    }
}