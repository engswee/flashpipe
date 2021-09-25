package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class ConfigurationBasicAuthIT extends Specification {
    @Shared
    Configuration configuration
    @Shared
    CSRFToken csrfToken
    @Shared
    IntegrationTestHelper testHelper

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        configuration = new Configuration(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
        testHelper = new IntegrationTestHelper(httpExecuter)
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Update')
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