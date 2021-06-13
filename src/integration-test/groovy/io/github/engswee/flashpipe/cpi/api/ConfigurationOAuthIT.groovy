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

    def setupSpec() {
        def host = System.getProperty('cpi.host.tmn')
        def clientid = System.getProperty('cpi.oauth.clientid')
        def clientsecret = System.getProperty('cpi.oauth.clientsecret')
        def oauthHost = System.getProperty('cpi.host.oauth')
        def token = OAuthToken.get('https', oauthHost, 443, clientid, clientsecret)

        httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, token)
        configuration = new Configuration(httpExecuter)
    }

    def 'Update'() {
        when:
        configuration.update('FlashPipe_Update', 'active', 'Sender Endpoint', '/update_oauth', null)

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