package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken

import java.util.concurrent.TimeUnit

if (args.length < 4) {
    println "Enter arguments in the format: <iflow_id> <cpi_host> <user> <password> <delay_length> <max_check_limit> <oauth_token_host>"
    System.exit(1)
}

def iFlowId = args[0]
def host_tmn = args[1]
def user = args[2]
def pw = args[3]
int delayLength = (args.length > 4) ? args[4] as int : 30
int maxCheckLimit = (args.length > 5) ? args[5] as int : 10
def oauthTokenHost = (args.length > 6) ? args[6] : null

String oauthToken = oauthTokenHost ? OAuthToken.get('https', oauthTokenHost, 443, user, pw) : null
HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host_tmn, 443, user, pw, oauthToken)

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)
RuntimeArtifact runtimeArtifact = new RuntimeArtifact(httpExecuter)

// Compare designtime version with runtime version to determine if deployment is needed
println "[INFO] Comparing designtime version with runtime version"
def designtimeVersion = designTimeArtifact.getVersion(iFlowId, 'active', false)
def runtimeVersion = runtimeArtifact.getVersion(iFlowId)

if (runtimeVersion == designtimeVersion) {
    println "[INFO] IFlow ${iFlowId} with version ${runtimeVersion} already deployed. Skipping runtime deployment"
} else {
    CSRFToken csrfToken = oauthTokenHost ? null: new CSRFToken(httpExecuter)
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