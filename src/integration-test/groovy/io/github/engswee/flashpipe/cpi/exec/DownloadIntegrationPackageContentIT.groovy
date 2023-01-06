package io.github.engswee.flashpipe.cpi.exec

import groovy.json.JsonSlurper
import io.github.engswee.flashpipe.cpi.util.FileUtility
import io.github.engswee.flashpipe.cpi.util.IntegrationTestHelper
import io.github.engswee.flashpipe.http.HTTPExecuter
import io.github.engswee.flashpipe.http.HTTPExecuterApacheImpl
import spock.lang.Shared
import spock.lang.Specification

class DownloadIntegrationPackageContentIT extends Specification {

    @Shared
    IntegrationTestHelper testHelper

    def setupSpec() {
        def host = System.getenv('HOST_TMN')
        def user = System.getenv('BASIC_USERID')
        def password = System.getenv('BASIC_PASSWORD')
        HTTPExecuter httpExecuter = HTTPExecuterApacheImpl.newInstance('https', host, 443, user, password)
        testHelper = new IntegrationTestHelper(httpExecuter)
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Update', 'FlashPipe Update', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update')
        testHelper.setupIFlow('FlashPipeIntegrationTest', 'FlashPipe Integration Test', 'FlashPipe_Upload', 'FlashPipe Upload', 'src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Upload')
    }

    def cleanupSpec() {
        testHelper.cleanupIFlow('FlashPipe_Update')
        testHelper.cleanupIFlow('FlashPipe_Upload')
    }

    def 'Download integration package by NAME'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        noExceptionThrown()
    }

    def 'Include sync on integration package'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')
        downloadIntegrationPackageContent.setSyncPackageLevelDetails('YES')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')
        downloadIntegrationPackageContent.setNormalizePackageIDPrefixOrSuffix('')
        downloadIntegrationPackageContent.setNormalizePackageNamePrefixOrSuffix('')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        Map packageContent = new JsonSlurper().parse(new File('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipeIntegrationTest.json'))
        verifyAll {
            packageContent.d.Id == 'FlashPipeIntegrationTest'
            packageContent.d.Name == 'FlashPipe Integration Test'
            packageContent.d.ShortText == 'FlashPipeIntegrationTest'
        }

        cleanup:
        new File('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipeIntegrationTest.json').delete()
    }

    def 'Download integration package by NAME with exclusion'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setExcludedIds(['FlashPipe_Upload'])
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        noExceptionThrown()
    }

    def 'Download integration package by ID with inclusion'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('ID')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setIncludedIds(['FlashPipe_Update'])
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        noExceptionThrown()

        cleanup:
        new File('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe_Update').deleteDir()
    }

    def 'Download integration package by ID with inclusion to non existent directory'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('target/DownloadIntegrationPackageContentIT/NewGitSrcDir')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('ID')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setIncludedIds(['FlashPipe_Update'])
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        noExceptionThrown()

        cleanup:
        new File('target/DownloadIntegrationPackageContentIT/NewGitSrcDir/FlashPipe_Update').deleteDir()
    }

    def 'Download integration package by NAME with inclusion with changes'() {
        given:
        // Setup temp Git Source directory for changes in target directory
        new File('target/DownloadIntegrationPackageContentIT/TempGitSrcDir/FlashPipe Update').mkdirs()
        FileUtility.copyDirectory('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows/FlashPipe Update', 'target/DownloadIntegrationPackageContentIT/TempGitSrcDir/FlashPipe Update')

        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('target/DownloadIntegrationPackageContentIT/TempGitSrcDir')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setIncludedIds(['FlashPipe_Update'])
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        File fileToDelete = new File('target/DownloadIntegrationPackageContentIT/TempGitSrcDir/FlashPipe Update/src/main/resources/scenarioflows/integrationflow/FlashPipe Update.iflw')
        fileToDelete.delete()

        when:
        downloadIntegrationPackageContent.execute()

        then:
        fileToDelete.exists() == true

        cleanup:
        new File('target/DownloadIntegrationPackageContentIT/TempGitSrcDir/FlashPipe Update').deleteDir()
    }

    def 'Exception thrown for invalid DIR_NAMING_TYPE'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('DUMMY')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Invalid value for DIR_NAMING_TYPE'
    }

    def 'Exception thrown for invalid DRAFT_HANDLING'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('DUMMY')
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Invalid value for DRAFT_HANDLING'
    }

    def 'Exception thrown when both INCLUDE_IDS and EXCLUDE_IDS are provided'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setIncludedIds(['ABC'])
        downloadIntegrationPackageContent.setExcludedIds(['XYZ'])
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'INCLUDE_IDS and EXCLUDE_IDS are mutually exclusive'
    }

    def 'Exception thrown when INCLUDE_IDS has invalid IFlow'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setIncludedIds(['DUMMY'])
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Invalid input in INCLUDE_IDS'
    }

    def 'Exception thrown when EXCLUDE_IDS has invalid IFlow'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setPackageId('FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setGitSrcDir('src/integration-test/resources/test-data/DesignTimeArtifact/IFlows')
        downloadIntegrationPackageContent.setWorkDir('target/DownloadIntegrationPackageContentIT/FlashPipeIntegrationTest')
        downloadIntegrationPackageContent.setDirNamingType('NAME')
        downloadIntegrationPackageContent.setDraftHandling('SKIP')
        downloadIntegrationPackageContent.setExcludedIds(['DUMMY'])
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')
        downloadIntegrationPackageContent.setNormalizePackageAction('NONE')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Invalid input in EXCLUDE_IDS'
    }

    def 'Exception thrown when COMMIT_MESSAGE contains'() {
        given:
        DownloadIntegrationPackageContent downloadIntegrationPackageContent = new DownloadIntegrationPackageContent()
        downloadIntegrationPackageContent.setCommitMessage(System.getenv('OAUTH_CLIENTID'))
        downloadIntegrationPackageContent.setNormalizeManifestAction('NONE')
        downloadIntegrationPackageContent.setNormalizeManifestPrefixOrSuffix('')
        downloadIntegrationPackageContent.setScriptCollectionMap('')

        when:
        downloadIntegrationPackageContent.execute()

        then:
        ExecutionException e = thrown()
        e.getMessage() == 'Environment variable contains value of secret variable'
    }
}