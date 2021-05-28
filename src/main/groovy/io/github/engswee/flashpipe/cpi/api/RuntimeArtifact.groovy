package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class RuntimeArtifact {

    final HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(RuntimeArtifact)

    RuntimeArtifact(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
    }

    String getStatus(String iFlowId) {
        return getDetails(iFlowId, 'Status', false)
    }

    String getVersion(String iFlowId) {
        return getDetails(iFlowId, 'Version', true)
    }

    private String getDetails(String iFlowId, String fieldName, boolean skipNotFoundException) {
        // Get deployed IFlow's details
        logger.info('Get runtime artifact details')
        this.httpExecuter.executeRequest("/api/v1/IntegrationRuntimeArtifacts('${iFlowId}')", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.d."${fieldName}"
        } else if (skipNotFoundException && code == 404) {
            def error = new XmlSlurper().parse(this.httpExecuter.getResponseBody())
            if (error.message == 'Requested entity could not be found.') {
                return null
            } else
                this.httpExecuter.logError('Get runtime artifact')
        } else
            this.httpExecuter.logError('Get runtime artifact')
    }

    String getErrorInfo(String iFlowId) {
        // Get deployed IFlow error information
        logger.info('Get runtime artifact error information')
        this.httpExecuter.executeRequest("/api/v1/IntegrationRuntimeArtifacts('${iFlowId}')/ErrorInformation/\$value", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.parameter
        } else
            this.httpExecuter.logError('Get runtime artifact error information')
    }
}