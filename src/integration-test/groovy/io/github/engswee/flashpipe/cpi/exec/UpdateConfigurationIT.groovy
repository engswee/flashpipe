package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class UpdateConfigurationIT extends Specification {

    @Shared
    IntegrationTestHelper testHelper

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        testHelper = new IntegrationTestHelper(httpExecuter)
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Update')
    }

    def 'No updates required'() {
        given:
        UpdateConfiguration updateConfiguration = new UpdateConfiguration()
        updateConfiguration.setiFlowId('FlashPipe_Update')
        updateConfiguration.setParamFilePath('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update/src/main/resources/parameters.prop')

        when:
        updateConfiguration.execute()

        then:
        noExceptionThrown()
    }

    def 'Configuration updated'() {
        given:
        UpdateConfiguration updateConfiguration = new UpdateConfiguration()
        updateConfiguration.setiFlowId('FlashPipe_Update')
        updateConfiguration.setParamFilePath('src/integration-test/resources/test-data/Configuration/parameters-update.prop')

        when:
        updateConfiguration.execute()

        then:
        noExceptionThrown()
    }
}