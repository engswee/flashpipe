package io.github.engswee.flashpipe.cpi.util

import spock.lang.Specification

import java.util.jar.Attributes

class ManifestHandlerSpec extends Specification {

    def 'ID, Name and Capability attributes updated'() {
        given:
        ManifestHandler manifestHandler = ManifestHandler.newInstance('src/test/resources/test-data/ManifestHandling/MANIFEST.MF')

        when:
        manifestHandler.updateAttributes('FlashPipe_Use_Script_Collection', 'FlashPipe Use Script Collection', ['Common_Scripts'])

        then:
        Attributes attributes = manifestHandler.getManifest().getMainAttributes()
        verifyAll {
            attributes.get(new Attributes.Name('Bundle-SymbolicName')) == 'FlashPipe_Use_Script_Collection'
            attributes.get(new Attributes.Name('Bundle-Name')) == 'FlashPipe Use Script Collection'
            attributes.get(new Attributes.Name('Require-Capability')) == 'scriptcollection.Common_Scripts;resolution:=optional;bundleType:String="ScriptCollection";source:String="reference"'
        }
    }

    def 'No script collection'() {
        given:
        ManifestHandler manifestHandler = ManifestHandler.newInstance('src/test/resources/test-data/ManifestHandling/MANIFEST_NoScriptCollection.MF')

        when:
        manifestHandler.updateAttributes('FlashPipe_Use_Script_Collection', 'FlashPipe Use Script Collection', [])

        then:
        Attributes attributes = manifestHandler.getManifest().getMainAttributes()
        attributes.get(new Attributes.Name('Require-Capability')) == null
    }
}