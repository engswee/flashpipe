package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact

if (args.length < 6) {
    println "Enter arguments in the format: <iflow_id> <version> <cpi_host> <user> <password> <output_file>"
    System.exit(1)
}

def iFlowId = args[0]
def iFlowVersion = args[1]
def host_tmn = args[2]
def user = args[3]
def pw = args[4]
def outputFile = args[5]

DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact('https', host_tmn, 443, user, pw)

File outputZip = new File(outputFile)
outputZip.bytes = designTimeArtifact.download(iFlowId, iFlowVersion)