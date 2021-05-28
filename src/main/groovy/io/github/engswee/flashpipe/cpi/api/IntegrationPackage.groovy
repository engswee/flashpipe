package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonBuilder
import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class IntegrationPackage {

    final HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(IntegrationPackage)

    IntegrationPackage(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
    }

    boolean iFlowInDraftVersion(String packageId, String iFlowId) {
        // Check version of IFlow
        logger.info("Checking version of IFlow ${iFlowId} in package ${packageId}")
        this.httpExecuter.executeRequest("/api/v1/IntegrationPackages('${packageId}')/IntegrationDesigntimeArtifacts", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            def iFlowMetadata = root.d.results.find { it.Id == iFlowId }
            if (iFlowMetadata) {
                logger.info("Version of IFlow = ${iFlowMetadata.Version}")
                return iFlowMetadata.Version == 'Active'
            } else {
                throw new HTTPExecuterException("IFlow ${iFlowId} not found in package ${packageId}")
            }
        } else
            this.httpExecuter.logError('Get artifacts of IntegrationPackages')
    }

    boolean exists(String packageId) {
        // Check existence of package
        logger.info("Checking existence of package ${packageId}")
        this.httpExecuter.executeRequest("/api/v1/IntegrationPackages('${packageId}')", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            return true
        } else {
            def responseBody = this.httpExecuter.getResponseBody().getText('UTF8')
            if (code == 404) {
                def root = new JsonSlurper().parseText(responseBody)
                if (root.error.message.value == 'Requested entity could not be found.') {
                    return false
                }
            }
            this.httpExecuter.logError('Get IntegrationPackages')
        }
    }

    String create(String packageId, String packageName, CSRFToken csrfToken) {
        // 1 - Get CSRF token
        String token = csrfToken ? csrfToken.get() : ''

        // Create package
        return createPackage(packageId, packageName, token)
    }

    void delete(String packageId, CSRFToken csrfToken) {
        // 1 - Get CSRF token
        String token = csrfToken ? csrfToken.get() : ''

        // Delete package
        this.httpExecuter.executeRequest('DELETE', "/api/v1/IntegrationPackages('${packageId}')", ['x-csrf-token': token], null)
        def code = this.httpExecuter.getResponseCode()
        if (code != 202)
            this.httpExecuter.logError('Delete integration package')
    }

    private String createPackage(String packageId, String packageName, String csrfToken) {
        logger.info('Create integration package')
        def builder = new JsonBuilder()
        builder {
            'Id' packageId
            'Name' packageName
            'ShortText' packageId
            'Version' '1.0.0'
            'SupportedPlatform' 'SAP Cloud Integration'
        }
        def payload = builder.toString()
        logger.debug("Request body = ${payload}")

        this.httpExecuter.executeRequest('POST', '/api/v1/IntegrationPackages', ['x-csrf-token': csrfToken, 'Accept': 'application/json'], null, payload, 'UTF-8', 'application/json')
        def code = this.httpExecuter.getResponseCode()
        if (code != 201)
            this.httpExecuter.logError('Create integration package')

        return this.httpExecuter.getResponseBody().getText('UTF-8')
    }
}