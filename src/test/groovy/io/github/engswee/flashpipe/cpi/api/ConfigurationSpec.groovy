package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.cpi.util.MockExpectation
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.mockserver.integration.ClientAndServer
import spock.lang.Shared
import spock.lang.Specification

class ConfigurationSpec extends Specification {

    @Shared
    ClientAndServer mockServer
    @Shared
    HTTPExecuter httpExecuter
    @Shared
    Configuration configuration
    @Shared
    CSRFToken csrfToken

    MockExpectation mockExpectation

    final static String LOCALHOST = 'localhost'

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
        httpExecuter = HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'dummy', 'dummy')
        configuration = new Configuration(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
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

    def 'Get IFlow parameters'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationDesigntimeArtifacts(Id='IFlow1',Version='active')/Configurations", 200, '{"d":{"results":[{"ParameterKey":"Param1","ParameterValue":"Value1","DataType":"xsd:string"}]}}')

        when:
        List parameters = configuration.getParameters('IFlow1', 'active')

        then:
        verifyAll {
            parameters.size() == 1
            parameters[0].ParameterKey == 'Param1'
            parameters[0].ParameterValue == 'Value1'
            parameters[0].DataType == 'xsd:string'
        }
    }

    def 'Failure during get IFlow parameters call'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationDesigntimeArtifacts(Id='IFlow1',Version='active')/Configurations", 500, '')

        when:
        configuration.getParameters('IFlow1', 'active')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get configuration parameters call failed with response code = 500'
    }

    def 'Successful parameter update'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('PUT', "/api/v1/IntegrationDesigntimeArtifacts(Id='IFlow1',Version='active')/\$links/Configurations('Param1')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 202, '')

        when:
        configuration.update('IFlow1', 'active', 'Param1', 'Value2', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Failure during parameter update'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('PUT', "/api/v1/IntegrationDesigntimeArtifacts(Id='IFlow1',Version='active')/\$links/Configurations('Param1')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 500, '')

        when:
        configuration.update('IFlow1', 'active', 'Param1', 'Value2', csrfToken)

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Update configuration parameter Param1 call failed with response code = 500'
    }
}