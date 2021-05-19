package com.equalize.flashpipe.cpi.exec

import com.equalize.flashpipe.cpi.api.DesignTimeArtifact
import com.equalize.flashpipe.cpi.api.IntegrationPackage
import org.zeroturnaround.zip.ZipUtil

if (args.length < 8) {
    println "Enter arguments in the format: <iflow_name> <iflow_id> <package_id> <package_name> <iflow_dir> <cpi_host> <user> <password>"
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

// Zip iFlow directory and encode to Base 64
ByteArrayOutputStream baos = new ByteArrayOutputStream()
ZipUtil.pack(new File(iFlowDir), baos)
def iFlowContent = baos.toByteArray().encodeBase64().toString()

IntegrationPackage integrationPackage = new IntegrationPackage('https', host_tmn, 443, user, pw)
if (!integrationPackage.packageExists(packageId)) {
    println "Package ${packageId} does not exist. Creating package..."
    def result = integrationPackage.create(packageId, packageName)
    println result
}

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact('https', host_tmn, 443, user, pw)
def response = designTimeArtifact.upload(iFlowContent, iFlowId, iFlowName, packageId)
println response