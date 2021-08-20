package io.github.engswee.flashpipe.cpi.util

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.IntegrationPackage
import io.github.engswee.flashpipe.http.HTTPExecuter
import org.zeroturnaround.zip.ZipUtil

class TestHelper {
    final HTTPExecuter httpExecuter

    TestHelper(HTTPExecuter httpExecuter) {
        this.httpExecuter = httpExecuter
    }

    void setupIFlow(String packageId, String packageName, String iFlowId, String iFlowName, String filePath) {
        IntegrationPackage integrationPackage = new IntegrationPackage(httpExecuter)
        CSRFToken csrfToken = new CSRFToken(httpExecuter)
        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(httpExecuter)
        // Create integration package if it doesn't exist
        if (!integrationPackage.exists(packageId)) {
            integrationPackage.create(packageId, packageName, csrfToken)
        }
        // Upload IFlow if it doesn't exist
        if (!designTimeArtifact.getVersion(iFlowId, 'active', true)) {
            def base64IFlowContent = generateBase64IFlowContent(filePath)
            designTimeArtifact.upload(base64IFlowContent, iFlowId, iFlowName, packageId, csrfToken)
        }
    }

    private String generateBase64IFlowContent(String filePath) {
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(filePath), baos)
        return baos.toByteArray().encodeBase64().toString()
    }
}