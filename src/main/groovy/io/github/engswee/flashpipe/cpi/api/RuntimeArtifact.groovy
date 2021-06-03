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
        Map responseRoot = getDetails(iFlowId, false)
        return responseRoot.d.Status
    }

    String getVersion(String iFlowId) {
        Map responseRoot = getDetails(iFlowId, true)
        if (responseRoot?.d?.Status == 'STARTED') {
            return responseRoot.d.Version
        } else {
            return null
        }
    }

    private Map getDetails(String iFlowId, boolean skipNotFoundException) {
        // Get deployed IFlow's details
        logger.info('Get runtime artifact details')
        this.httpExecuter.executeRequest("/api/v1/IntegrationRuntimeArtifacts('${iFlowId}')", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            return new JsonSlurper().parse(this.httpExecuter.getResponseBody())
        } else if (skipNotFoundException && code == 404) {
            return [:]
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