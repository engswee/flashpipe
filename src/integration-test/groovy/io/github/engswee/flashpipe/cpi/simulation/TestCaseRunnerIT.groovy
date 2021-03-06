package io.github.engswee.flashpipe.cpi.simulation

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import spock.lang.Shared
import spock.lang.Specification
import spock.lang.Unroll

class TestCaseRunnerIT extends Specification {
    @Shared
    TestCaseRunner testCaseRunner
    @Shared
    IntegrationTestHelper testHelper

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        testCaseRunner = new TestCaseRunner(host, user, password)
        testHelper = new IntegrationTestHelper(testCaseRunner.getHttpExecuter())
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Simulation_JSON_Mapping', 'FlashPipe Simulation JSON Mapping', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Simulation JSON Mapping')
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Simulation_XML_Mapping', 'FlashPipe Simulation XML Mapping', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Simulation XML Mapping')
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Simulation_JSON_Mapping')
        testHelper.cleanupIFlow('FlashPipe_Simulation_XML_Mapping')
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

    def 'Exception thrown when output file not available'() {
        when:
        testCaseRunner.run('/test-data/Simulation/TestCase/Exception/TestCase5-Exception.json')

        then:
        SimulationException e = thrown()
        e.getMessage() == 'File at resources directory /test-data/Simulation/TestCase/Exception/output/purchaseOrder.json cannot be found'
    }
}