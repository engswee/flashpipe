package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage

class QueryDesignTimeArtifact extends APIExecuter {

    static void main(String[] args) {
        QueryDesignTimeArtifact queryDesignTimeArtifact = new QueryDesignTimeArtifact()
        queryDesignTimeArtifact.execute()
    }

    @Override
    void execute() {
        def iFlowId = getMandatoryEnvVar('IFLOW_ID')
        def packageId = getMandatoryEnvVar('PACKAGE_ID')

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)

        if (designTimeArtifact.getVersion(iFlowId, 'active', true)) {
            println "[INFO] Active version of IFlow ${iFlowId} exists"
            //  Check if version is in draft mode
            IntegrationPackage integrationPackage = new IntegrationPackage(this.httpExecuter)
            if (integrationPackage.iFlowInDraftVersion(packageId, iFlowId)) {
                println "[ERROR] IFlow ${iFlowId} is in Draft state. Save Version of IFlow in Web UI first!"
                System.exit(1)
            }
        } else {
            println "[INFO] Active version of IFlow ${iFlowId} does not exist"
            System.exit(99)
        }
    }
}