package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import io.github.engswee.flashpipe.cpi.util.StringUtility
import org.slf4j.Logger
import org.slf4j.LoggerFactory

import java.util.concurrent.TimeUnit

class DeployDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(DeployDesignTimeArtifact)

    List<String> iFlows
    int delayLength
    int maxCheckLimit
    boolean compareVersions

    static void main(String[] args) {
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.getEnvironmentVariables()
        try {
            deployDesignTimeArtifact.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        // Get list of IFlow IDs to be processed
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        this.iFlows = StringUtility.extractDelimitedValues(iFlowId, ',')

        this.delayLength = (System.getenv('DELAY_LENGTH') ?: 30) as int
        this.maxCheckLimit = (System.getenv('MAX_CHECK_LIMIT') ?: 10) as int
        this.compareVersions = (System.getenv('COMPARE_VERSIONS') ?: true)
    }

    @Override
    void execute() {
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        RuntimeArtifact runtimeArtifact = new RuntimeArtifact(this.httpExecuter)

        // Loop and deploy each IFlow
        this.iFlows.eachWithIndex { String id, index ->
            logger.info("Processing IFlow ${index + 1} - ${id}")
            deploySingleIFlow(designTimeArtifact, id, runtimeArtifact)
        }

        // Delay to allow deployment to start before checking the status
        // Only applicable if there is only 1 IFlow, because if there are many, then there is an inherent delay already
        if (this.iFlows.size() == 1)
            TimeUnit.SECONDS.sleep(this.delayLength)

        // Check deployment status of IFlows
        this.iFlows.eachWithIndex { String id, index ->
            checkDeploymentStatus(this.delayLength, this.maxCheckLimit, runtimeArtifact, id)
            logger.info("IFlow ${index + 1} - ${id} deployed successfully")
        }

        logger.info('???? IFlow(s) deployment completed successfully')
    }

    private void deploySingleIFlow(DesignTimeArtifact designTimeArtifact, String iFlowId, RuntimeArtifact runtimeArtifact) {
        def designtimeVersion = designTimeArtifact.getVersion(iFlowId, 'active', false)
        if (this.compareVersions) {
            // Compare designtime version with runtime version to determine if deployment is needed
            logger.info('Comparing designtime version with runtime version')
            def runtimeVersion = runtimeArtifact.getVersion(iFlowId)

            if (runtimeVersion == designtimeVersion) {
                logger.info("IFlow ${iFlowId} with version ${runtimeVersion} already deployed. Skipping runtime deployment")
            } else {
                CSRFToken csrfToken = new CSRFToken(this.httpExecuter)
                logger.info("???? IFlow previously not deployed, or versions differ. Proceeding to deploy IFlow ${iFlowId} with version ${designtimeVersion}")
                designTimeArtifact.deploy(iFlowId, csrfToken)
                logger.info("IFlow ${iFlowId} deployment triggered")
            }
        } else {
            CSRFToken csrfToken = new CSRFToken(this.httpExecuter)
            logger.info("???? Proceeding to deploy IFlow ${iFlowId} with version ${designtimeVersion}")
            designTimeArtifact.deploy(iFlowId, csrfToken)
            logger.info("IFlow ${iFlowId} deployment triggered")
        }
    }

    private void checkDeploymentStatus(int delayLength, int maxCheckLimit, RuntimeArtifact runtimeArtifact, String iFlowId) {
        logger.info("Checking deployment status for IFlow ${iFlowId} every ${delayLength} seconds up to ${maxCheckLimit} times")
        int checkCounter = 0
        while (true) { // TODO - Switch to do - while loop in Groovy 3.x
            def status = runtimeArtifact.getStatus(iFlowId)
            logger.info("Check ${checkCounter} - Current IFlow status = ${status}")
            if (status != 'STARTING') {
                if (status != 'STARTED') {
                    def errorMessage = runtimeArtifact.getErrorInfo(iFlowId)
                    logger.error("???? IFlow deployment unsuccessful, ended with status ${status}")
                    logger.error("???? Error message = ${errorMessage}")
                    throw new ExecutionException(errorMessage)
                } else {
                    break
                }
            }
            checkCounter++
            if (checkCounter == maxCheckLimit && status != 'STARTED') {
                logger.error("???? IFlow status remained in ${status} after ${maxCheckLimit} checks")
                throw new ExecutionException('Max check limit reached')
            }
            TimeUnit.SECONDS.sleep(delayLength)
        }
    }
}