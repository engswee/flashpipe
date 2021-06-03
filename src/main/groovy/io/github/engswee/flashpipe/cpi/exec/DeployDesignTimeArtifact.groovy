package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact

import java.util.concurrent.TimeUnit

class DeployDesignTimeArtifact extends APIExecuter {

    static void main(String[] args) {
        DeployDesignTimeArtifact deployDesignTimeArtifact = new DeployDesignTimeArtifact()
        deployDesignTimeArtifact.execute()
    }

    @Override
    void execute() {
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        int delayLength = (System.getenv('DELAY_LENGTH') ?: 30) as int
        int maxCheckLimit = (System.getenv('MAX_CHECK_LIMIT') ?: 10) as int

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        RuntimeArtifact runtimeArtifact = new RuntimeArtifact(this.httpExecuter)

        // Compare designtime version with runtime version to determine if deployment is needed
        println "[INFO] Comparing designtime version with runtime version"
        def designtimeVersion = designTimeArtifact.getVersion(iFlowId, 'active', false)
        def runtimeVersion = runtimeArtifact.getVersion(iFlowId)

        if (runtimeVersion == designtimeVersion) {
            println "[INFO] IFlow ${iFlowId} with version ${runtimeVersion} already deployed. Skipping runtime deployment"
        } else {
            CSRFToken csrfToken = this.oauthTokenHost ? null : new CSRFToken(this.httpExecuter)
            println "[INFO] 🚀 Deploying IFlow ${iFlowId} with version ${designtimeVersion}"
            designTimeArtifact.deploy(iFlowId, csrfToken)

            println "[INFO] Checking deployment status every ${delayLength} seconds up to ${maxCheckLimit} times"
            Boolean deploymentComplete = false
            int checkCounter = 0
            while (!deploymentComplete) {
                TimeUnit.SECONDS.sleep(delayLength)
                def status = runtimeArtifact.getStatus(iFlowId)
                println "[INFO] Current IFlow status = ${status}"
                if (status != 'STARTING') {
                    deploymentComplete = true
                    if (status != 'STARTED') {
                        def errorMessage = runtimeArtifact.getErrorInfo(iFlowId)
                        println "[ERROR] IFlow deployment unsuccessful, ended with status ${status}"
                        println "[ERROR] Error message = ${errorMessage}"
                        System.exit(1)
                    }
                }
                checkCounter++
                if (checkCounter == maxCheckLimit && status != 'STARTED') {
                    println "[ERROR] IFlow status remained in ${status} after ${maxCheckLimit} checks"
                    System.exit(1)
                }
            }
        }
    }
}