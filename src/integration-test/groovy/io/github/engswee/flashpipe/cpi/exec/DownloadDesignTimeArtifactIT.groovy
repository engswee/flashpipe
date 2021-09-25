package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Specification

class DownloadDesignTimeArtifactIT extends Specification {

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        new IntegrationTestHelper(httpExecuter).setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
    }

    def 'Download of existing IFlow'() {
        given:
        DownloadDesignTimeArtifact downloadDesignTimeArtifact = new DownloadDesignTimeArtifact()
        downloadDesignTimeArtifact.setiFlowId('FlashPipe_Update')
        downloadDesignTimeArtifact.setiFlowVersion('active')
        downloadDesignTimeArtifact.setOutputFile('target/FlashPipe_Update.zip')

        when:
        downloadDesignTimeArtifact.execute()

        then:
        noExceptionThrown()
    }
}