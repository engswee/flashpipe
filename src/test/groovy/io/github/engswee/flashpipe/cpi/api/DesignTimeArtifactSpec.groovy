package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.mockserver.client.MockServerClient
import org.mockserver.integration.ClientAndServer
import org.mockserver.model.HttpRequest
import org.mockserver.model.HttpResponse
import spock.lang.Shared
import spock.lang.Specification

class DesignTimeArtifactSpec extends Specification {

    @Shared
    ClientAndServer mockServer
    @Shared
    HTTPExecuter httpExecuter

    final static String LOCALHOST = 'localhost'

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
        httpExecuter = HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'dummy', 'dummy')
    }

    def setup() {
        mockServer.reset()
    }

    def cleanupSpec() {
        mockServer.stop()
    }

    def setupCSRFTokenExpectation(MockServerClient mockServerClient, int httpResponseStatusCode) {
        def request = HttpRequest.request()
                .withMethod('GET')
                .withPath('/api/v1/')
                .withHeader('x-csrf-token', 'fetch')
        def response = HttpResponse.response()
                .withStatusCode(httpResponseStatusCode)
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
        mockServerClient.when(request).respond(response)
    }

    def setupCSRFTokenExpectation(MockServerClient mockServerClient) {
        setupCSRFTokenExpectation(mockServerClient, 200)
    }

    def 'Query - IFlow exists'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')")
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withBody('{"d": {"Version": "1.0.4"}}')
        mockServerClient.when(httpRequest).respond(httpResponse)

        when:
        def iFlowExists = designTimeArtifact.getVersion('FlashPipe_IFlow', 'active', true)

        then:
        iFlowExists == '1.0.4'
    }

    def 'Query - IFlow does not exist'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')")
        def httpResponse = HttpResponse.response()
                .withStatusCode(404)
                .withBody('{"error": {"message": {"value": "Integration design time artifact not found"}}}')
        mockServerClient.when(httpRequest).respond(httpResponse)

        when:
        def iFlowExists = designTimeArtifact.getVersion('FlashPipe_IFlow', 'active', true)

        then:
        iFlowExists == null
    }

    def 'Failure during query call'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')")
        def httpResponse = HttpResponse.response()
                .withStatusCode(500)
                .withBody('Error')
        mockServerClient.when(httpRequest).respond(httpResponse)

        when:
        designTimeArtifact.getVersion('FlashPipe_IFlow', 'active', true)

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get design time artifact call failed with response code = 500'
    }

    def 'Successful download'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')/\$value")
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withBody('Success')
        mockServerClient.when(httpRequest).respond(httpResponse)

        when:
        byte[] responseBody = designTimeArtifact.download('FlashPipe_IFlow', 'active')

        then:
        new String(responseBody, 'UTF-8') == 'Success'
    }

    def 'Failure during download'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')/\$value")
        def httpResponse = HttpResponse.response()
                .withStatusCode(400)
        mockServerClient.when(httpRequest).respond(httpResponse)

        when:
        designTimeArtifact.download('FlashPipe_IFlow', 'active')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Download design time artifact call failed with response code = 400'
    }

    def 'Successful upload'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('POST')
                .withPath('/api/v1/IntegrationDesigntimeArtifacts')
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
                .withHeader('Accept', 'application/json')
        def response = HttpResponse.response()
                .withStatusCode(201)
                .withBody('Success')
        mockServerClient.when(request).respond(response)

        when:
        def uploadResponseBody = designTimeArtifact.upload('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        uploadResponseBody == 'Success'
    }

    def 'Failure during upload - CSRF step'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient, 400)

        when:
        designTimeArtifact.upload('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get CSRF Token call failed with response code = 400'
    }

    def 'Failure during upload - upload step'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('POST')
                .withPath('/api/v1/IntegrationDesigntimeArtifacts')
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
                .withHeader('Accept', 'application/json')
        def response = HttpResponse.response()
                .withStatusCode(500)
        mockServerClient.when(request).respond(response)

        when:
        designTimeArtifact.upload('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Upload design time artifact call failed with response code = 500'
    }

    def 'Successful update'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('PUT')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')")
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
                .withHeader('Accept', 'application/json')
        def response = HttpResponse.response()
                .withStatusCode(200)
        mockServerClient.when(request).respond(response)

        when:
        designTimeArtifact.update('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'Failure during update'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('PUT')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')")
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
                .withHeader('Accept', 'application/json')
        def response = HttpResponse.response()
                .withStatusCode(500)
        mockServerClient.when(request).respond(response)

        when:
        designTimeArtifact.update('dummy', 'FlashPipe_IFlow', 'dummy', 'dummy', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Update design time artifact call failed with response code = 500'
    }


    def 'Successful delete'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('DELETE')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')")
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
        def response = HttpResponse.response()
                .withStatusCode(200)
        mockServerClient.when(request).respond(response)

        when:
        designTimeArtifact.delete('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'Failure during delete'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('DELETE')
                .withPath("/api/v1/IntegrationDesigntimeArtifacts(Id='FlashPipe_IFlow',Version='active')")
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
        def response = HttpResponse.response()
                .withStatusCode(500)
        mockServerClient.when(request).respond(response)

        when:
        designTimeArtifact.delete('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Delete design time artifact call failed with response code = 500'
    }

    def 'Successful deployment'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('POST')
                .withPath("/api/v1/DeployIntegrationDesigntimeArtifact")
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
                .withHeader('Accept', 'application/json')
                .withQueryStringParameter('Id', "'FlashPipe_IFlow'")
                .withQueryStringParameter('Version', "'active'")
        def response = HttpResponse.response()
                .withStatusCode(202)
        mockServerClient.when(request).respond(response)

        when:
        designTimeArtifact.deploy('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'Failure during deployment'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

        and:
        MockServerClient mockServerClient = new MockServerClient(LOCALHOST, 9443)
        setupCSRFTokenExpectation(mockServerClient)

        def request = HttpRequest.request()
                .withMethod('POST')
                .withPath("/api/v1/DeployIntegrationDesigntimeArtifact")
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
                .withHeader('Accept', 'application/json')
                .withQueryStringParameter('Id', "'FlashPipe_IFlow'")
                .withQueryStringParameter('Version', "'active'")
        def response = HttpResponse.response()
                .withStatusCode(500)
        mockServerClient.when(request).respond(response)

        when:
        designTimeArtifact.deploy('FlashPipe_IFlow', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Deploy design time artifact call failed with response code = 500'
    }

    def 'JSON payload generation'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

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