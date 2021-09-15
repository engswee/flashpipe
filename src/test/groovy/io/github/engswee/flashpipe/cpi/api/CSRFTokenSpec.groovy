package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.cpi.util.MockExpectation
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.mockserver.integration.ClientAndServer
import spock.lang.Shared
import spock.lang.Specification

class CSRFTokenSpec extends Specification {

    @Shared
    ClientAndServer mockServer
    @Shared
    HTTPExecuter httpExecuter

    CSRFToken csrfToken
    MockExpectation mockExpectation

    final static String LOCALHOST = 'localhost'

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
        httpExecuter = HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'dummy', 'dummy')
    }

    def setup() {
        this.mockExpectation = MockExpectation.newInstance(LOCALHOST, 9443)
        this.csrfToken = new CSRFToken(httpExecuter)
    }

    def cleanup() {
        mockServer.reset()
    }

    def cleanupSpec() {
        mockServer.stop()
    }

    def 'Successful get token'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')

        when:
        def token = this.csrfToken.get()

        then:
        token == '50B5187CDE58A345C8A713959F9A4893'
    }

    def 'Failure during get token'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893', 400)

        when:
        this.csrfToken.get()

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get CSRF Token call failed with response code = 400'
    }
}