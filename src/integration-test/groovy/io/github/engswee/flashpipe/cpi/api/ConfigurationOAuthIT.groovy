package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import spock.lang.Shared
import spock.lang.Specification

class ConfigurationOAuthIT extends Specification {

    @Shared
    HTTPExecuter httpExecuter
    @Shared
    Configuration configuration
    @Shared
    CSRFToken csrfToken

    def setupSpec() {
        def host = System.getProperty('cpi.host.tmn')
        def clientid = System.getProperty('cpi.oauth.clientid')
        def clientsecret = System.getProperty('cpi.oauth.clientsecret')
        def oauthHost = System.getProperty('cpi.host.oauth')
        def oauthTokenPath = System.getProperty('cpi.host.oauthpath')
        def token = OAuthToken.get('https', oauthHost, 443, clientid, clientsecret, oauthTokenPath)

        httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, token)
        configuration = new Configuration(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
    }

    def 'Update'() {
        when:
        configuration.update('FlashPipe_Update', 'active', 'Sender Endpoint', '/update_oauth', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Get'() {
        when:
        List parameters = configuration.getParameters('FlashPipe_Update', 'active')

        then:
        parameters.find { it.ParameterKey == 'Sender Endpoint' }.ParameterValue == '/update_oauth'
    }
}