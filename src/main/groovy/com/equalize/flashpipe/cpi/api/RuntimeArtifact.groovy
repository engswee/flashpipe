package com.equalize.flashpipe.cpi.api

import com.equalize.flashpipe.http.HTTPExecuter
import com.equalize.flashpipe.http.HTTPExecuterApacheImpl
import com.equalize.flashpipe.http.HTTPExecuterException
import groovy.json.JsonSlurper
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
        // Get deployed IFlow's status
        logger.info('Get runtime artifact status')
        this.httpExecuter.executeRequest("/api/v1/IntegrationRuntimeArtifacts('${iFlowId}')", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.d.Status
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
