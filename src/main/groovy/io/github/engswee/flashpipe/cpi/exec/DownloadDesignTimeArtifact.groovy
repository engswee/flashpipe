package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class DownloadDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(DownloadDesignTimeArtifact)
    
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
        logger.info("Downloading IFlow ${iFlowId}")
        outputZip.bytes = designTimeArtifact.download(iFlowId, iFlowVersion)
        logger.info("IFlow downloaded to ${outputFile}")
    }
}
