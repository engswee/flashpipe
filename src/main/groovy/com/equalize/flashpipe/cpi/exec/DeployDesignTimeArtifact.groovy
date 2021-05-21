package com.equalize.flashpipe.cpi.exec

import com.equalize.flashpipe.cpi.api.DesignTimeArtifact
import com.equalize.flashpipe.cpi.api.RuntimeArtifact

import java.util.concurrent.TimeUnit

if (args.length < 4) {
    println "Enter arguments in the format: <iflow_id> <cpi_host> <user> <password> <delay_length> <max_check_limit>"
    System.exit(1)
}

def iFlowId = args[0]
def host_tmn = args[1]
def user = args[2]
def pw = args[3]
int delayLength = (args.length > 4) ? args[4] as int : 30
int maxCheckLimit = (args.length > 5) ? args[5] as int : 10

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact('https', host_tmn, 443, user, pw)
// TODO - before deploying, check timestamp of latest design time, and also runtime - do not deploy if latest design time already deployed

designTimeArtifact.deploy(iFlowId)

println "[INFO] Checking deployment status every ${delayLength} seconds up to ${maxCheckLimit} times"
RuntimeArtifact runtimeArtifact = new RuntimeArtifact('https', host_tmn, 443, user, pw)
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