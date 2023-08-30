package cmd

import (
	"bytes"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {

	// ------------ Set up ------------
	println("---------- Setting up test - start ----------")
	exe := odata.InitHTTPExecuter(&odata.ServiceDetails{
		Host:              os.Getenv("FLASHPIPE_TMN_HOST"),
		OauthHost:         os.Getenv("FLASHPIPE_OAUTH_HOST"),
		OauthPath:         os.Getenv("FLASHPIPE_OAUTH_PATH"),
		OauthClientId:     os.Getenv("FLASHPIPE_OAUTH_CLIENTID"),
		OauthClientSecret: os.Getenv("FLASHPIPE_OAUTH_CLIENTSECRET"),
	})
	ip := odata.NewIntegrationPackage(exe)
	dt := odata.NewDesigntimeArtifact("Integration", exe)
	rt := odata.NewRuntime(exe)
	println("---------- Setting up test - end ----------")

	updateCmd := NewUpdateCommand()
	updateCmd.AddCommand(NewArtifactCommand())
	updateCmd.AddCommand(NewPackageCommand())
	rootCmd := NewCmdRoot()
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(NewDeployCommand())
	rootCmd.AddCommand(NewSyncCommand())

	// 1 - Create integration package
	var args []string
	args = append(args, "update", "package")
	args = append(args, "--package-file", "../../test/testdata/FlashPipeIntegrationTest.json")
	rootCmd.SetArgs(args)

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
	rootCmd.SetArgs(args)

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update artifact failed with error %v", err)
	}

	// Check integration was created
	_, integrationExists, err := dt.Get("Integration_Test_IFlow", "active")
	if err != nil {
		t.Fatalf("Get integration flow failed with error %v", err)
	}
	assert.True(t, integrationExists, "Integration flow was not created")

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

	// 4 - Sync to local
	args = nil
	args = append(args, "sync", "package")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--dir-git-repo", "../../")
	args = append(args, "--dir-artifacts", "../../output/sync/artifact")
	args = append(args, "--dir-work", "../../output/sync/work")
	args = append(args, "--sync-package-details")
	args = append(args, "--git-skip-commit")
	rootCmd.SetArgs(args)

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync failed with error %v", err)
	}
	assert.True(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/META-INF/MANIFEST.MF"), "MANIFEST.MF does not exist")
	assert.False(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/src/main/resources/parameters.prop"), "parameters.prop exists")

	// 5 - Update integration package
	args = nil
	args = append(args, "update", "package")
	args = append(args, "--package-file", "../../test/testdata/FlashPipeIntegrationTest_Update.json")
	rootCmd.SetArgs(args)

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
	rootCmd.SetArgs(args)

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update artifact failed with error %v", err)
	}

	// Check integration was updated
	integrationVersion, _, err := dt.Get("Integration_Test_IFlow", "active")
	if err != nil {
		t.Fatalf("Get integration flow failed with error %v", err)
	}
	assert.Equal(t, "1.0.1", integrationVersion, "Integration flow was not updated to version 1.0.1")

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

	// 8 - Sync updates to local
	args = nil
	args = append(args, "sync", "package")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--dir-git-repo", "../../")
	args = append(args, "--dir-artifacts", "../../output/sync/artifact")
	args = append(args, "--dir-work", "../../output/sync/work")
	args = append(args, "--sync-package-details")
	args = append(args, "--git-skip-commit")
	rootCmd.SetArgs(args)

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("sync failed with error %v", err)
	}
	assert.True(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/META-INF/MANIFEST.MF"), "MANIFEST.MF does not exist")
	assert.True(t, file.Exists("../../output/sync/artifact/Integration_Test_IFlow/src/main/resources/parameters.prop"), "parameters.prop does not exist")

	// ------------ Clean up ------------
	println("---------- Tearing down test - start ----------")
	err = ip.Delete("FlashPipeIntegrationTest")
	if err != nil {
		t.Fatalf("Delete package failed with error %v", err)
	}
	err = rt.UnDeploy("Integration_Test_IFlow")
	if err != nil {
		t.Fatalf("Undeploy integration failed with error %v", err)
	}
	err = os.RemoveAll("../../output/update")
	if err != nil {
		t.Fatalf("Directory removal failed with error - %v", err)
	}
	err = os.RemoveAll("../../output/sync")
	if err != nil {
		t.Fatalf("Directory removal failed with error - %v", err)
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
