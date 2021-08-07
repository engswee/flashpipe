package io.github.engswee.flashpipe.cpi.simulation

import groovy.json.JsonBuilder
import groovy.json.JsonException
import groovy.json.JsonOutput
import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterException
import org.slf4j.Logger
import org.slf4j.LoggerFactory

import java.util.concurrent.TimeUnit

class Simulator {
    final HTTPExecuter httpExecuter

    static Logger logger = LoggerFactory.getLogger(Simulator)

    Simulator(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
    }

    String generateSimulationInput(String startPoint, String endPoint, String processName, byte[] inputBody, Object iFlowModel, Map inputHeaders, Map inputProperties) {
        Map mock = [:]
        Map trace = [:]
        JsonBuilder builder = new JsonBuilder()
        builder {
            'startSeqID' startPoint
            'endSeqID' endPoint
            'process' processName
            'inputPayload' {
                'headers' inputHeaders
                'properties' inputProperties
                'body' inputBody
            }
            'mockPayloads' mock
            'traceCache' trace
            'traceStartSeqId' ''
            'traceEndSeqId' ''
            'iflowModelTO' iFlowModel
        }
        return builder.toString()
    }

    String getCSRFToken() {
        logger.debug('Get CSRF Token')
        this.httpExecuter.executeRequest('/itspaces/api/1.0/workspace', ['x-csrf-token': 'fetch'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200)
            return this.httpExecuter.getResponseHeader('x-csrf-token')
        else
            this.httpExecuter.logError('Get CSRF Token')
    }

    Map getIFlowGuid(String iFlowId) {
        logger.debug('Get IFlow GUID')
        if (!iFlowId)
            throw new SimulationException('iFlowId is not populated')
        this.httpExecuter.executeRequest("/itspaces/odata/1.0/workspace.svc/Artifacts(Name='${iFlowId}',Type='IFlow')", ['Accept': 'application/json'], ['$expand': 'ContentPackages'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            try {
                def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
                def iFlowGuid = root.d.reg_id
                if (!iFlowGuid)
                    throw new SimulationException('iFlowGuid not found in response')
                def packageGuid = root.d.ContentPackages.results[0].reg_id
                if (!packageGuid)
                    throw new SimulationException('packageGuid not found in response')
                return ['iFlowGuid': iFlowGuid, 'packageGuid': packageGuid]
            } catch (JsonException e) {
                logger.error('🛑 Error running simulation - feature only supported on Neo environment')
                throw new SimulationException("${e.getMessage()}")
            }
        } else {
            this.httpExecuter.logError('Get IFlow GUID')
        }
    }

    Object getIFlowModel(String packageGuid, String iFlowGuid) {
        logger.debug('Get IFlow Model')
        this.httpExecuter.executeRequest("/itspaces/api/1.0/workspace/${packageGuid}/artifacts/${iFlowGuid}/entities/${iFlowGuid}/iflows/dummy", ['Accept': 'application/json'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200)
            return new JsonSlurper().parse(this.httpExecuter.getResponseBody())
        else
            this.httpExecuter.logError('Get IFlow Model')
    }

    String submitSimulationRequest(String packageGuid, String iFlowGuid, String csrfToken, String input) {
        logger.debug('Submit Simulation Request')
        this.httpExecuter.executeRequest('SIMULATE', "/itspaces/api/1.0/workspace/${packageGuid}/artifacts/${iFlowGuid}/entities/${iFlowGuid}/iflows/dummy/simulations", ['x-csrf-token': csrfToken], ['id': 'dummy'], input, 'UTF-8', 'application/json')
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            return root.stepTestTaskId
        } else {
            this.httpExecuter.logError('Submit Simulation Request')
        }
    }

    Object querySimulationResult(String packageGuid, String iFlowGuid, String taskId, String endPoint) {
        logger.debug('Query Simulation Result')
        this.httpExecuter.executeRequest("/itspaces/api/1.0/workspace/${packageGuid}/artifacts/${iFlowGuid}/entities/${iFlowGuid}/iflows/dummy/simulations/${taskId}", null, ['id': 'dummy'])
        def code = this.httpExecuter.getResponseCode()
        if (code == 200) {
            def root = new JsonSlurper().parse(this.httpExecuter.getResponseBody())
            logger.debug("Response body - ${JsonOutput.toJson(root)}")
            int percentage = root.percentageComplete as int
            if (percentage == 100) {
                if (root.statusCode == 'TEST_EXECUTION_FAILED')
                    throw new HTTPExecuterException("🛑 Simulation failed. Error message = ${root.statusMessage}")
                return root.traceData.get(endPoint).tracePages.'1' // Map containing headers, properties and body
            } else
                return percentage
        } else
            this.httpExecuter.logError('Query Simulation Result')
    }

    Map getSimulationOutput(String packageGuid, String iFlowGuid, String taskId, String endPoint, long delay) {
        Boolean complete
        while (!complete) { // TODO - Switch to do - while loop in Groovy 3.x
            TimeUnit.SECONDS.sleep(delay)
            def result = this.querySimulationResult(packageGuid, iFlowGuid, taskId, endPoint)
            switch (result) {
                case Integer:
                    complete = false
                    break
                default:
                    logger.debug("Simulation Message Body Output = ${new String(result.body as byte[], 'UTF-8')}")
                    return result
            }
        }
    }

    void undeployTestIFlow(String packageGuid, String iFlowGuid, String csrfToken) {
        logger.debug('Undeploy Test IFlow')
        this.httpExecuter.executeRequest('CLEAN', "/itspaces/api/1.0/workspace/${packageGuid}/artifacts/${iFlowGuid}/entities/${iFlowGuid}/iflows/dummy/simulations", ['x-csrf-token': csrfToken], ['id': 'dummy'], '{}', 'UTF-8', 'application/json')
        logger.debug(this.httpExecuter.getResponseBody().getText('UTF-8'))
        if (this.httpExecuter.getResponseCode() == 200)
            logger.debug('Test IFlow undeployed')
    }

    Map simulate(byte[] inputBody, String iFlowId, String startPoint, String endPoint, String processName, Map inputHeaders, Map inputProperties) {
        logger.info('🚀 IFlow Simulation Begin')
        def token = getCSRFToken()

        Map ids = getIFlowGuid(iFlowId)
        String iFlowGuid = ids.get('iFlowGuid')
        String packageGuid = ids.get('packageGuid')

        def iFlowModel = getIFlowModel(packageGuid, iFlowGuid)

        String input = generateSimulationInput(startPoint, endPoint, processName, inputBody, iFlowModel, inputHeaders, inputProperties)
        logger.info('Submitting simulation input')
        def taskId = submitSimulationRequest(packageGuid, iFlowGuid, token, input)
        logger.info('Collecting simulation results')
        Map outputMessage = getSimulationOutput(packageGuid, iFlowGuid, taskId, endPoint, 5)

        undeployTestIFlow(packageGuid, iFlowGuid, token)
        logger.info('🏆 IFlow Simulation End')
        return outputMessage
    }
}