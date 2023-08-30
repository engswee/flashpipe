package cmd

import (
	"bytes"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestArtifact(t *testing.T) {

	// Set up
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

	// Create integration package
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

	// Create integration flow
	args = nil
	args = append(args, "update", "artifact")
	args = append(args, "--artifact-id", "Integration_Test_IFlow")
	args = append(args, "--artifact-name", "Integration Test IFlow")
	args = append(args, "--package-id", "FlashPipeIntegrationTest")
	args = append(args, "--package-name", "FlashPipe Integration Test")
	args = append(args, "--dir-artifact", "../../test/testdata/artifacts/create/Integration_Test_IFlow")
	args = append(args, "--dir-work", "../../output/work")
	rootCmd.SetArgs(args)

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update artifact failed with error %v", err)
	}

	// Create integration was created
	_, integrationExists, err := dt.Get("Integration_Test_IFlow", "active")
	if err != nil {
		t.Fatalf("Get integration flow failed with error %v", err)
	}
	assert.True(t, integrationExists, "Integration flow was not created")

	// Deploy integration flow
	args = nil
	args = append(args, "deploy")
	args = append(args, "--artifact-ids", "Integration_Test_IFlow")

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

	// Clean up
	println("---------- Tearing down test - start ----------")
	err = ip.Delete("FlashPipeIntegrationTest")
	if err != nil {
		t.Fatalf("Delete package failed with error %v", err)
	}
	err = rt.UnDeploy("Integration_Test_IFlow")
	if err != nil {
		t.Fatalf("Undeploy integration failed with error %v", err)
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
