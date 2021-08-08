package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

import java.util.concurrent.TimeUnit

class RuntimeArtifactBasicAuthIT extends Specification {
    @Shared
    RuntimeArtifact runtimeArtifact
    @Shared
    DesignTimeArtifact designTimeArtifact
    @Shared
    CSRFToken csrfToken

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        runtimeArtifact = new RuntimeArtifact(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
        designTimeArtifact = new DesignTimeArtifact(httpExecuter)
    }

    def 'Undeploy'() {
        when:
        runtimeArtifact.undeploy('FlashPipe_Update', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Deploy'() {
        when:
        designTimeArtifact.deploy('FlashPipe_Update', csrfToken)
        TimeUnit.SECONDS.sleep(5)

        then:
        noExceptionThrown()
    }

    def 'Query'() {
        when:
        runtimeArtifact.getStatus('FlashPipe_Update')

        then:
        noExceptionThrown()
    }
}