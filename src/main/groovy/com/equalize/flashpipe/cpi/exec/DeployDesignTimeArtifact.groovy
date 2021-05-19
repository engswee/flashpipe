package com.equalize.flashpipe.cpi.exec

import com.equalize.flashpipe.cpi.api.DesignTimeArtifact
import com.equalize.flashpipe.cpi.api.RuntimeArtifact

import java.util.concurrent.TimeUnit

if (args.length < 4) {
    println "Enter arguments in the format: <iflow_id> <cpi_host> <user> <password>"
    System.exit(1)
}

def iFlowId = args[0]
def host_tmn = args[1]
def user = args[2]
def pw = args[3]

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact('https', host_tmn, 443, user, pw)
designTimeArtifact.deploy(iFlowId)

println '[INFO] Checking deployment status'
RuntimeArtifact runtimeArtifact = new RuntimeArtifact('https', host_tmn, 443, user, pw)
Boolean deploymentComplete = false
while (!deploymentComplete) {
    TimeUnit.SECONDS.sleep(30)
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
}