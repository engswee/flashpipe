package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import org.zeroturnaround.zip.ZipUtil

if (args.length < 8) {
    println "Enter arguments in the format: <iflow_name> <iflow_id> <package_id> <iflow_dir> <tenant_iflow_version> <cpi_host> <user> <password>"
    System.exit(1)
}

def iFlowName = args[0]
def iFlowId = args[1]
def packageId = args[2]
def iFlowDir = args[3]
def currentiFlowVersion = args[4]
def host_tmn = args[5]
def user = args[6]
def pw = args[7]

// Get current iFlow Version and bump up the number before upload
println "[INFO] Current IFlow Version in Tenant - ${currentiFlowVersion}"
def matcher = (currentiFlowVersion =~ /(\S+\.)(\d+)\s*/)
if (matcher.size()) {
    def patchNo = matcher[0][2] as int
    currentiFlowVersion = "${matcher[0][1]}${patchNo + 1}"
}
println "[INFO] New IFlow Version to be updated - ${currentiFlowVersion}"

// Update the manifest file with new version number
println "[INFO] Updating MANIFEST.MF"
File manifestFile = new File("${iFlowDir}/META-INF/MANIFEST.MF")
def manifestContent = manifestFile.getText('UTF-8')
def updatedContent = manifestContent.replaceFirst(/Bundle-Version: \S+/, "Bundle-Version: ${currentiFlowVersion}")
manifestFile.setText(updatedContent, 'UTF-8')

// Zip iFlow directory and encode to Base 64
ByteArrayOutputStream baos = new ByteArrayOutputStream()
ZipUtil.pack(new File(iFlowDir), baos)
def iFlowContent = baos.toByteArray().encodeBase64().toString()

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact('https', host_tmn, 443, user, pw)
designTimeArtifact.update(iFlowContent, iFlowId, iFlowName, packageId)