package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import spock.lang.Shared
import spock.lang.Specification

class ConfigurationBasicIT extends Specification {

    @Shared
    HTTPExecuter httpExecuter
    @Shared
    Configuration configuration
    @Shared
    CSRFToken csrfToken

    def setupSpec() {
        def host = System.getProperty('cpi.host.tmn')
        def user = System.getProperty('cpi.basic.userid')
        def password = System.getProperty('cpi.basic.password')
        httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        configuration = new Configuration(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
    }

    def 'Update'() {
        when:
        configuration.update('FlashPipe_Update', 'active', 'Sender Endpoint', '/update_basic', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Get'() {
        when:
        List parameters = configuration.getParameters('FlashPipe_Update', 'active')

        then:
        parameters.find { it.ParameterKey == 'Sender Endpoint' }.ParameterValue == '/update_basic'
    }
}