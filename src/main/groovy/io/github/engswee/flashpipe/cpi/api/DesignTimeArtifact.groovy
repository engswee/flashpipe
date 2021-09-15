package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonBuilder
import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class DesignTimeArtifact {

    final HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(DesignTimeArtifact)

    DesignTimeArtifact(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
    }

    String getVersion(String iFlowId, String iFlowVersion, boolean skipNotFoundException) {
        logger.debug('Get Design time artifact')
        this.httpExecuter.executeRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='$iFlowId',Version='$iFlowVersion')", ['Accept': 'application/json'])

        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.d.Version
        } else if (skipNotFoundException && code == 404) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            if (root.error.message.value == 'Integration design time artifact not found') {
                return null
            } else
                this.httpExecuter.logError('Get design time artifact')
        } else
            this.httpExecuter.logError('Get design time artifact')
    }

    byte[] download(String iFlowId, String iFlowVersion) {
        logger.debug("Download Design time artifact ${iFlowId}")
        this.httpExecuter.executeRequest("/api/v1/IntegrationDesigntimeArtifacts(Id='$iFlowId',Version='$iFlowVersion')/\$value")

        def code = this.httpExecuter.getResponseCode()
        if (code == 200)
            return this.httpExecuter.getResponseBody().getBytes()
        else
            this.httpExecuter.logError('Download design time artifact')
    }

    void update(String iFlowContent, String iFlowId, String iFlowName, String packageId, CSRFToken csrfToken) {
        // 1 - Get CSRF token
        String token = csrfToken.get()

        // 2 - Update IFlow
        updateArtifact(iFlowName, iFlowId, packageId, iFlowContent, token)
    }

    void delete(String iFlowId, CSRFToken csrfToken) {
        // 1 - Get CSRF token
        String token = csrfToken.get()

        // 2 - Update IFlow
        deleteArtifact(iFlowId, token)
    }

    String upload(String base64EncodedIFlowContent, String iFlowId, String iFlowName, String packageId, CSRFToken csrfToken) {
        // 1 - Get CSRF token
        String token = csrfToken.get()

        // 3 - Upload IFlow
        return uploadArtifact(iFlowName, iFlowId, packageId, base64EncodedIFlowContent, token)
    }

    void deploy(String iFlowId, CSRFToken csrfToken) {
        // 1 - Get CSRF token
        String token = csrfToken.get()

        // 2 - Deploy IFlow
        logger.debug('Deploy design time artifact')
        this.httpExecuter.executeRequest('POST', '/api/v1/DeployIntegrationDesigntimeArtifact', ['x-csrf-token': token, 'Accept': 'application/json'], ['Id': "'${iFlowId}'", 'Version': "'active'"])
        def code = this.httpExecuter.getResponseCode()
        if (code != 202)
            this.httpExecuter.logError('Deploy design time artifact')
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

    private void updateArtifact(String iFlowName, String iFlowId, String packageId, String iFlowContent, String token) {
        logger.info("Update design time artifact ${iFlowId}")
        def payload = constructPayload(iFlowName, iFlowId, packageId, iFlowContent)
        logger.debug("Request body = ${payload}")
        this.httpExecuter.executeRequest('PUT', "/api/v1/IntegrationDesigntimeArtifacts(Id='${iFlowId}',Version='active')", ['x-csrf-token': token, 'Accept': 'application/json'], null, payload, 'UTF-8', 'application/json')
        def code = this.httpExecuter.getResponseCode()
        if (code != 200)
            this.httpExecuter.logError('Update design time artifact')
    }

    private void deleteArtifact(String iFlowId, String token) {
        logger.debug('Delete existing design time artifact')
        this.httpExecuter.executeRequest('DELETE', "/api/v1/IntegrationDesigntimeArtifacts(Id='$iFlowId',Version='active')", ['x-csrf-token': token], null)
        def code = this.httpExecuter.getResponseCode()
        if (code != 200)
            this.httpExecuter.logError('Delete design time artifact')
    }

    private String uploadArtifact(String iFlowName, String iFlowId, String packageId, String iFlowContent, String token) {
        logger.info("Upload design time artifact ${iFlowId}")
        def payload = constructPayload(iFlowName, iFlowId, packageId, iFlowContent)
        logger.debug("Request body = ${payload}")

        this.httpExecuter.executeRequest('POST', '/api/v1/IntegrationDesigntimeArtifacts', ['x-csrf-token': token, 'Accept': 'application/json'], null, payload, 'UTF-8', 'application/json')
        def code = this.httpExecuter.getResponseCode()
        if (code != 201)
            this.httpExecuter.logError('Upload design time artifact')

        return this.httpExecuter.getResponseBody().getText('UTF-8')
    }
}