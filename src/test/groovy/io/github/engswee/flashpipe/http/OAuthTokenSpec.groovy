package io.github.engswee.flashpipe.http

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.util.MockExpectation
import org.mockserver.integration.ClientAndServer
import spock.lang.Shared
import spock.lang.Specification

class OAuthTokenSpec extends Specification {

    @Shared
    ClientAndServer mockServer

    OAuthToken oAuthToken
    MockExpectation mockExpectation

    final static String LOCALHOST = 'localhost'

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
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

    def 'Get OAuth token'() {
        given:
        this.mockExpectation.set('POST', '/oauth/token', [:], ['grant_type': 'client_credentials'], 200, '{"access_token": "token1"}', [:])

        when:
        String token = OAuthToken.get('http', LOCALHOST, 9443, 'dummy', 'dummy', '')

        then:
        token == 'token1'
    }
    

    def 'Get OAuth token with Neo path'() {
        given:
        this.mockExpectation.set('POST', '/oauth2/api/v1/token', [:], ['grant_type': 'client_credentials'], 200, '{"access_token": "token1"}', [:])

        when:
        String token = OAuthToken.get('http', LOCALHOST, 9443, 'dummy', 'dummy', '/oauth2/api/v1/token')

        then:
        token == 'token1'
    }
    
    def 'Failure during get token'() {
        given:
        this.mockExpectation.set('POST', '/oauth/token', [:], ['grant_type': 'client_credentials'], 500, '', [:])

        when:
        OAuthToken.get('http', LOCALHOST, 9443, 'dummy', 'dummy', '')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get OAuth token call failed with response code = 500'
    }
}