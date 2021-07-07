package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import spock.lang.Shared
import spock.lang.Specification

import java.util.concurrent.TimeUnit

class RuntimeArtifactOAuthIT extends Specification {

    @Shared
    HTTPExecuter httpExecuter
    @Shared
    RuntimeArtifact runtimeArtifact
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
        runtimeArtifact = new RuntimeArtifact(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
    }

    def 'Undeploy'() {
        when:
        runtimeArtifact.undeploy('FlashPipe_Update', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Deploy'() {
        given:
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

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