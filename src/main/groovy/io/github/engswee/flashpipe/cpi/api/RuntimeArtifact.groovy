package io.github.engswee.flashpipe.cpi.api


import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class RuntimeArtifact {

    HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(RuntimeArtifact)

    RuntimeArtifact(String scheme, String host, int port, String user, String password) {
        if (!host || !user || !password)
            throw new HTTPExecuterException('Mandatory input host/user/password is missing')
        this.httpExecuter = new HTTPExecuterApacheImpl()
        this.httpExecuter.setBaseURL(scheme, host, port)
        this.httpExecuter.setBasicAuth(user, password)
    }

    String getStatus(String iFlowId) {
        return getDetails(iFlowId, 'Status')
    }

    String getVersion(String iFlowId) {
        try {
            return getDetails(iFlowId, 'Version')
        } catch (ignored) {
            return null
        }
    }

    private String getDetails(String iFlowId, String fieldName) {
        // Get deployed IFlow's status
        logger.info('Get runtime artifact details')
        this.httpExecuter.executeRequest("/api/v1/IntegrationRuntimeArtifacts('${iFlowId}')", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.d."${fieldName}"
        } else {
            logger.info("Response body = ${this.httpExecuter.getResponseBody().getText('UTF8')}")
            throw new HTTPExecuterException("Get runtime artifact call failed with response code = ${code}")
        }
    }

    String getErrorInfo(String iFlowId) {
        // Get deployed IFlow error information
        logger.info('Get runtime artifact error information')
        this.httpExecuter.executeRequest("/api/v1/IntegrationRuntimeArtifacts('${iFlowId}')/ErrorInformation/\$value", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.parameter
        } else {
            logger.info("Response body = ${this.httpExecuter.getResponseBody().getText('UTF8')}")
            throw new HTTPExecuterException("Get runtime artifact error information call failed with response code = ${code}")
        }
    }
}
