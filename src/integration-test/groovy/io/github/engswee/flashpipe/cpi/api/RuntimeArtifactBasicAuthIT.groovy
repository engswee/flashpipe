package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class RuntimeArtifactBasicAuthIT extends Specification {

    @Shared
    HTTPExecuter httpExecuter
    @Shared
    RuntimeArtifact runtimeArtifact
    @Shared
    CSRFToken csrfToken

    def setupSpec() {
        def host = System.getProperty('cpi.host.tmn')
        def user = System.getProperty('cpi.basic.userid')
        def password = System.getProperty('cpi.basic.password')
        httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        runtimeArtifact = new RuntimeArtifact(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
    }

    def 'Query'() {
        when:
        runtimeArtifact.getStatus('FlashPipe_Update')

        then:
        noExceptionThrown()
    }
}