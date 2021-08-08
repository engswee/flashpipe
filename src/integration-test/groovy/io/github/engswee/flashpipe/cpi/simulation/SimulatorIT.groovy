package io.github.engswee.flashpipe.cpi.simulation

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
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
}