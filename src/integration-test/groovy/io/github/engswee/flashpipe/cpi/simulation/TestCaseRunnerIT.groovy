package io.github.engswee.flashpipe.cpi.simulation

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.util.TestHelper
import spock.lang.Shared
import spock.lang.Specification
import spock.lang.Unroll

class TestCaseRunnerIT extends Specification {
    @Shared
    TestCaseRunner testCaseRunner

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        testCaseRunner = new TestCaseRunner(host, user, password)
        TestHelper testHelper = new TestHelper(testCaseRunner.getHttpExecuter())
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Simulation_JSON_Mapping', 'FlashPipe Simulation JSON Mapping', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Simulation JSON Mapping')
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Simulation_XML_Mapping', 'FlashPipe Simulation XML Mapping', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Simulation XML Mapping')
    }

    @Unroll
    def 'Simulation Test: #testCaseName'() {
        when:
        testCaseRunner.run(TestCaseContentFile)
        Map expectedHeaders = testCaseRunner.getExpectedOutputHeaders()
        Map expectedProperties = testCaseRunner.getExpectedOutputProperties()
        String expectedBody = testCaseRunner.getExpectedOutputBody()

        then:
        verifyAll {
            // Headers
            if (expectedHeaders.size() > 0) {
                expectedHeaders.each { k, v ->
                    assert testCaseRunner.getActualOutputHeaders().get(k) == v
                }
            }
            // Properties
            if (expectedProperties.size() > 0) {
                expectedProperties.each { k, v ->
                    assert testCaseRunner.getActualOutputProperties().get(k) == v
                }
            }
            // Body
            if (expectedBody)
                testCaseRunner.getActualOutputBody() == expectedBody
        }

        where:
        TestCaseContentFile                                                      | _
        '/test-data/Simulation/TestCase/JSONMapping/TestCase1-Body.json'         | _
        '/test-data/Simulation/TestCase/JSONMapping/TestCase2-Property.json'     | _
        '/test-data/Simulation/TestCase/XMLMapping/TestCase3.json'               | _
        '/test-data/Simulation/TestCase/XMLMapping/TestCase4-InputProperty.json' | _

        testCaseName = new JsonSlurper().parse(this.getClass().getResource(TestCaseContentFile)).TestCase.Name
    }
}