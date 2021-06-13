package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonBuilder
import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class Configuration {

    final HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(Configuration)

    Configuration(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
    }

    List getParameters(String iFlowId, String iFlowVersion) {
        logger.info('Get configuration parameters')
        this.httpExecuter.executeRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='$iFlowId',Version='$iFlowVersion')/Configurations", ['Accept': 'application/json'])

        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.d.results
        } else
            this.httpExecuter.logError('Get configuration parameters')

    }

    void update(String iFlowId, String iFlowVersion, String parameterKey, String parameterValue, CSRFToken csrfToken) {
        // 1 - Get CSRF token
        String token = csrfToken ? csrfToken.get() : ''

        logger.info("Update configuration parameter ${parameterKey}")
        def builder = new JsonBuilder()
        builder {
            'ParameterValue' parameterValue
        }
        def payload = builder.toString()
        logger.debug("Request body = ${payload}")
        this.httpExecuter.executeRequest('PUT', "/api/v1/IntegrationDesigntimeArtifacts(Id='${iFlowId}',Version='${iFlowVersion}')/\$links/Configurations('${parameterKey}')", ['x-csrf-token': token], null, payload, 'UTF-8', 'application/json')
        def code = this.httpExecuter.getResponseCode()
        if (code != 202)
            this.httpExecuter.logError("Update configuration parameter ${parameterKey}")
    }
}
