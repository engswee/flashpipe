package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.Configuration
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class UpdateConfiguration extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateConfiguration)

    String iFlowId
    String paramFilePath

    static void main(String[] args) {
        UpdateConfiguration updateConfiguration = new UpdateConfiguration()
        updateConfiguration.getEnvironmentVariables()
        try {
            updateConfiguration.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.iFlowId = getMandatoryEnvVar('IFLOW_ID')
        this.paramFilePath = getMandatoryEnvVar('PARAM_FILE')
    }

    @Override
    void execute() {
        Configuration configuration = new Configuration(this.httpExecuter)
        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)

        // Get configured parameters from tenant
        logger.info('Getting current configured parameters of IFlow')
        List tenantParameters = configuration.getParameters(this.iFlowId, 'active')

        // Get parameters from parameters.prop file
        logger.info("Getting parameters from ${this.paramFilePath} file")
        Properties fileParameters = new Properties()
        fileParameters.load(new FileInputStream(this.paramFilePath))

        logger.info('Comparing parameters and updating where necessary')
        def atLeastOneUpdated = false
        // Compare and update where necessary
        tenantParameters.each {
            if (it.DataType != 'custom:schedule') { // TODO - handle translation to Cron
                // Skip updating for schedulers which require translation to Cron values
                String parameterKey = it.ParameterKey
                String tenantValue = it.ParameterValue
                String fileValue = fileParameters.getProperty(parameterKey)
                if (fileValue != null && fileValue != tenantValue) {
                    if (parameterKey.contains('/')) {
                        logger.error("🛑 Parameter name with / character is not possible to be updated via API. Please rename parameter ${parameterKey}")
                        throw new ExecutionException('Parameter not possible to be updated via API')
                    }
                    logger.info("Parameter ${parameterKey} to be updated from ${tenantValue} to ${fileValue}")
                    configuration.update(this.iFlowId, 'active', parameterKey, fileValue, csrfToken)
                    atLeastOneUpdated = true
                }
            }
        }
        if (atLeastOneUpdated) {
            logger.info('🏆 Undeploying existing runtime artifact due to changes in configured parameters')
            RuntimeArtifact runtimeArtifact = new RuntimeArtifact(this.httpExecuter)
            if (runtimeArtifact.getStatus(this.iFlowId, true))
                runtimeArtifact.undeploy(this.iFlowId, csrfToken)
        } else
            logger.info('🏆 No updates required for configured parameters')
    }
}