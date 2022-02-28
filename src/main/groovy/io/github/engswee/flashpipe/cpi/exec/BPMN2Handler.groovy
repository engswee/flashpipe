package io.github.engswee.flashpipe.cpi.exec

import groovy.xml.XmlUtil
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class BPMN2Handler extends APIExecuter {
    static Logger logger = LoggerFactory.getLogger(BPMN2Handler)

    Map collections
    String iFlowDir

    static void main(String[] args) {
        BPMN2Handler bpmn2Handler = new BPMN2Handler()
        bpmn2Handler.getEnvironmentVariables()
        try {
            bpmn2Handler.execute()
        } catch (ExecutionException ignored) {
            System.exit(1)
        }
    }

    @Override
    void getEnvironmentVariables() {
        String scriptCollectionMap = System.getenv('SCRIPT_COLLECTION_MAP')

        this.collections = scriptCollectionMap?.split(',')?.toList()?.collectEntries {
            String[] pair = it.split('=')
            [(pair[0]): pair[1]]
        }
        if (this.collections?.size()) {
            // Check that input environment variables do not have any of the secrets in their values
            validateInputContainsNoSecrets('SCRIPT_COLLECTION_MAP', scriptCollectionMap)

            this.iFlowDir = getMandatoryEnvVar('GIT_SRC_DIR')
        } else {
            logger.info('No update required for BPMN2 file as there are no script collections')
        }
    }

    @Override
    void execute() {
        if (this.collections?.size()) {
            updateFiles(this.collections, this.iFlowDir)
        }
    }

    void updateFiles(Map collections, String iFlowDir) {
        XmlParser parser = new XmlParser(false, false)
        logger.debug("Updating files in ${iFlowDir} with collection ${collections}")
        File bpmnDir = new File("${iFlowDir}/src/main/resources/scenarioflows/integrationflow")
        bpmnDir.listFiles().each { iFlowFile ->
            boolean contentUpdated = false
            logger.info("Processing BPMN2 file ${iFlowFile.toPath()}")
            Node root = parser.parse(iFlowFile)
            List scriptBundles = root.'**'.'bpmn2:callActivity'.'bpmn2:extensionElements'.'ifl:property'.findAll { it.key.text() == 'scriptBundleId' }
            scriptBundles.each { Node bundle ->
                def sourceValue = bundle.value.text()
                def targetValue = collections.get(sourceValue)
                if (sourceValue && targetValue) {
                    bundle.children().each { Node field ->
                        if (field.name() == 'value') {
                            logger.debug("Changing scriptBundleId from ${sourceValue} to ${targetValue}")
                            field.setValue(targetValue)
                            contentUpdated = true
                        }
                    }
                }
            }
            if (contentUpdated)
                iFlowFile.write(XmlUtil.serialize(root))
        }
    }
}