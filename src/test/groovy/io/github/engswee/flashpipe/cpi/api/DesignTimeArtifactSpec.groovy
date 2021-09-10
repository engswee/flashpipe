package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.util.MockExpectation
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.mockserver.integration.ClientAndServer
import spock.lang.Shared
import spock.lang.Specification

class DesignTimeArtifactSpec extends Specification {

    @Shared
    ClientAndServer mockServer
    @Shared
    HTTPExecuter httpExecuter
    @Shared
    DesignTimeArtifact designTimeArtifact
    MockExpectation mockExpectation

    final static String LOCALHOST = 'localhost'

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
        httpExecuter = HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'dummy', 'dummy')
        designTimeArtifact = new DesignTimeArtifact(httpExecuter)
    }

    def setup() {
        this.mockExpectation = MockExpectation.newInstance(LOCALHOST, 9443)
    }

    def cleanup() {
        mockServer.reset()
    }

    def cleanupSpec() {
        mockServer.stop()
    }

    def 'Query - IFlow exists'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')", 200, '{"d": {"Version": "1.0.4"}}')

        when:
        def iFlowExists = designTimeArtifact.getVersion('FlashPipe_IFlow', 'active', true)

        then:
        iFlowExists == '1.0.4'
    }

    def 'Query - IFlow does not exist'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')", 404, '{"error": {"message": {"value": "Integration design time artifact not found"}}}')

        when:
        def iFlowExists = designTimeArtifact.getVersion('FlashPipe_IFlow', 'active', true)

        then:
        iFlowExists == null
    }

    def 'Failure during query call'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')", 500, 'Error')

        when:
        designTimeArtifact.getVersion('FlashPipe_IFlow', 'active', true)

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get design time artifact call failed with response code = 500'
    }

    def 'Successful download'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')/\$value", 200, 'Success')

        when:
        byte[] responseBody = designTimeArtifact.download('FlashPipe_IFlow', 'active')

        then:
        new String(responseBody, 'UTF-8') == 'Success'
    }

    def 'Failure during download'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')/\$value", 400, '')

        when:
        designTimeArtifact.download('FlashPipe_IFlow', 'active')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Download design time artifact call failed with response code = 400'
    }

    def 'Successful upload'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('POST', '/api/v1/IntegrationDesigntimeArtifacts', ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], 201, 'Success')

        when:
        def uploadResponseBody = designTimeArtifact.upload('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        uploadResponseBody == 'Success'
    }

    def 'Failure during upload - CSRF step'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893', 400)

        when:
        designTimeArtifact.upload('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get CSRF Token call failed with response code = 400'
    }

    def 'Failure during upload - upload step'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('POST', '/api/v1/IntegrationDesigntimeArtifacts', ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], 500, '')

        when:
        designTimeArtifact.upload('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Upload design time artifact call failed with response code = 500'
    }

    def 'Successful update'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('PUT', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], 200, '')

        when:
        designTimeArtifact.update('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'Failure during update'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('PUT', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], 500, '')

        when:
        designTimeArtifact.update('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Update design time artifact call failed with response code = 500'
    }


    def 'Successful delete'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('DELETE', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 200, '')

        when:
        designTimeArtifact.delete('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'Failure during delete'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('DELETE', "/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 500, '')

        when:
        designTimeArtifact.delete('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Delete design time artifact call failed with response code = 500'
    }

    def 'Successful deployment'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('POST', '/api/v1/DeployIntegrationDesigntimeArtifact', ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], ['Id': "'FlashPipe_IFlow'", 'Version': "'active'"], 202, '', [:])

        when:
        designTimeArtifact.deploy('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'Failure during deployment'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('POST', '/api/v1/DeployIntegrationDesigntimeArtifact', ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], ['Id': "'FlashPipe_IFlow'", 'Version': "'active'"], 500, '', [:])

        when:
        designTimeArtifact.deploy('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Deploy design time artifact call failed with response code = 500'
    }

    def 'JSON payload generation'() {
        when:
        def output = designTimeArtifact.constructPayload('FlashPipe IFlow', 'FlashPipe_IFlow', 'FlashPipe_Package', 'base64_dummy')

        then:
        def root = new JsonSlurper().parseText(output)
        verifyAll {
            root.Name == 'FlashPipe IFlow'
            root.Id == 'FlashPipe_IFlow'
            root.PackageId == 'FlashPipe_Package'
            root.ArtifactContent == 'base64_dummy'
        }
    }
}