package io.github.engswee.flashpipe.cpi.util

import org.slf4j.Logger
import org.slf4j.LoggerFactory

import java.util.jar.Attributes
import java.util.jar.Manifest

class ManifestHandler {
    static Logger logger = LoggerFactory.getLogger(ManifestHandler)

    final Manifest manifest
    final File file
    boolean attributesUpdated = false

    static void main(String[] args) {
        String filePath = args[0]
        String iFlowId = args[1]
        String iFlowName = args[2]
        ScriptCollection scriptCollection = (args.size() > 3) ? ScriptCollection.newInstance(args[3]) : ScriptCollection.newInstance('')
        newInstance(filePath).normalizeAttributesInFile(iFlowId, iFlowName, scriptCollection.getTargetCollectionValues())
    }

    static ManifestHandler newInstance(String filePath) {
        return new ManifestHandler(filePath)
    }

    private ManifestHandler(String filePath) {
        this.file = new File(filePath)
        this.manifest = new Manifest(this.file.newInputStream())
    }

    void normalizeAttributesInFile(String iFlowId, String iFlowName, List collections) {
        updateAttributes(iFlowId, iFlowName, collections)
        updateFile()
    }

    void updateAttributes(String iFlowId, String iFlowName, List collections) {
        logger.debug("Updating MANIFEST.MF with ID=${iFlowId}, Name=${iFlowName} and Script Collection: ${collections.join(',')}")
        Attributes attributes = this.manifest.getMainAttributes()

        def bundleIdAttribute = new Attributes.Name('Bundle-SymbolicName')
        def bundleNameAttribute = new Attributes.Name('Bundle-Name')
        if (attributes.get(bundleIdAttribute).replace('; singleton:=true', '') != iFlowId) {
            attributes.put(bundleIdAttribute, iFlowId)
            this.attributesUpdated = true
        }
        if (attributes.get(bundleNameAttribute) != iFlowName) {
            attributes.put(bundleNameAttribute, iFlowName)
            this.attributesUpdated = true
        }
        // For each script collection, iterate and separate with comma
        if (collections.size()) {
            String capability = collections.collect { "scriptcollection.${it};resolution:=optional;bundleType:String=\"ScriptCollection\";source:String=\"reference\"" }.join(',')
            attributes.put(new Attributes.Name('Require-Capability'), capability)
            this.attributesUpdated = true
        }
    }

    private void updateFile() {
        if (this.attributesUpdated) {
            this.manifest.write(this.file.newOutputStream())
        }
    }
}