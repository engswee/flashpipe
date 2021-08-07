package io.github.engswee.flashpipe.cpi.simulation

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class TestCaseRunner {
    Map expectedOutputHeaders
    Map expectedOutputProperties
    String expectedOutputBody
    Map actualOutputHeaders
    Map actualOutputProperties
    String actualOutputBody
    final HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(TestCaseRunner)

    TestCaseRunner(String host, String user, String password) {
        this.httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
    }

    void run(String testContentFile) {
        def testContentRoot = new JsonSlurper().parse(this.getClass().getResource(testContentFile))
        // Get configuration of test case from test content file
        String testCaseName = testContentRoot.TestCase.Name
        Map inputHeaders = testContentRoot.TestCase.Input.Headers
        Map inputProperties = testContentRoot.TestCase.Input.Properties
        String inputBodyFilePath = testContentRoot.TestCase.Input.Body
        byte[] inputBody = getResourceContent(inputBodyFilePath, false)
        this.expectedOutputHeaders = testContentRoot.TestCase.ExpectedOutput.Headers
        this.expectedOutputProperties = testContentRoot.TestCase.ExpectedOutput.Properties
        String expectedOutputBodyFilePath = testContentRoot.TestCase.ExpectedOutput.Body
        this.expectedOutputBody = getResourceContent(expectedOutputBodyFilePath, true)

        String iFlowId = testContentRoot.TestCase.IFlowID
        String startPoint = testContentRoot.TestCase.StartPoint
        String endPoint = testContentRoot.TestCase.EndPoint
        String processName = testContentRoot.TestCase.Process

        logger.info("Executing test case - ${testCaseName}")
        Simulator simulator = new Simulator(this.httpExecuter)
        Map outputMessage = simulator.simulate(inputBody, iFlowId, startPoint, endPoint, processName, inputHeaders, inputProperties)

        this.actualOutputBody = new String(outputMessage.body as byte[], 'UTF-8')
        this.actualOutputHeaders = outputMessage.headers as Map
        this.actualOutputProperties = outputMessage.properties as Map

        logger.info('Test case execution completed')
    }

    private Object getResourceContent(String resourcePath, boolean returnText) {
        if (resourcePath) {
            URL resourceURL = this.getClass().getResource(resourcePath)
            if (resourceURL) {
                return (returnText) ? resourceURL.getText('UTF-8') : resourceURL.getBytes()
            } else
                throw new SimulationException("File at resources directory ${resourcePath} cannot be found")
        } else
            return (returnText) ? null : new byte[0]
    }
}