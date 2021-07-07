package io.github.engswee.flashpipe.cpi.api

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import spock.lang.Shared
import spock.lang.Specification

class IntegrationPackageOAuthIT extends Specification {

    @Shared
    HTTPExecuter httpExecuter
    @Shared
    IntegrationPackage integrationPackage
    @Shared
    CSRFToken csrfToken

    def setupSpec() {
        def host = System.getProperty('cpi.host.tmn')
        def clientid = System.getProperty('cpi.oauth.clientid')
        def clientsecret = System.getProperty('cpi.oauth.clientsecret')
        def oauthHost = System.getProperty('cpi.host.oauth')
        def oauthTokenPath = System.getProperty('cpi.host.oauthpath')
        def token = OAuthToken.get('https', oauthHost, 443, clientid, clientsecret, oauthTokenPath)

        httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, token)
        integrationPackage = new IntegrationPackage(httpExecuter)
        csrfToken = new CSRFToken(httpExecuter)
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
        verifyAll {
            flows.size() == 1
            flows[0].id == 'FlashPipe_Update'
            flows[0].name == 'FlashPipe Update'
            flows[0].isDraft == false
        }
    }
}