package cmd

import (
	"bytes"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/spf13/cobra"
	"os"
	"testing"
)

func TestArtifact(t *testing.T) {

	updateCmd := NewUpdateCommand()
	updateCmd.AddCommand(NewArtifactCommand())
	updateCmd.AddCommand(NewPackageCommand())
	rootCmd := NewCmdRoot()
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(NewDeployCommand())

	var args []string
	args = append(args, "update", "package")
	args = append(args, "--package-file", "../../test/testdata/FlashPipeIntegrationTest.json")
	rootCmd.SetArgs(args)

	_, _, err := ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("update package failed with error %v", err)
	}

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

	args = nil
	args = append(args, "deploy")
	args = append(args, "--artifact-ids", "Integration_Test_IFlow")

	_, _, err = ExecuteCommandC(rootCmd, args...)
	if err != nil {
		t.Fatalf("deploy failed with error %v", err)
	}

	// Clean up
	exe := odata.InitHTTPExecuter(&odata.ServiceDetails{
		Host:              os.Getenv("FLASHPIPE_TMN_HOST"),
		OauthHost:         os.Getenv("FLASHPIPE_OAUTH_HOST"),
		OauthPath:         os.Getenv("FLASHPIPE_OAUTH_PATH"),
		OauthClientId:     os.Getenv("FLASHPIPE_OAUTH_CLIENTID"),
		OauthClientSecret: os.Getenv("FLASHPIPE_OAUTH_CLIENTSECRET"),
	})
	ip := odata.NewIntegrationPackage(exe)
	err = ip.Delete("FlashPipeIntegrationTest")
	if err != nil {
		t.Fatalf("Delete package failed with error %v", err)
	}
}

func ExecuteCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
