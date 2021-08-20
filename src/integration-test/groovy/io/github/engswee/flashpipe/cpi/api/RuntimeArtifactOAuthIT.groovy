package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.cpi.util.TestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import spock.lang.Shared
import spock.lang.Specification

import java.util.concurrent.TimeUnit

class RuntimeArtifactOAuthIT extends Specification {
    @Shared
    RuntimeArtifact runtimeArtifact
    @Shared
    DesignTimeArtifact designTimeArtifact
    @Shared
    CSRFToken csrfToken

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def clientid = System.getenv('OAUTH_CLIENTID')
        def clientsecret = System.getenv('OAUTH_CLIENTSECRET')
        def oauthHost = System.getenv('HOST_OAUTH')
        def oauthTokenPath = System.getenv('HOST_OAUTH_PATH')
        def token = OAuthToken.get('https', oauthHost, 443, clientid, clientsecret, oauthTokenPath)

        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, token)
        runtimeArtifact = new RuntimeArtifact(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
        designTimeArtifact = new DesignTimeArtifact(httpExecuter)
        new TestHelper(httpExecuter).setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
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