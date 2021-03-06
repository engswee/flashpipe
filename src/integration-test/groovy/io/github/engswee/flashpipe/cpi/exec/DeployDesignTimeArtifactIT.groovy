package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class DeployDesignTimeArtifactIT extends Specification {

    @Shared
    IntegrationTestHelper testHelper
    @Shared
    RuntimeArtifact runtimeArtifact

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        testHelper = new IntegrationTestHelper(httpExecuter)
        // Upload IFlows to design time
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update_Error', 'FlashPipe Update Error', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update Error')
        runtimeArtifact = new RuntimeArtifact(httpExecuter)
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Update')
        testHelper.cleanupIFlow('FlashPipe_Update_Error')
    }

    def 'Exception thrown if max limit reached'() {
        given:
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update'])
        deployDesignTimeArtifact.setDelayLength(10)
        deployDesignTimeArtifact.setMaxCheckLimit(1)
        deployDesignTimeArtifact.setCompareVersions(true)

        when:
        deployDesignTimeArtifact.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Max check limit reached'
    }

    def 'Deploy new IFlow'() {
        given:
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update'])
        deployDesignTimeArtifact.setDelayLength(30)
        deployDesignTimeArtifact.setMaxCheckLimit(10)
        deployDesignTimeArtifact.setCompareVersions(true)

        when:
        deployDesignTimeArtifact.execute()

        then:
        verifyAll {
            runtimeArtifact.getVersion('FlashPipe_Update') == '1.0.1'
            runtimeArtifact.getStatus('FlashPipe_Update') == 'STARTED'
        }
    }

    def 'Deployment skipped as same version IFlow already deployed'() {
        given:
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update'])
        deployDesignTimeArtifact.setDelayLength(1)
        deployDesignTimeArtifact.setMaxCheckLimit(1)
        deployDesignTimeArtifact.setCompareVersions(true)

        when:
        deployDesignTimeArtifact.execute()

        then:
        verifyAll {
            runtimeArtifact.getVersion('FlashPipe_Update') == '1.0.1'
            runtimeArtifact.getStatus('FlashPipe_Update') == 'STARTED'
        }
    }

    // TODO - currently not working for Cloud Foundry as it return 204 No content. To uncomment once SAP corrects behavior
//    def 'Exception thrown if deployment unsuccessful'() {
//        given: 'Deploy another IFlow with the same endpoint'
//        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
//        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update_Error'])
//        deployDesignTimeArtifact.setDelayLength(30)
//        deployDesignTimeArtifact.setMaxCheckLimit(10)
//        deployDesignTimeArtifact.setCompareVersions(true)
//
//        when:
//        deployDesignTimeArtifact.execute()
//
//        then:
//        ExecutionException e = thrown()
//        e.getMessage().contains("Http Address '/flashpipe-update' already registered for another iflow 'FlashPipe_Update'") == true
//    }
}