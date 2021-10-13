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

    def 'Update using MANIFEST version'() {
        given:
        
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.setiFlowId('FlashPipe_Update')
        updateDesignTimeArtifact.setiFlowName('FlashPipe Update')
        updateDesignTimeArtifact.setiFlowDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
        updateDesignTimeArtifact.setPackageId('FlashPipeIntegrationTest')
        updateDesignTimeArtifact.setVersionHandling('MANIFEST')

        when:
        updateDesignTimeArtifact.execute()

        then:
        designTimeArtifact.getVersion('FlashPipe_Update', 'active', false) == '1.0.1'
    }

    def 'Update using AUTO_INCREMENT version'() {
        given:
        
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.setiFlowId('FlashPipe_Update')
        updateDesignTimeArtifact.setiFlowName('FlashPipe Update')
        updateDesignTimeArtifact.setiFlowDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
        updateDesignTimeArtifact.setPackageId('FlashPipeIntegrationTest')
        updateDesignTimeArtifact.setVersionHandling('AUTO_INCREMENT')
        updateDesignTimeArtifact.setCurrentiFlowVersion('1.0.1')

        when:
        updateDesignTimeArtifact.execute()

        then:
        designTimeArtifact.getVersion('FlashPipe_Update', 'active', false) == '1.0.2'

        cleanup:
        File manifestFile = new File("src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update/META-INF/MANIFEST.MF")
        def manifestContent = manifestFile.getText('UTF-8')
        def updatedContent = manifestContent.replaceFirst(/Bundle-Version: \S+/, "Bundle-Version: 1.0.1")
        manifestFile.setText(updatedContent, 'UTF-8')
    }

    def 'Exception thrown for invalid VERSION_HANDLING'() {
        given:
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.setVersionHandling('INVALID')

        when:
        updateDesignTimeArtifact.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Invalid entry for VERSION_HANDLING'
    }
}