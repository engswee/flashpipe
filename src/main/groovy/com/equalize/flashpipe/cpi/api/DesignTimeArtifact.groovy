package com.equalize.flashpipe.cpi.api

import com.equalize.flashpipe.http.HTTPExecuter
import com.equalize.flashpipe.http.HTTPExecuterApacheImpl
import com.equalize.flashpipe.http.HTTPExecuterException
import groovy.json.JsonBuilder
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class DesignTimeArtifact {

    HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(DesignTimeArtifact)

    DesignTimeArtifact(String scheme, String host, int port, String user, String password) {
        if (!host || !user || !password)
            throw new HTTPExecuterException('Mandatory input host/user/password is missing')
        this.httpExecuter = new HTTPExecuterApacheImpl()
        this.httpExecuter.setBaseURL(scheme, host, port)
        this.httpExecuter.setBasicAuth(user, password)
    }

    boolean exists(String iFlowId, String iFlowVersion) {
        logger.info('Query Design time artifact')
        this.httpExecuter.executeRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='$iFlowId',Version='$iFlowVersion')")

        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code == 200)
            return true
        else if (code == 404)
            return false
        else
            throw new HTTPExecuterException("Query design time artifact call failed with response code = ${code}")
    }

    byte[] download(String iFlowId, String iFlowVersion) {
        logger.info('Download Design time artifact')
        this.httpExecuter.executeRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='$iFlowId',Version='$iFlowVersion')/\$value")

        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code == 200) {
            byte[] responseBody = this.httpExecuter.getResponseBody().getBytes()
            return responseBody
        } else
            throw new HTTPExecuterException("Download design time artifact call failed with response code = ${code}")
    }

    void update(String iFlowContent, String iFlowId, String iFlowName, String packageId) {
        // 1 - Get CSRF token
        String csrfToken = getCSRFToken()

        // 2 - Update IFlow
        updateArtifact(iFlowName, iFlowId, packageId, iFlowContent, csrfToken)
    }

    void delete(String iFlowId) {
        // 1 - Get CSRF token
        String csrfToken = getCSRFToken()

        // 2 - Update IFlow
        deleteArtifact(iFlowId, csrfToken)
    }

    String upload(String iFlowContent, String iFlowId, String iFlowName, String packageId) {
        // 1 - Get CSRF token
        String csrfToken = getCSRFToken()

        // 3 - Upload IFlow
        return uploadArtifact(iFlowName, iFlowId, packageId, iFlowContent, csrfToken)
    }

    void deploy(String iFlowId) {
        // 1 - Get CSRF token
        String csrfToken = getCSRFToken()

        // 2 - Deploy IFlow
        logger.info('Deploy design time artifact')
        this.httpExecuter.executeRequest('POST', '/api/v1/DeployIntegrationDesigntimeArtifact', ['x-csrf-token': csrfToken, 'Accept': 'application/json'], ['Id': "'${iFlowId}'", 'Version': "'active'"])
        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code != 202) {
            logger.info("Response body = ${this.httpExecuter.getResponseBody().getText('UTF8')}")
            throw new HTTPExecuterException("Deploy design time artifact call failed with response code = ${code}")
        }
    }

    private String constructPayload(String iFlowName, String iFlowId, String packageId, String iFlowContent) {
        def builder = new JsonBuilder()
        builder {
            'Name' iFlowName
            'Id' iFlowId
            'PackageId' packageId
            'ArtifactContent' iFlowContent
        }
        return builder.toString()
    }

    private String getCSRFToken() {
        logger.info('Get CSRF Token')
        this.httpExecuter.executeRequest('/api/v1/', ['x-csrf-token': 'fetch'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200)
            return this.httpExecuter.getResponseHeader('x-csrf-token')
        else
            throw new HTTPExecuterException("Get CSRF Token call failed with response code = ${code}")
    }

    private void updateArtifact(String iFlowName, String iFlowId, String packageId, String iFlowContent, String csrfToken) {
        logger.info('Update design time artifact')
        def payload = constructPayload(iFlowName, iFlowId, packageId, iFlowContent)
        logger.debug("Request body = ${payload}")
        this.httpExecuter.executeRequest('PUT', "/api/v1/IntegrationDesigntimeArtifacts(Id='${iFlowId}',Version='active')", ['x-csrf-token': csrfToken, 'Accept': 'application/json'], null, payload, 'UTF-8', 'application/json')
        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code != 200) {
            logger.info("Response body = ${this.httpExecuter.getResponseBody().getText('UTF8')}")
            throw new HTTPExecuterException("Update design time artifact call failed with response code = ${code}")
        }
    }

    private void deleteArtifact(String iFlowId, String csrfToken) {
        logger.info('Delete existing design time artifact')
        this.httpExecuter.executeRequest('DELETE', "/api/v1/IntegrationDesigntimeArtifacts(Id='$iFlowId',Version='active')", ['x-csrf-token': csrfToken], null)
        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code != 200)
            throw new HTTPExecuterException("Delete design time artifact call failed with response code = ${code}")
    }

    private String uploadArtifact(String iFlowName, String iFlowId, String packageId, String iFlowContent, String csrfToken) {
        logger.info('Upload design time artifact')
        def payload = constructPayload(iFlowName, iFlowId, packageId, iFlowContent)
        logger.debug("Request body = ${payload}")

        this.httpExecuter.executeRequest('POST', '/api/v1/IntegrationDesigntimeArtifacts', ['x-csrf-token': csrfToken, 'Accept': 'application/json'], null, payload, 'UTF-8', 'application/json')
        def code = this.httpExecuter.getResponseCode()
        logger.info("HTTP Response code = ${code}")
        if (code != 201) {
            logger.info("Response body = ${this.httpExecuter.getResponseBody().getText('UTF8')}")
            throw new HTTPExecuterException("Upload design time artifact call failed with response code = ${code}")
        }

        return this.httpExecuter.getResponseBody().getText('UTF-8')
    }
}