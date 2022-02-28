package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class UploadDesignTimeArtifactIT extends Specification {

    @Shared
    IntegrationTestHelper testHelper
    @Shared
    DesignTimeArtifact designTimeArtifact
    @Shared
    IntegrationPackage integrationPackage

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        testHelper = new IntegrationTestHelper(httpExecuter)
        designTimeArtifact = new DesignTimeArtifact(httpExecuter)
        integrationPackage = new IntegrationPackage(httpExecuter)
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Upload')
        testHelper.cleanupIFlow('FlashPipe_Use_Script_Collection')
        testHelper.deletePackage('FlashPipeIntegrationTestUpload')
    }

    def 'Upload design time with package creation'() {
        given:
        UploadDesignTimeArtifact uploadDesignTimeArtifact = new UploadDesignTimeArtifact()
        uploadDesignTimeArtifact.setiFlowId('FlashPipe_Upload')
        uploadDesignTimeArtifact.setiFlowName('FlashPipe Upload')
        uploadDesignTimeArtifact.setiFlowDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Upload')
        uploadDesignTimeArtifact.setPackageId('FlashPipeIntegrationTestUpload')
        uploadDesignTimeArtifact.setPackageName('FlashPipe Integration Test Upload')

        when:
        uploadDesignTimeArtifact.execute()

        then:
        verifyAll {
            designTimeArtifact.getVersion('FlashPipe_Upload', 'active', false) == '1.0.0'
            integrationPackage.exists('FlashPipeIntegrationTestUpload') == true
        }
        
        cleanup:
        testHelper.cleanupIFlow('FlashPipe_Upload')
    }

    def 'Upload design time to existing package'() {
        given:
        UploadDesignTimeArtifact uploadDesignTimeArtifact = new UploadDesignTimeArtifact()
        uploadDesignTimeArtifact.setiFlowId('FlashPipe_Upload')
        uploadDesignTimeArtifact.setiFlowName('FlashPipe Upload')
        uploadDesignTimeArtifact.setiFlowDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Upload')
        uploadDesignTimeArtifact.setPackageId('FlashPipeIntegrationTestUpload')
        uploadDesignTimeArtifact.setPackageName('FlashPipe Integration Test Upload')

        when:
        uploadDesignTimeArtifact.execute()

        then:
        designTimeArtifact.getVersion('FlashPipe_Upload', 'active', false) == '1.0.0'
    }

    def 'Upload IFlow with script collection handling'() {
        given:
        UploadDesignTimeArtifact uploadDesignTimeArtifact = new UploadDesignTimeArtifact()
        uploadDesignTimeArtifact.setiFlowId('FlashPipe_Use_Script_Collection')
        uploadDesignTimeArtifact.setiFlowName('FlashPipe Use Script Collection')
        uploadDesignTimeArtifact.setiFlowDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/DEV FlashPipe Use Script Collection')
        uploadDesignTimeArtifact.setPackageId('FlashPipeIntegrationTest')
        uploadDesignTimeArtifact.setPackageName('FlashPipe Integration Test')
        uploadDesignTimeArtifact.setScriptCollectionMap('DEV_Common_Scripts=Common_Scripts')

        when:
        uploadDesignTimeArtifact.execute()

        then:
        designTimeArtifact.getVersion('FlashPipe_Use_Script_Collection', 'active', false) == '1.0.0'
    }
}