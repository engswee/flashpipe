package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken

if (args.length < 6) {
    println "Enter arguments in the format: <iflow_id> <version> <cpi_host> <user> <password> <output_file> <oauth_token_host>"
    System.exit(1)
}

def iFlowId = args[0]
def iFlowVersion = args[1]
def host_tmn = args[2]
def user = args[3]
def pw = args[4]
def outputFile = args[5]
def oauthTokenHost = (args.length > 6) ? args[6] : null

String token = oauthTokenHost ? OAuthToken.get('https', oauthTokenHost, 443, user, pw) : null
HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host_tmn, 443, user, pw, token)
DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)

File outputZip = new File(outputFile)
outputZip.bytes = designTimeArtifact.download(iFlowId, iFlowVersion)
println "[INFO] IFlow downloaded to ${outputFile}"