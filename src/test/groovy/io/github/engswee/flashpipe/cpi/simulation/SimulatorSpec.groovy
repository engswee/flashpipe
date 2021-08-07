package io.github.engswee.flashpipe.cpi.simulation

import groovy.json.JsonOutput
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

class SimulatorSpec extends Specification {

    @Shared
    ClientAndServer mockServer
    @Shared
    HTTPExecuter httpExecuter

    final static String LOCALHOST = 'localhost'
    MockServerClient mockServerClient

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
        httpExecuter = HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'dummy', 'dummy')
    }

    def setup() {
        mockServerClient = new MockServerClient(LOCALHOST, 9443)
        mockServer.reset()
    }

    def cleanupSpec() {
        mockServer.stop()
    }

    def 'Test token in GET CSRF Token call'() {
        given: 'Token provided in header of successful call'
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath('/itspaces/api/1.0/workspace')
                .withHeader('x-csrf-token', 'fetch')
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withHeader('x-csrf-token', '50B5187CDE58A345C8A713959F9A4893')
        mockServerClient.when(httpRequest).respond(httpResponse)

        and: 'Simulator is instantiated with HTTPExecuter'
        Simulator simulator = new Simulator(httpExecuter)

        when: 'CSRF Token is retrieved'
        def token = simulator.getCSRFToken()

        then: 'Token returned is same as header'
        token == '50B5187CDE58A345C8A713959F9A4893'
    }

    def 'GET CSRF Token call failed'() {
        given:
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath('/itspaces/api/1.0/workspace')
                .withHeader('x-csrf-token', 'fetch')
        def httpResponse = HttpResponse.response()
                .withStatusCode(403)
        mockServerClient.when(httpRequest).respond(httpResponse)

        and:
        Simulator simulator = new Simulator(httpExecuter)

        when:
        simulator.getCSRFToken()

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get CSRF Token call failed with response code = 403'
    }

    def 'Test iFlow and package GUID in GET iFlow Artifact call'() {
        given: 'JSON response is provided in body of successful call'
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/odata/1.0/workspace.svc/Artifacts(Name='JSONToJSONTransformation',Type='IFlow')")
                .withQueryStringParameter('$expand', 'ContentPackages')
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withBody(this.getClass().getResource('/test-data/Simulation/iFlowIDResponse.json').getText('UTF-8'))
        mockServerClient.when(httpRequest).respond(httpResponse)

        and: 'Simulator is instantiated with HTTPExecuter'
        Simulator simulator = new Simulator(httpExecuter)

        when: 'iFlow and package GUID are retrieved'
        Map ids = simulator.getIFlowGuid('JSONToJSONTransformation')

        then: 'IDs returned correctly from JSON body'
        verifyAll {
            ids.get('iFlowGuid') == '19030a788e3f4efd94beea9217d6804a'
            ids.get('packageGuid') == '6dbcfdfc969749f581bb5ee89b15f1a2'
        }
    }

    def 'GET IFlow GUID call failed'() {
        given:
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/odata/1.0/workspace.svc/Artifacts(Name='JSONToJSONTransformation',Type='IFlow')")
                .withQueryStringParameter('$expand', 'ContentPackages')
        def httpResponse = HttpResponse.response()
                .withStatusCode(403)
        mockServerClient.when(httpRequest).respond(httpResponse)

        and:
        Simulator simulator = new Simulator(httpExecuter)

        when:
        simulator.getIFlowGuid('JSONToJSONTransformation')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get IFlow GUID call failed with response code = 403'
    }

    def 'Test iFlow Model in GET iFlow Model call'() {
        given: 'JSON response is provided in body of successful call'
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/api/1.0/workspace/6dbcfdfc969749f581bb5ee89b15f1a2/artifacts/19030a788e3f4efd94beea9217d6804a/entities/19030a788e3f4efd94beea9217d6804a/iflows/dummy")
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withBody(this.getClass().getResource('/test-data/Simulation/iFlowModelResponse.json').getText('UTF-8'))
        mockServerClient.when(httpRequest).respond(httpResponse)

        and: 'Simulator is instantiated with HTTPExecuter'
        Simulator simulator = new Simulator(httpExecuter)

        when: 'iFlow ID is retrieved'
        def model = simulator.getIFlowModel('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a')

        then: 'Model data is parsed correctly from JSON body'
        model == new JsonSlurper().parse(this.getClass().getResource('/test-data/Simulation/iFlowModelResponse.json'))
    }

    def 'GET IFlow Model call failed'() {
        given:
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/api/1.0/workspace/6dbcfdfc969749f581bb5ee89b15f1a2/artifacts/19030a788e3f4efd94beea9217d6804a/entities/19030a788e3f4efd94beea9217d6804a/iflows/dummy")
        def httpResponse = HttpResponse.response()
                .withStatusCode(403)
        mockServerClient.when(httpRequest).respond(httpResponse)

        and:
        Simulator simulator = new Simulator(httpExecuter)

        when:
        simulator.getIFlowModel('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get IFlow Model call failed with response code = 403'
    }

    def 'Test task ID in SIMULATE submit call'() {
        given: 'JSON response is provided in body of successful call'
        // SIMULATE method not supported by Mock Server
        HTTPExecuter stubbedHTTPExecuter = Stub(HTTPExecuter)
        stubbedHTTPExecuter.getResponseCode() >> 200
        stubbedHTTPExecuter.getResponseBody() >> this.getClass().getResource('/test-data/Simulation/submitSimulationResponse.json').newInputStream()

        and: 'Simulator is instantiated with stubbed HTTPExecuter'
        Simulator simulator = new Simulator(stubbedHTTPExecuter)

        when: 'Task ID is retrieved'
        def taskId = simulator.submitSimulationRequest('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a', '50B5187CDE58A345C8A713959F9A4893', 'dummy')

        then: 'ID returned correctly from JSON body'
        taskId == 'f52b4c67-befb-41df-82d7-89d7771bfbb5'
    }

    def 'Submit simulation request call failed'() {
        given:
        HTTPExecuter stubbedHTTPExecuter = Stub(HTTPExecuter)
        stubbedHTTPExecuter.getResponseCode() >> 403
        stubbedHTTPExecuter.logError('Submit Simulation Request') >> { throw new HTTPExecuterException('Submit Simulation Request call failed with response code = 403') }

        and:
        Simulator simulator = new Simulator(stubbedHTTPExecuter)

        when:
        simulator.submitSimulationRequest('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a', '50B5187CDE58A345C8A713959F9A4893', 'dummy')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Submit Simulation Request call failed with response code = 403'
    }

    def 'Test incomplete execution in GET simulation results call'() {
        given: 'JSON response is provided in body of successful call'
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/api/1.0/workspace/6dbcfdfc969749f581bb5ee89b15f1a2/artifacts/19030a788e3f4efd94beea9217d6804a/entities/19030a788e3f4efd94beea9217d6804a/iflows/dummy/simulations/f52b4c67-befb-41df-82d7-89d7771bfbb5")
                .withQueryStringParameter('id', 'dummy')
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withBody(this.getClass().getResource('/test-data/Simulation/getTestStartedResponse.json').getText('UTF-8'))
        mockServerClient.when(httpRequest).respond(httpResponse)

        and: 'Simulator is instantiated with HTTPExecuter'
        Simulator simulator = new Simulator(httpExecuter)

        when: 'Simulation result is queried'
        def result = simulator.querySimulationResult('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a', 'f52b4c67-befb-41df-82d7-89d7771bfbb5', 'SequenceFlow_6')

        then: 'Percentage is returned correctly from JSON response'
        result == 10
    }

    def 'Test completed execution in GET simulation results call'() {
        given: 'JSON response is provided in body of successful call'
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/api/1.0/workspace/6dbcfdfc969749f581bb5ee89b15f1a2/artifacts/19030a788e3f4efd94beea9217d6804a/entities/19030a788e3f4efd94beea9217d6804a/iflows/dummy/simulations/f52b4c67-befb-41df-82d7-89d7771bfbb5")
                .withQueryStringParameter('id', 'dummy')
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withBody(this.getClass().getResource('/test-data/Simulation/getTestSuccessResponse.json').getText('UTF-8'))
        mockServerClient.when(httpRequest).respond(httpResponse)

        and: 'Simulator is instantiated with HTTPExecuter'
        Simulator simulator = new Simulator(httpExecuter)

        when: 'Simulation result is queried'
        Map result = simulator.querySimulationResult('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a', 'f52b4c67-befb-41df-82d7-89d7771bfbb5', 'SequenceFlow_6')

        then: 'Payload is returned correctly from JSON response'
        result.body == this.getClass().getResource('/test-data/Simulation/simulationOutput.json').getText('UTF-8').normalize()
    }

    def 'Test execution loop for GET simulation results call'() {
        given: 'JSON response is provided in body of successful call'
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/api/1.0/workspace/6dbcfdfc969749f581bb5ee89b15f1a2/artifacts/19030a788e3f4efd94beea9217d6804a/entities/19030a788e3f4efd94beea9217d6804a/iflows/dummy/simulations/f52b4c67-befb-41df-82d7-89d7771bfbb5")
                .withQueryStringParameter('id', 'dummy')
        def httpResponse = HttpResponse.response()
                .withStatusCode(200)
                .withBody(this.getClass().getResource('/test-data/Simulation/getTestSuccessResponse.json').getText('UTF-8'))
        mockServerClient.when(httpRequest).respond(httpResponse)

        and: 'Simulator is instantiated with HTTPExecuter'
        Simulator simulator = new Simulator(httpExecuter)

        when: 'Simulation output is retrieved'
        Map result = simulator.getSimulationOutput('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a', 'f52b4c67-befb-41df-82d7-89d7771bfbb5', 'SequenceFlow_6', 1)

        then: 'Simulation result is same as expected JSON payload'
        result.body == this.getClass().getResource('/test-data/Simulation/simulationOutput.json').getText('UTF-8').normalize()
    }

    def 'Query simulation result call failed'() {
        given:
        def httpRequest = HttpRequest.request()
                .withMethod('GET')
                .withPath("/itspaces/api/1.0/workspace/6dbcfdfc969749f581bb5ee89b15f1a2/artifacts/19030a788e3f4efd94beea9217d6804a/entities/19030a788e3f4efd94beea9217d6804a/iflows/dummy/simulations/f52b4c67-befb-41df-82d7-89d7771bfbb5")
                .withQueryStringParameter('id', 'dummy')
        def httpResponse = HttpResponse.response()
                .withStatusCode(403)
        mockServerClient.when(httpRequest).respond(httpResponse)

        and:
        Simulator simulator = new Simulator(httpExecuter)

        when:
        simulator.querySimulationResult('6dbcfdfc969749f581bb5ee89b15f1a2', '19030a788e3f4efd94beea9217d6804a', 'f52b4c67-befb-41df-82d7-89d7771bfbb5', 'SequenceFlow_6')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Query Simulation Result call failed with response code = 403'
    }

    def 'Validation generateSimulationInput'() {
        given:
        Simulator simulator = new Simulator(httpExecuter)

        when:
        def inputBody = this.getClass().getResource('/test-data/Simulation/orderInputBody.json').getBytes()
        def iFlowModel = new JsonSlurper().parse(this.getClass().getResource('/test-data/Simulation/iFlowModelResponse.json'))
        def simulationInput = simulator.generateSimulationInput('SequenceFlow_3', 'SequenceFlow_6', 'Process_1', inputBody, iFlowModel, [:], [:])

        then:
        JsonOutput.prettyPrint(simulationInput) == JsonOutput.prettyPrint(this.getClass().getResource('/test-data/Simulation/simulationInput.json').getText('UTF-8'))
    }
}