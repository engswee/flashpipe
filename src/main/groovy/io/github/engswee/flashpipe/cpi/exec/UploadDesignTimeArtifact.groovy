package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import io.github.engswee.flashpipe.http.OAuthToken
import org.zeroturnaround.zip.ZipUtil

if (args.length < 8) {
    println "Enter arguments in the format: <iflow_name> <iflow_id> <package_id> <package_name> <iflow_dir> <cpi_host> <user> <password> <oauth_token_host>"
    System.exit(1)
}

def iFlowName = args[0]
def iFlowId = args[1]
def packageId = args[2]
def packageName = args[3]
def iFlowDir = args[4]
def host_tmn = args[5]
def user = args[6]
def pw = args[7]
def oauthTokenHost = (args.length > 8) ? args[8] : null

String token = oauthTokenHost ? OAuthToken.get('https', oauthTokenHost, 443, user, pw) : null
HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host_tmn, 443, user, pw, token)
CSRFToken csrfToken = oauthTokenHost ? null : new CSRFToken(httpExecuter)

IntegrationPackage integrationPackage = new IntegrationPackage(httpExecuter)
if (!integrationPackage.packageExists(packageId)) {
    println "[INFO] Package ${packageId} does not exist. Creating package..."
    def result = integrationPackage.create(packageId, packageName, csrfToken)
    println result
}

// Zip iFlow directory and encode to Base 64
ByteArrayOutputStream baos = new ByteArrayOutputStream()
ZipUtil.pack(new File(iFlowDir), baos)
def iFlowContent = baos.toByteArray().encodeBase64().toString()

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)
def response = designTimeArtifact.upload(iFlowContent, iFlowId, iFlowName, packageId, csrfToken)
println response