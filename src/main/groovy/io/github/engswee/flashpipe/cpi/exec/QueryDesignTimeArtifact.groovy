package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class QueryDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(QueryDesignTimeArtifact)

    String iFlowId
    String packageId

    static void main(String[] args) {
        QueryDesignTimeArtifact queryDesignTimeArtifact = new QueryDesignTimeArtifact()
        queryDesignTimeArtifact.getEnvironmentVariables()
        try {
            queryDesignTimeArtifact.execute()
        } catch (ExecutionException e) {
            if (e.getMessage() == 'Active version of IFlow does not exist') {
                println('Active version of IFlow does not exist')
                System.exit(99)
            } else {
                System.exit(1)
            }
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.iFlowId = getMandatoryEnvVar('IFLOW_ID')
        this.packageId = getMandatoryEnvVar('PACKAGE_ID')
    }

    @Override
    void execute() {
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        logger.info("Checking if ${this.iFlowId} exists")
        if (designTimeArtifact.getVersion(this.iFlowId, 'active', true)) {
            logger.info("Active version of IFlow ${this.iFlowId} exists")
            //  Check if version is in draft mode
            IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
            if (integrationPackage.iFlowInDraftVersion(this.packageId, this.iFlowId)) {
                logger.error("🛑 IFlow ${this.iFlowId} is in Draft state. Save Version of IFlow in Web UI first!")
                throw new ExecutionException('IFlow is in Draft state')
            }
        } else {
            logger.info("Active version of IFlow ${this.iFlowId} does not exist")
            throw new ExecutionException('Active version of IFlow does not exist')
        }
    }
}