package cmd

import (
	"bytes"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestCPICommands(t *testing.T) {

	// ------------ Set up ------------
	println("---------- Setting up test - start ----------")
	exe := api.InitHTTPExecuter(&api.ServiceDetails{
		Host:              os.Getenv("FLASHPIPE_TMN_HOST"),
		OauthHost:         os.Getenv("FLASHPIPE_OAUTH_HOST"),
		OauthPath:         os.Getenv("FLASHPIPE_OAUTH_PATH"),
		OauthClientId:     os.Getenv("FLASHPIPE_OAUTH_CLIENTID"),
		OauthClientSecret: os.Getenv("FLASHPIPE_OAUTH_CLIENTSECRET"),
	})
	ip := api.NewIntegrationPackage(exe)
	dt := api.NewDesigntimeArtifact("Integration", exe)
	rt := api.NewRuntime(exe)
	println("---------- Setting up test - end ----------")

	updateCmd := NewUpdateCommand()
	updateCmd.AddCommand(NewArtifactCommand())
	updateCmd.AddCommand(NewPackageCommand())
	rootCmd := NewCmdRoot()
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(NewDeployCommand())
	rootCmd.AddCommand(NewSyncCommand())
	rootCmd.AddCommand(NewSnapshotCommand())

	// 1 - Create integration package
	var args []string
	args = append(args, "update", "package")
	args = append(args, "--package-file", "../../test/testdata/FlashPipeIntegrationTest.json")

	_, _, err := ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update package failed with error %v", err)
	}

	// Check package was created
	_, _, packageExists, err := ip.Get("FlashPipeIntegrationTest")
	if err != nil {
		t.Fatalf("Get integration package failed with error %v", err)
	}
	assert.True(t, packageExists, "Integration package was not created")

	// 2 - Create integration flow
	args = nil
	args = append(args, "update", "artifact")
	args = append(args, "--artifact-id", "Integration_Test_IFlow")
	args = append(args, "--artifact-name", "Integration Test IFlow")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--package-name", "FlashPipe Integration Test")
	args = append(args, "--dir-artifact", "../../test/testdata/artifacts/create/Integration_Test_IFlow")
	args = append(args, "--dir-work", "../../output/update/work")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update artifact failed with error %v", err)
	}

	// Check integration was created
	_, artifactDescription, integrationExists, err := dt.Get("Integration_Test_IFlow", "active")
	if err != nil {
		t.Fatalf("Get integration flow failed with error %v", err)
	}
	assert.True(t, integrationExists, "Integration flow was not created")
	assert.Equal(t, "Integration Created", artifactDescription, "Artifact has incorrect description")

	// 3 - Deploy integration flow
	args = nil
	args = append(args, "deploy")
	args = append(args, "--artifact-ids", "Integration_Test_IFlow")
	args = append(args, "--delay-length", "15")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("deploy failed with error %v", err)
	}

	// Check runtime was deployed
	_, status, err := rt.Get("Integration_Test_IFlow")
	if err != nil {
		t.Fatalf("Get runtime artifact failed with error %v", err)
	}
	assert.True(t, strings.HasPrefix(status, "START"), "Integration flow was not deployed")

	// 4 - Sync to Git
	args = nil
	args = append(args, "sync")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--dir-git-repo", "../../")
	args = append(args, "--dir-artifacts", "../../output/sync/artifact")
	args = append(args, "--dir-work", "../../output/sync/git/work")
	args = append(args, "--sync-package-details")
	args = append(args, "--git-skip-commit")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync failed with error %v", err)
	}
	assert.True(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/META-INF/MANIFEST.MF"), "MANIFEST.MF does not exist")
	assert.False(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/src/main/resources/parameters.prop"), "parameters.prop exists")
	assert.True(t, file.Exists("../../output/sync/artifact/FlashPipeIntegrationTest.json"), "FlashPipeIntegrationTest.json does not exist")

	// 5 - Update integration package
	args = nil
	args = append(args, "update", "package")
	args = append(args, "--package-file", "../../test/testdata/FlashPipeIntegrationTest_Update.json")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update package failed with error %v", err)
	}
	// Check package was updated
	packageData, _, _, err := ip.Get("FlashPipeIntegrationTest")
	if err != nil {
		t.Fatalf("Get integration package failed with error %v", err)
	}
	assert.Equal(t, "1.0.1", packageData.Root.Version, "Integration package was not updated to version 1.0.1")

	// 6 - Update integration flow
	args = nil
	args = append(args, "update", "artifact")
	args = append(args, "--artifact-id", "Integration_Test_IFlow")
	args = append(args, "--artifact-name", "Integration Test IFlow")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--package-name", "FlashPipe Integration Test")
	args = append(args, "--dir-artifact", "../../test/testdata/artifacts/update/Integration_Test_IFlow")
	args = append(args, "--dir-work", "../../output/update/work")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update artifact failed with error %v", err)
	}

	// Check integration was updated
	integrationVersion, artifactDescription, _, err := dt.Get("Integration_Test_IFlow", "active")
	if err != nil {
		t.Fatalf("Get integration flow failed with error %v", err)
	}
	assert.Equal(t, "1.0.1", integrationVersion, "Integration flow was not updated to version 1.0.1")
	assert.Equal(t, "Integration Updated", artifactDescription, "Artifact has incorrect description")

	// 7 - Deploy integration flow
	args = nil
	args = append(args, "deploy")
	args = append(args, "--artifact-ids", "Integration_Test_IFlow")
	args = append(args, "--delay-length", "15")
	args = append(args, "--compare-versions=false")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("deploy failed with error %v", err)
	}

	// Check runtime was updated
	runtimeVersion, _, err := rt.Get("Integration_Test_IFlow")
	if err != nil {
		t.Fatalf("Get runtime artifact failed with error %v", err)
	}
	assert.Equal(t, "1.0.1", runtimeVersion, "Runtime version of integration flow was not updated to version 1.0.1")

	// 8 - Sync updates to Git
	args = nil
	args = append(args, "sync")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--dir-git-repo", "../../")
	args = append(args, "--dir-artifacts", "../../output/sync/artifact")
	args = append(args, "--dir-work", "../../output/sync/git/work")
	args = append(args, "--sync-package-details")
	args = append(args, "--git-skip-commit")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync failed with error %v", err)
	}
	assert.True(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/META-INF/MANIFEST.MF"), "MANIFEST.MF does not exist")
	assert.True(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/src/main/resources/parameters.prop"), "parameters.prop does not exist")
	packageDataFromTenant, err := api.GetPackageDetails("../../output/sync/artifact/FlashPipeIntegrationTest.json")
	if err != nil {
		t.Fatalf("Unable to read integration package file with error %v", err)
	}
	assert.Equal(t, "1.0.1", packageDataFromTenant.Root.Version, "Integration package was not updated to version 1.0.1")

	// 9 - Snapshot to Git
	args = nil
	args = append(args, "snapshot")
	args = append(args, "--dir-git-repo", "../../output/snapshot/repo")
	args = append(args, "--dir-work", "../../output/snapshot/work")
	args = append(args, "--sync-package-details")
	args = append(args, "--git-skip-commit")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync failed with error %v", err)
	}
	assert.True(t, file.Exists("../../output/snapshot/repo/FlashPipeIntegrationTest/Integration_Test_IFlow/META-INF/MANIFEST.MF"), "MANIFEST.MF does not exist")
	assert.True(t, file.Exists("../../output/snapshot/repo/FlashPipeIntegrationTest/Integration_Test_IFlow/src/main/resources/parameters.prop"), "parameters.prop does not exist")

	// 10 - Sync updates to tenant
	args = nil
	args = append(args, "sync")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--dir-git-repo", "../../test/testdata/artifacts/create")
	args = append(args, "--dir-artifacts", "")
	args = append(args, "--target", "tenant")
	args = append(args, "--dir-work", "../../output/sync/tenant/work")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync failed with error %v", err)
	}
	artifacts, err := ip.GetAllArtifacts("FlashPipeIntegrationTest")
	if err != nil {
		t.Fatalf("GetAllArtifacts failed with error - %v", err)
	}

	assert.Equal(t, "1.0.0", api.FindArtifactById("Integration_Test_IFlow", artifacts).Version, "Integration_Test_IFlow was not updated to version 1.0.0")
	assert.Equal(t, "1.0.0", api.FindArtifactById("Integration_Test_Message_Mapping", artifacts).Version, "Integration_Test_Message_Mapping was not updated to version 1.0.0")
	assert.Equal(t, "1.0.0", api.FindArtifactById("Integration_Test_Script_Collection", artifacts).Version, "Integration_Test_Script_Collection was not updated to version 1.0.0")
	assert.Equal(t, "1.0.0", api.FindArtifactById("Integration_Test_Value_Mapping", artifacts).Version, "Integration_Test_Value_Mapping was not updated to version 1.0.0")

	// ------------ Clean up ------------
	println("---------- Tearing down test - start ----------")
	err = ip.Delete("FlashPipeIntegrationTest")
	if err != nil {
		t.Logf("WARNING - Delete package failed with error %v", err)
	}
	err = rt.UnDeploy("Integration_Test_IFlow")
	if err != nil {
		t.Logf("WARNING - Undeploy integration failed with error %v", err)
	}
	err = os.RemoveAll("../../output/update")
	if err != nil {
		t.Logf("WARNING - Directory removal failed with error - %v", err)
	}
	err = os.RemoveAll("../../output/sync")
	if err != nil {
		t.Logf("WARNING - Directory removal failed with error - %v", err)
	}
	err = os.RemoveAll("../../output/snapshot")
	if err != nil {
		t.Logf("WARNING - Directory removal failed with error - %v", err)
	}
	println("---------- Tearing down test - end ----------")
}

func TestAPIMCommands(t *testing.T) {

	// ------------ Set up ------------
	println("---------- Setting up test - start ----------")
	exe := api.InitHTTPExecuter(&api.ServiceDetails{
		Host:              os.Getenv("FLASHPIPE_APIPORTAL_HOST"),
		OauthHost:         os.Getenv("FLASHPIPE_OAUTH_HOST"),
		OauthPath:         os.Getenv("FLASHPIPE_OAUTH_PATH"),
		OauthClientId:     os.Getenv("FLASHPIPE_APIPORTAL_OAUTH_CLIENTID"),
		OauthClientSecret: os.Getenv("FLASHPIPE_APIPORTAL_OAUTH_CLIENTSECRET"),
	})
	a := api.NewAPIProxy(exe)
	println("---------- Setting up test - end ----------")

	rootCmd := NewCmdRoot()
	syncCmd := NewSyncCommand()
	syncCmd.AddCommand(NewAPIMCommand())
	rootCmd.AddCommand(syncCmd)

	var args []string
	// 1 - Sync APIM to Git
	args = append(args, "sync", "apim")
	args = append(args, "--tmn-host", os.Getenv("FLASHPIPE_APIPORTAL_HOST"))
	args = append(args, "--oauth-clientid", os.Getenv("FLASHPIPE_APIPORTAL_OAUTH_CLIENTID"))
	args = append(args, "--oauth-clientsecret", os.Getenv("FLASHPIPE_APIPORTAL_OAUTH_CLIENTSECRET"))
	args = append(args, "--dir-git-repo", "../../")
	args = append(args, "--dir-artifacts", "../../output/apim/artifact")
	args = append(args, "--dir-work", "../../output/apim/git/work")
	args = append(args, "--ids-include", "HelloWorldAPI")
	args = append(args, "--git-skip-commit")

	_, _, err := ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync apim git failed with error %v", err)
	}
	assert.True(t, file.Exists("../../output/apim/artifact/HelloWorldAPI/manifest.json"), "manifest.json does not exist")

	// 2 - Sync APIM to tenant
	args = nil
	args = append(args, "sync", "apim")
	args = append(args, "--tmn-host", os.Getenv("FLASHPIPE_APIPORTAL_HOST"))
	args = append(args, "--oauth-clientid", os.Getenv("FLASHPIPE_APIPORTAL_OAUTH_CLIENTID"))
	args = append(args, "--oauth-clientsecret", os.Getenv("FLASHPIPE_APIPORTAL_OAUTH_CLIENTSECRET"))
	args = append(args, "--dir-artifacts", "../../test/testdata/apim")
	args = append(args, "--dir-work", "../../output/apim/tenant/work")
	args = append(args, "--ids-include", "Northwind_V4")
	args = append(args, "--target", "tenant")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync apim tenant failed with error %v", err)
	}
	proxyExists, err := a.Get("Northwind_V4")
	if err != nil {
		t.Fatalf("Get APIProxy failed with error %v", err)
	}
	assert.True(t, proxyExists, "APIProxy was not uploaded")

	// ------------ Clean up ------------
	println("---------- Tearing down test - start ----------")
	err = a.Delete("Northwind_V4")
	if err != nil {
		t.Logf("WARNING - Delete failed with error - %v", err)
	}
	err = os.RemoveAll("../../output/apim")
	if err != nil {
		t.Logf("WARNING - Directory removal failed with error - %v", err)
	}
	println("---------- Tearing down test - end ----------")
}

func ExecuteCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
