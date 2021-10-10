package io.github.engswee.flashpipe.cpi.util

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import io.github.engswee.flashpipe.http.HTTPExecuter
import org.zeroturnaround.zip.ZipUtil

import java.util.concurrent.TimeUnit

class IntegrationTestHelper {
    final CSRFToken csrfToken
    final DesignTimeArtifact designTimeArtifact
    final RuntimeArtifact runtimeArtifact
    final IntegrationPackage integrationPackage

    IntegrationTestHelper(HTTPExecuter httpExecuter) {
        this.csrfToken = new CSRFToken(httpExecuter)
        this.designTimeArtifact = new DesignTimeArtifact(httpExecuter)
        this.runtimeArtifact = new RuntimeArtifact(httpExecuter)
        this.integrationPackage = new IntegrationPackage(httpExecuter)
    }

    void setupIFlow(String packageId, String packageName, String iFlowId, String iFlowName, String filePath) {
        // Create integration package if it doesn't exist
        if (!this.integrationPackage.exists(packageId)) {
            this.integrationPackage.create(packageId, packageName, this.csrfToken)
        }
        // Upload IFlow if it doesn't exist
        if (!this.designTimeArtifact.getVersion(iFlowId, 'active', true)) {
            def base64IFlowContent = generateBase64IFlowContent(filePath)
            this.designTimeArtifact.upload(base64IFlowContent, iFlowId, iFlowName, packageId, this.csrfToken)
        }
    }

    void cleanupIFlow(String iFlowId) {
        this.designTimeArtifact.delete(iFlowId, this.csrfToken)
        this.runtimeArtifact.undeploy(iFlowId, this.csrfToken)
    }

    void undeployIFlow(String iFlowId) {
        this.runtimeArtifact.undeploy(iFlowId, this.csrfToken)
    }
    
    void deployIFlow(String iFlowId, boolean waitForCompletion) {
        this.designTimeArtifact.deploy(iFlowId, this.csrfToken)
        if (waitForCompletion) {
            while (true) {
                TimeUnit.SECONDS.sleep(10)
                def status = this.runtimeArtifact.getStatus(iFlowId)
                if (status != 'STARTING') {
                    if (status == 'STARTED') {
                        break
                    } else {
                        throw new RuntimeException('IFlow deployment unsuccessful')
                    }
                }
            }
        }
    }

    private String generateBase64IFlowContent(String filePath) {
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(filePath), baos)
        return baos.toByteArray().encodeBase64().toString()
    }
}