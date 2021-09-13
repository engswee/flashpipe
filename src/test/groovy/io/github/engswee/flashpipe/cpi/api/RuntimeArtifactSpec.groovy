package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.cpi.util.MockExpectation
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.mockserver.integration.ClientAndServer
import spock.lang.Shared
import spock.lang.Specification

class RuntimeArtifactSpec extends Specification {

    @Shared
    ClientAndServer mockServer
    @Shared
    HTTPExecuter httpExecuter
    @Shared
    RuntimeArtifact runtimeArtifact
    MockExpectation mockExpectation

    final static String LOCALHOST = 'localhost'

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
        httpExecuter = HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'dummy', 'dummy')
        runtimeArtifact = new RuntimeArtifact(httpExecuter)
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

    def 'Get IFlow status'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", 200, '{"d":{"Id":"IFlow1","Status":"STARTED"}}')

        when:
        def iFlowStatus = runtimeArtifact.getStatus('IFlow1')

        then:
        iFlowStatus == 'STARTED'
    }

    def 'Failure during get IFlow status'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", 500, '')

        when:
        runtimeArtifact.getStatus('IFlow1')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get runtime artifact call failed with response code = 500'
    }

    def 'Get version of STARTED IFlow'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", 200, '{"d":{"Id":"IFlow1","Status":"STARTED", "Version":"1.0.4"}}')

        when:
        def iFlowVersion = runtimeArtifact.getVersion('IFlow1')

        then:
        iFlowVersion == '1.0.4'
    }

    def 'Get version of non STARTED IFlow'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", 200, '{"d":{"Id":"IFlow1","Status":"STARTING", "Version":"1.0.4"}}')

        when:
        def iFlowVersion = runtimeArtifact.getVersion('IFlow1')

        then:
        iFlowVersion == null
    }

    def 'Get version of IFlow not deployed'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", 404, '')

        when:
        def iFlowVersion = runtimeArtifact.getVersion('IFlow1')

        then:
        iFlowVersion == null
    }

    def 'Successful IFlow undeployment'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('DELETE', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 202, '')

        when:
        runtimeArtifact.undeploy('IFlow1', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'IFlow deployment skipped due to runtime artifact not found'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('DELETE', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 404, '')

        when:
        runtimeArtifact.undeploy('IFlow1', new CSRFToken(httpExecuter))

        then:
        noExceptionThrown()
    }

    def 'Failure during IFlow undeployment'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('DELETE', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 500, '')

        when:
        runtimeArtifact.undeploy('IFlow1', new CSRFToken(httpExecuter))

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Undeploy runtime artifact call failed with response code = 500'
    }

    def 'Get IFlow deployment error'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')/ErrorInformation/\$value", 200, '{"parameter":["Error"]}')

        when:
        def iFlowError = runtimeArtifact.getErrorInfo('IFlow1')

        then:
        iFlowError == '[Error]'
    }

    def 'Failure during get IFlow deployment error'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationRuntimeArtifacts('IFlow1')/ErrorInformation/\$value", 500, '')

        when:
        runtimeArtifact.getErrorInfo('IFlow1')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get runtime artifact error information call failed with response code = 500'
    }
}