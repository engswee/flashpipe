package analytics

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestConstructQueryParameters(t *testing.T) {
	cmd := &cobra.Command{
		Use:     "artifact",
		Version: "3.2.0",
	}
	cmd.Flags().String("tmn-host", "", "")
	cmd.Flags().String("tmn-userid", "", "")
	cmd.Flags().String("oauth-clientid", "", "")
	cmd.Flags().String("artifact-type", "", "")
	cmd.Flags().String("file-param", "", "")
	cmd.Flags().String("file-manifest", "", "")
	cmd.Flags().StringSlice("script-collection-map", nil, "")

	cmd.Flags().Set("tmn-host", "test_host")
	cmd.Flags().Set("tmn-userid", "test_userid")
	cmd.Flags().Set("oauth-clientid", "test_clientid")
	cmd.Flags().Set("artifact-type", "Integration")
	cmd.Flags().Set("file-param", "parameter.prop")
	cmd.Flags().Set("file-manifest", "MANIFEST.MF")
	cmd.Flags().Set("script-collection-map", "CollectionMap")

	startTime := time.Now()
	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("FLASHPIPE_ACTION", "true")

	params := constructQueryParameters(cmd, nil, "2", startTime)

	// Site Id
	assert.Equal(t, "2", params["idsite"], "Expected parameter idsite = 2")
	// Action name
	assert.Equal(t, "artifact", params["action_name"], "Expected parameter action_name = artifact")
	// Version
	assert.Equal(t, "3.2.0", params["dimension2"], "Expected parameter dimension2 = 3.2.0")
	// Host name (hashed)
	assert.Equal(t, HashString("test_host"), params["uid"], "Expected parameter uid = Hash of test_host")
	// User ID or Client ID (hashed)
	assert.Equal(t, HashString("test_userid:test_clientid"), params["dimension1"], "Expected parameter uid = Hash of test_userid:test_clientid")
	// CI/CD platform
	assert.Equal(t, "GitHubActions", params["dimension3"], "Expected parameter dimension3 = GitHubActions")
	// FlashPipe Action used
	assert.Equal(t, "true", params["dimension18"], "Expected parameter dimension18 = true")
	// Processing Status
	assert.Equal(t, "Success", params["dimension4"], "Expected parameter dimension4 = Success")
	// Artifact Type
	assert.Equal(t, "Integration", params["dimension6"], "Expected parameter dimension6 = Integration")
	// Custom Parameters Used
	assert.Equal(t, "true", params["dimension7"], "Expected parameter dimension7 = true")
	// Custom Manifest Used
	assert.Equal(t, "true", params["dimension8"], "Expected parameter dimension8 = true")
	// Script Collection Used
	assert.Equal(t, "true", params["dimension9"], "Expected parameter dimension9 = true")
}
