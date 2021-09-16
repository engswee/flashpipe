package io.github.engswee.flashpipe.cpi.simulation

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.util.TestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.HTTPExecuterException
import spock.lang.Shared
import spock.lang.Specification

class SimulatorIT extends Specification {
    @Shared
    Simulator simulator

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        simulator = new Simulator(httpExecuter)
        new TestHelper(httpExecuter).setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Check_Groovy_Camel_Versions', 'FlashPipe Check Groovy Camel Versions', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Check Groovy Camel Versions')
    }

    def 'Check Groovy and Camel version'() {
        when:
        Map outputMessage = simulator.simulate(''.getBytes('UTF-8'), 'FlashPipe_Check_Groovy_Camel_Versions', 'SequenceFlow_3', 'SequenceFlow_6', 'Process_1', [:], [:])

        then:
        def root = new JsonSlurper().parse(outputMessage.body as byte[], 'UTF-8')
        verifyAll {
            root.versions.groovy == '2.4.21'
            root.versions.camel == '2.24.2'
        }
    }

    def 'Incorrect startPoint triggers exception'() {
        when:
        simulator.simulate(''.getBytes('UTF-8'), 'FlashPipe_Check_Groovy_Camel_Versions', 'dummy', 'SequenceFlow_6', 'Process_1', [:], [:])

        then:
        HTTPExecuterException e = thrown()
        e.getMessage() == 'Submit Simulation Request call failed with response code = 500'
    }

    def 'Incorrect endPoint triggers exception'() {
        when:
        simulator.simulate(''.getBytes('UTF-8'), 'FlashPipe_Check_Groovy_Camel_Versions', 'SequenceFlow_3', 'dummy', 'Process_1', [:], [:])

        then:
        SimulationException e = thrown()
        e.getMessage() == 'Missing stepTestTaskId in simulation response. Please check that startPoint, endPoint and processName are configured correctly'
    }

    def 'Incorrect processName triggers exception'() {
        when:
        simulator.simulate(''.getBytes('UTF-8'), 'FlashPipe_Check_Groovy_Camel_Versions', 'SequenceFlow_3', 'SequenceFlow_6', 'dummy', [:], [:])

        then:
        SimulationException e = thrown()
        e.getMessage() == '🛑 Simulation failed. Error message = Test execution has failed; please try again'
    }
}