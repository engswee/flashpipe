package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact

class DownloadDesignTimeArtifact extends APIExecuter {

    static void main(String[] args) {
        DownloadDesignTimeArtifact downloadDesignTimeArtifact = new DownloadDesignTimeArtifact()
        downloadDesignTimeArtifact.execute()
    }

    @Override
    void execute() {
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        def iFlowVersion = getMandatoryEnvVar('IFLOW_VER')
        def outputFile = getMandatoryEnvVar('OUTPUT_FILE')

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)

        File outputZip = new File(outputFile)
        outputZip.bytes = designTimeArtifact.download(iFlowId, iFlowVersion)
        println "[INFO] IFlow downloaded to ${outputFile}"
    }
}
