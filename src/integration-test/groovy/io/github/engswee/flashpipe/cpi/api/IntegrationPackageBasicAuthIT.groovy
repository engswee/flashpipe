package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.util.TestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class IntegrationPackageBasicAuthIT extends Specification {
    @Shared
    IntegrationPackage integrationPackage
    @Shared
    CSRFToken csrfToken

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        integrationPackage = new IntegrationPackage(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
        new TestHelper(httpExecuter).setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
    }

    def 'Create'() {
        when:
        def responseBody = integrationPackage.create('FlashPipeIntegrationTestCreate', 'FlashPipe Integration Test Create', csrfToken)

        then:
        def root = new JsonSlurper().parseText(responseBody)
        verifyAll {
            root.d.Id == 'FlashPipeIntegrationTestCreate'
            root.d.Name == 'FlashPipe Integration Test Create'
            root.d.ShortText == 'FlashPipeIntegrationTestCreate'
            root.d.Version == '1.0.0'
        }
    }

    def 'Query'() {
        when:
        def exists = integrationPackage.exists('FlashPipeIntegrationTestCreate')

        then:
        exists == true
    }

    def 'Delete'() {
        when:
        integrationPackage.delete('FlashPipeIntegrationTestCreate', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Check Designtime Artifact in draft'() {
        when:
        def draftVersion = integrationPackage.iFlowInDraftVersion('FlashPipeIntegrationTest', 'FlashPipe_Update')

        then:
        draftVersion == false
    }

    def 'Get Designtime Artifacts'() {
        when:
        List flows = integrationPackage.getIFlowsWithDraftState('FlashPipeIntegrationTest')

        then:
        def updateIFlow = flows.find {it.id == 'FlashPipe_Update'}
        verifyAll {
            updateIFlow.name == 'FlashPipe Update'
            updateIFlow.isDraft == false
        }
    }

    def 'Get Packages List'() {
        when:
        List packages = integrationPackage.getPackagesList()

        then:
        packages.any { it.Id == 'FlashPipeIntegrationTest' } == true
    }
}