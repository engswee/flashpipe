package io.github.engswee.flashpipe.cpi.exec

import io.github.engswee.flashpipe.cpi.api.CSRFToken
import io.github.engswee.flashpipe.cpi.api.DesignTimeArtifact
import io.github.engswee.flashpipe.cpi.api.RuntimeArtifact
import io.github.engswee.flashpipe.cpi.util.ManifestHandler
import io.github.engswee.flashpipe.cpi.util.ScriptCollection
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.zeroturnaround.zip.ZipUtil

class UpdateDesignTimeArtifact extends APIExecuter {

    static Logger logger = LoggerFactory.getLogger(UpdateDesignTimeArtifact)

    String iFlowId
    String iFlowName
    String iFlowDir
    String packageId
    String versionHandling
    String scriptCollectionMap

    static void main(String[] args) {
        UpdateDesignTimeArtifact updateDesignTimeArtifact = new UpdateDesignTimeArtifact()
        updateDesignTimeArtifact.getEnvironmentVariables()
        try {
            updateDesignTimeArtifact.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        this.iFlowId = getMandatoryEnvVar('IFLOW_ID')
        this.iFlowName = getMandatoryEnvVar('IFLOW_NAME')
        this.iFlowDir = getMandatoryEnvVar('IFLOW_DIR')
        this.packageId = getMandatoryEnvVar('PACKAGE_ID')
        this.versionHandling = System.getenv('VERSION_HANDLING')
        this.scriptCollectionMap = System.getenv('SCRIPT_COLLECTION_MAP')
    }

    @Override
    void execute() {
        if (this.versionHandling) {
            logger.warn('⚠️ VERSION_HANDLING is deprecated and will be removed in a future release!')
            logger.info('The current behavior will be as though VERSION_HANDLING = MANIFEST, meaning version number will depend on Bundle-Version set in META-INF/MANIFEST.MF file')
        }
        // Check that input environment variables do not have any of the secrets in their values
        validateInputContainsNoSecrets('IFLOW_ID', this.iFlowId)
        validateInputContainsNoSecrets('IFLOW_NAME', this.iFlowName)
        validateInputContainsNoSecrets('PACKAGE_ID', this.packageId)
        validateInputContainsNoSecrets('SCRIPT_COLLECTION_MAP', this.scriptCollectionMap)

        ManifestHandler manifestHandler = new ManifestHandler("${this.iFlowDir}/META-INF/MANIFEST.MF")
        manifestHandler.updateAttributes(this.iFlowId, this.iFlowName, ScriptCollection.newInstance(this.scriptCollectionMap).getTargetCollectionValues())
        manifestHandler.updateFile()

        // Zip iFlow directory and encode to Base 64
        ByteArrayOutputStream baos = new ByteArrayOutputStream()
        ZipUtil.pack(new File(this.iFlowDir), baos)
        def iFlowContent = baos.toByteArray().encodeBase64().toString()

        DesignTimeArtifact designTimeArtifact = new DesignTimeArtifact(this.httpExecuter)
        CSRFToken csrfToken = new CSRFToken(this.httpExecuter)
        designTimeArtifact.update(iFlowContent, this.iFlowId, this.iFlowName, this.packageId, csrfToken)
        logger.info("IFlow ${this.iFlowId} updated")

        // If runtime has the same version no, then undeploy it, otherwise it gets skipped during deployment
        def designtimeVersion = designTimeArtifact.getVersion(this.iFlowId, 'active', false)
        RuntimeArtifact runtimeArtifact = new RuntimeArtifact(this.httpExecuter)
        def runtimeVersion = runtimeArtifact.getVersion(this.iFlowId)

        if (runtimeVersion == designtimeVersion) {
            logger.info('Undeploying existing runtime artifact with same version number due to changes in design')
            runtimeArtifact.undeploy(this.iFlowId, csrfToken)
        }
    }
}