package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.util.FileUtility
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Specification

class BPMN2HandlerIT extends Specification {

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
    }

    def 'Script collection references are updated'() {
        given:
        // Setup temp Git Source directory for changes in target directory
        new File('target/BPMN2HandlerIT/TempGitSrcDir/FlashPipe Use Script Collection').mkdirs()
        FileUtility.copyDirectory('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Use Script Collection', 'target/BPMN2HandlerIT/TempGitSrcDir/FlashPipe Use Script Collection')

        BPMN2Handler bpmn2Handler = new BPMN2Handler()
        bpmn2Handler.setScriptCollectionMap('DEV_Common_Scripts=NEW_Common_Scripts')
        bpmn2Handler.setiFlowDir('target/BPMN2HandlerIT/TempGitSrcDir/FlashPipe Use Script Collection')

        when:
        bpmn2Handler.execute()

        then:
        String iflowFileContents = new File('target/BPMN2HandlerIT/TempGitSrcDir/FlashPipe Use Script Collection/src/main/resources/scenarioflows/integrationflow/DEV FlashPipe Use Script Collection.iflw').getText('UTF-8')
        verifyAll {
            iflowFileContents.contains('NEW_Common_Scripts') == true
            iflowFileContents.contains('DEV_Common_Scripts') == false
        }

        cleanup:
        new File('target/BPMN2HandlerIT/TempGitSrcDir/FlashPipe Use Script Collection').deleteDir()
    }
}