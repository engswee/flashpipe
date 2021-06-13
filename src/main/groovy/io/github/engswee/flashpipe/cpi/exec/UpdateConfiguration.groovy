package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.Configuration
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class UpdateConfiguration extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateConfiguration)

    static void main(String[] args) {
        UpdateConfiguration updateConfiguration = new UpdateConfiguration()
        updateConfiguration.execute()
    }

    @Override
    void execute() {
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        def paramFilePath = getMandatoryEnvVar('PARAM_FILE')

        Configuration configuration = new Configuration(this.httpExecuter)
        CSRFToken csrfToken = this.oauthTokenHost ? null : new CSRFToken(this.httpExecuter)

        // Get configured parameters from tenant
        logger.info('Getting current configured parameters of IFlow')
        Map tenantParameters = configuration.getParameters(iFlowId, 'active')

        // Get parameters from parameters.prop file
        logger.info("Getting parameters from ${paramFilePath} file")
        Properties fileParameters = new Properties()
        fileParameters.load(new FileInputStream(paramFilePath))
        
        logger.info("Comparing parameters and updating where necessary")
        // Compare and update where necessary
        tenantParameters.each { String parameterKey, String tenantValue ->
            String fileValue = fileParameters.getProperty(parameterKey)
            if (fileValue != null && fileValue != tenantValue) {
                logger.info("Parameter ${parameterKey} to be updated from ${tenantValue} to ${fileValue}")
                configuration.update(iFlowId, 'active', parameterKey, fileValue, csrfToken)
            }
        }
    }
}