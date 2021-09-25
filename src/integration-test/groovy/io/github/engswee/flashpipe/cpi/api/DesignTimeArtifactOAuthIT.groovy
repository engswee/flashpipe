package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import org.zeroturnaround.zip.ZipUtil
import spock.lang.Shared
import spock.lang.Specification

class DesignTimeArtifactOAuthIT extends Specification {
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
        designTimeArtifact = new DesignTimeArtifact(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
        new IntegrationTestHelper(httpExecuter).setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
    }

    def 'Upload'() {
        given:
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Upload'), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        when:
        def responseBody = designTimeArtifact.upload(iFlowContent, 'FlashPipeUpload', 'FlashPipe Upload', 'FlashPipeIntegrationTest', csrfToken)

        then:
        def root = new JsonSlurper().parseText(responseBody)
        verifyAll {
            root.d.Id == 'FlashPipeUpload'
            root.d.Name == 'FlashPipe Upload'
            root.d.PackageId == 'FlashPipeIntegrationTest'
            root.d.Version == '1.0.0'
        }
    }

    def 'Query'() {
        when:
        def iFlowExists = designTimeArtifact.getVersion('FlashPipeUpload', 'active', true)

        then:
        iFlowExists == '1.0.0'
    }

    def 'Download'() {
        when:
        designTimeArtifact.download('FlashPipeUpload', 'active')

        then:
        noExceptionThrown()
    }

    def 'Delete'() {
        given:

        when:
        designTimeArtifact.delete('FlashPipeUpload', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Update'() {
        given:
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update'), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        when:
        designTimeArtifact.update(iFlowContent, 'FlashPipe_Update', 'FlashPipe Update', 'FlashPipeIntegrationTest', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Deploy'() {
        when:
        designTimeArtifact.deploy('FlashPipe_Update', csrfToken)

        then:
        noExceptionThrown()
    }
}