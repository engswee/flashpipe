package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class DeployDesignTimeArtifactIT extends Specification {

    @Shared
    IntegrationTestHelper testHelper

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        testHelper = new IntegrationTestHelper(httpExecuter)
        // Upload IFlows to design time
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update_Error', 'FlashPipe Update Error', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update Error')
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Update')
        testHelper.cleanupIFlow('FlashPipe_Update_Error')
    }

    def 'Deploy new IFlow'() {
        given:
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update'])
        deployDesignTimeArtifact.setDelayLength(30)
        deployDesignTimeArtifact.setMaxCheckLimit(10)

        when:
        deployDesignTimeArtifact.execute()

        then:
        noExceptionThrown()
    }

    def 'Deployment skipped as same version IFlow already deployed'() {
        given:
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update'])
        deployDesignTimeArtifact.setDelayLength(1)
        deployDesignTimeArtifact.setMaxCheckLimit(1)

        when:
        deployDesignTimeArtifact.execute()

        then:
        noExceptionThrown()
    }

    def 'Exception thrown if deployment unsuccessful'() {
        given: 'Deploy another IFlow with the same endpoint'
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update_Error'])
        deployDesignTimeArtifact.setDelayLength(30)
        deployDesignTimeArtifact.setMaxCheckLimit(10)

        when:
        deployDesignTimeArtifact.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage().reverse().take(88).reverse() == "Http Address '/flashpipe-update' already registered for another iflow 'FlashPipe_Update'"
    }

    def 'Exception thrown if max limit reached'() {
        given:
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.setiFlows(['FlashPipe_Update'])
        deployDesignTimeArtifact.setDelayLength(5)
        deployDesignTimeArtifact.setMaxCheckLimit(1)

        // Undeploy existing runtime artifact
        testHelper.undeployIFlow('FlashPipe_Update')

        when:
        deployDesignTimeArtifact.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Max check limit reached'
    }
}