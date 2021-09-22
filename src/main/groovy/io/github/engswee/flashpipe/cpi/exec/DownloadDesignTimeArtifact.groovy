package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class DownloadDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(DownloadDesignTimeArtifact)

    String iFlowId
    String iFlowVersion
    String outputFile

    static void main(String[] args) {
        DownloadDesignTimeArtifact downloadDesignTimeArtifact = new DownloadDesignTimeArtifact()
        downloadDesignTimeArtifact.getEnvironmentVariables()
        downloadDesignTimeArtifact.execute()
    }

    @Override
    void getEnvironmentVariables() {
        this.iFlowId = getMandatoryEnvVar('IFLOW_ID')
        this.iFlowVersion = getMandatoryEnvVar('IFLOW_VER')
        this.outputFile = getMandatoryEnvVar('OUTPUT_FILE')
    }

    @Override
    void execute() {
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)

        File outputZip = new File(this.outputFile)
        logger.info("Downloading IFlow ${this.iFlowId}")
        outputZip.bytes = designTimeArtifact.download(this.iFlowId, this.iFlowVersion)
        logger.info("IFlow downloaded to ${this.outputFile}")
    }
}