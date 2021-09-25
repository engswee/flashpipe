package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Specification

class QueryDesignTimeArtifactIT extends Specification {

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        new IntegrationTestHelper(httpExecuter).setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
    }

    def 'Query of existing IFlow'() {
        given:
        QueryDesignTimeArtifact queryDesignTimeArtifact = new QueryDesignTimeArtifact()
        queryDesignTimeArtifact.setiFlowId('FlashPipe_Update')
        queryDesignTimeArtifact.setPackageId('FlashPipeIntegrationTest')

        when:
        queryDesignTimeArtifact.execute()

        then:
        noExceptionThrown()
    }
    
    def 'Query of non existent IFlow'() {
        given:
        QueryDesignTimeArtifact queryDesignTimeArtifact = new QueryDesignTimeArtifact()
        queryDesignTimeArtifact.setiFlowId('IDoNotExist')
        queryDesignTimeArtifact.setPackageId('FlashPipeIntegrationTest')

        when:
        queryDesignTimeArtifact.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Active version of IFlow does not exist'
    }
}