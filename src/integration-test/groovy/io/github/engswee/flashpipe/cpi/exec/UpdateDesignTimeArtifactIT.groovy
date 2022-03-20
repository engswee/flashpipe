package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class UpdateDesignTimeArtifactIT extends Specification {

    @Shared
    IntegrationTestHelper testHelper
    @Shared
    DesignTimeArtifact designTimeArtifact

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        testHelper = new IntegrationTestHelper(httpExecuter)
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
        testHelper.deployIFlow('FlashPipe_Update', true)
        designTimeArtifact = new DesignTimeArtifact(httpExecuter)
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Update')
    }

    def 'Update IFlow'() {
        given:
        
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.setiFlowId('FlashPipe_Update')
        updateDesignTimeArtifact.setiFlowName('FlashPipe Update')
        updateDesignTimeArtifact.setiFlowDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
        updateDesignTimeArtifact.setPackageId('FlashPipeIntegrationTest')

        when:
        updateDesignTimeArtifact.execute()

        then:
        designTimeArtifact.getVersion('FlashPipe_Update', 'active', false) == '1.0.1'
    }
}