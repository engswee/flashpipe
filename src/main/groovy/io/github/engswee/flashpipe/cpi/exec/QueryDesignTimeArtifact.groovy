package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken

if (args.length < 5) {
    println "Enter arguments in the format: <iflow_id> <package_id> <cpi_host> <user> <password> <oauth_token_host>"
    System.exit(1)
}

def iFlowId = args[0]
def packageId = args[1]
def host_tmn = args[2]
def user = args[3]
def pw = args[4]
def oauthTokenHost = (args.length > 5) ? args[5] : null

String token = oauthTokenHost ? OAuthToken.get('https', oauthTokenHost, 443, user, pw) : null
HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host_tmn, 443, user, pw, token)
DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

if (designTimeArtifact.getVersion(iFlowId, 'active', true)) {
    println "[INFO] Active version of IFlow ${iFlowId} exists"
//  Check if version is in draft mode
    IntegrationPackage integrationPackage = new IntegrationPackage(httpExecuter)
    if (integrationPackage.iFlowInDraftVersion(packageId, iFlowId)) {
        println "[ERROR] IFlow ${iFlowId} is in Draft state. Save Version of IFlow in Web UI first!"
        System.exit(1)
    }
} else {
    println "[INFO] Active version of IFlow ${iFlowId} does not exist"
    System.exit(99)
}