package io.github.engswee.flashpipe.cpi.api

import io.github.engswee.flashpipe.cpi.util.MockExpectation
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.mockserver.integration.ClientAndServer
import spock.lang.Shared
import spock.lang.Specification

class IntegrationPackageSpec extends Specification {

    @Shared
    ClientAndServer mockServer
    @Shared
    HTTPExecuter httpExecuter
    @Shared
    IntegrationPackage integrationPackage
    @Shared
    CSRFToken csrfToken
    
    MockExpectation mockExpectation

    final static String LOCALHOST = 'localhost'

    def setupSpec() {
        mockServer = ClientAndServer.startClientAndServer(9443)
        httpExecuter = HTTPExecuterApacheImpl.newInstance('http', LOCALHOST, 9443, 'dummy', 'dummy')
        integrationPackage = new IntegrationPackage(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
    }

    def setup() {
        this.mockExpectation = MockExpectation.newInstance(LOCALHOST, 9443)
    }

    def cleanup() {
        mockServer.reset()
    }

    def cleanupSpec() {
        mockServer.stop()
    }

    def 'Check IFlow in draft version'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')/IntegrationDesigntimeArtifacts", 200, '{"d":{"results":[{"Id":"IFlow1","Version":"Active"}]}}')

        when:
        def iFlowinDraft = integrationPackage.iFlowInDraftVersion('FlashPipeUnitTest', 'IFlow1')

        then:
        iFlowinDraft == true
    }

    def 'Check IFlow not in draft version'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')/IntegrationDesigntimeArtifacts", 200, '{"d":{"results":[{"Id":"IFlow1","Version":"1.0.4"}]}}')

        when:
        def iFlowinDraft = integrationPackage.iFlowInDraftVersion('FlashPipeUnitTest', 'IFlow1')

        then:
        iFlowinDraft == false
    }

    def 'IFlow not in package'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')/IntegrationDesigntimeArtifacts", 200, '{"d":{"results":[{"Id":"IFlow1","Version":"1.0.4"}]}}')

        when:
        integrationPackage.iFlowInDraftVersion('FlashPipeUnitTest', 'IFlow2')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'IFlow IFlow2 not found in package FlashPipeUnitTest'
    }

    def 'Failure during IntegrationPackages call'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')/IntegrationDesigntimeArtifacts", 500, '')

        when:
        integrationPackage.iFlowInDraftVersion('FlashPipeUnitTest', 'IFlow1')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get artifacts of IntegrationPackages call failed with response code = 500'
    }

    def 'Check draft IFlows list in package'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')/IntegrationDesigntimeArtifacts", 200, '{"d":{"results":[{"Id":"IFlow1","Name":"IFlow 1","Version":"Active"}]}}')

        when:
        List iFlows = integrationPackage.getIFlowsWithDraftState('FlashPipeUnitTest')

        then:
        verifyAll {
            iFlows.size() == 1
            iFlows[0].id == 'IFlow1'
            iFlows[0].name == 'IFlow 1'
            iFlows[0].isDraft == true
        }
    }

    def 'Failure during get draft IFlows IntegrationPackages call'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')/IntegrationDesigntimeArtifacts", 500, '')

        when:
        integrationPackage.getIFlowsWithDraftState('FlashPipeUnitTest')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get designtime artifacts of IntegrationPackages call failed with response code = 500'
    }

    def 'Package existence check - exists'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')", 200, '{"d":{"Id":"FlashPipeUnitTest","Name":"FlashPipe Unit Test","Version":"1.0.0"}}')

        when:
        def packageExists = integrationPackage.exists('FlashPipeUnitTest')

        then:
        packageExists == true
    }

    def 'Package existence check - does not exist'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')", 404, '{"error":{"message":{"value":"Requested entity could not be found."}}}')

        when:
        def packageExists = integrationPackage.exists('FlashPipeUnitTest')

        then:
        packageExists == false
    }

    def 'Failure during package existence check'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')", 500, '')

        when:
        integrationPackage.exists('FlashPipeUnitTest')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get IntegrationPackages by ID call failed with response code = 500'
    }

    def 'Successful package creation'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('POST', "/api/v1/IntegrationPackages", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], 201, 'Success')

        when:
        def response = integrationPackage.create('FlashPipeUnitTest', 'FlashPipe Unit Test', csrfToken)

        then:
        response == 'Success'
    }

    def 'Failure during package creation'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('POST', "/api/v1/IntegrationPackages", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893', 'Accept': 'application/json'], 500, '')

        when:
        integrationPackage.create('FlashPipeUnitTest', 'FlashPipe Unit Test', csrfToken)

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Create integration package call failed with response code = 500'
    }

    def 'Successful package deletion'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('DELETE', "/api/v1/IntegrationPackages('FlashPipeUnitTest')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 202, 'Success')

        when:
        integrationPackage.delete('FlashPipeUnitTest', csrfToken)

        then:
        noExceptionThrown()
    }

    def 'Failure during package deletion'() {
        given:
        this.mockExpectation.setCSRFTokenExpectation('/api/v1/', '50B5187CDE58A345C8A713959F9A4893')
        this.mockExpectation.set('DELETE', "/api/v1/IntegrationPackages('FlashPipeUnitTest')", ['x-csrf-token': '50B5187CDE58A345C8A713959F9A4893'], 500, '')

        when:
        integrationPackage.delete('FlashPipeUnitTest', csrfToken)

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Delete integration package call failed with response code = 500'
    }

    def 'Package read only check'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')", 200, '{"d":{"Id":"FlashPipeUnitTest","Name":"FlashPipe Unit Test","Mode":"READ_ONLY"}}')

        when:
        def packageIsReadOnly = integrationPackage.isReadOnly('FlashPipeUnitTest')

        then:
        packageIsReadOnly == true
    }

    def 'Failure during package read only check'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages('FlashPipeUnitTest')", 500, '')

        when:
        integrationPackage.isReadOnly('FlashPipeUnitTest')

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get IntegrationPackages by ID call failed with response code = 500'
    }

    def 'Get list of integration packages'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages", 200, '{"d":{"results":[{"Id":"FlashPipeUnitTest","Name":"FlashPipe Unit Test","Version":"1.0.0"}]}}')

        when:
        List packages = integrationPackage.getPackagesList()

        then:
        verifyAll {
            packages.size() == 1
            packages[0].Id == 'FlashPipeUnitTest'
            packages[0].Name == 'FlashPipe Unit Test'
            packages[0].Version == '1.0.0'
        }
    }

    def 'Failure during get list of integration packages call'() {
        given:
        this.mockExpectation.set('GET', "/api/v1/IntegrationPackages", 500, '')

        when:
        integrationPackage.getPackagesList()

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Get IntegrationPackages list call failed with response code = 500'
    }
}