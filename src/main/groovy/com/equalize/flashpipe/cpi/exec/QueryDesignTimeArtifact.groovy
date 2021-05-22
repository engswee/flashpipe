package com.equalize.flashpipe.cpi.exec

import com.equalize.flashpipe.cpi.api.DesignTimeArtifact
import com.equalize.flashpipe.cpi.api.IntegrationPackage

if (args.length < 5) {
    println "Enter arguments in the format: <iflow_id> <package_id> <cpi_host> <user> <password>"
    System.exit(1)
}

def iFlowId = args[0]
def packageId = args[1]
def host_tmn = args[2]
def user = args[3]
def pw = args[4]

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact('https', host_tmn, 443, user, pw)
if (designTimeArtifact.getVersion(iFlowId, 'active')) {
    println "[INFO] Active version of IFlow ${iFlowId} exists"
//  Check if version is in draft mode
    IntegrationPackage integrationPackage = new IntegrationPackage('https', host_tmn, 443, user, pw)
    if (integrationPackage.iFlowInDraftVersion(packageId, iFlowId)) {
        println "[ERROR] IFlow ${iFlowId} is in Draft state. Save Version of IFlow in Web UI first!"
        System.exit(1)
    }
} else {
    println "[INFO] Active version of IFlow ${iFlowId} does not exist"
    System.exit(99)
}