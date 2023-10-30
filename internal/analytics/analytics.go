package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

type analyticsData struct {
	UniqueID         string `json:"UniqueID"`
	HashID           string `json:"HashID"`
	Command          string `json:"Command"`
	ArtifactType     string `json:"ArtifactType,omitempty"`
	ArtifactNameUsed bool   `json:"ArtifactNameUsed"`
}

var Host string

func Log(cmd *cobra.Command) {
	if Host != "" {
		log.Info().Msgf("Logging to %v", Host)
		logToAnalytics(cmd)
	}
}

func logToAnalytics(cmd *cobra.Command) {

	tmnHost := config.GetString(cmd, "tmn-host")
	tmnUserId := config.GetString(cmd, "tmn-userid")
	oauthClientId := config.GetString(cmd, "oauth-clientid")
	uniqueKey := fmt.Sprintf("%v:%v:%v", tmnHost, tmnUserId, oauthClientId)

	ctx := cmd.Context()
	executedCmd := ctx.Value("command").(string)

	v := &analyticsData{
		UniqueID: uniqueKey,
		HashID:   HashString(uniqueKey),
		Command:  executedCmd,
	}

	// Some environment variable - GitHub Action, Azure, Travis CI, Bitbucket, Gitlab, scrape for Jenkins
	//os.Environ()
	log.Info().Msgf("Environment variables: %v", os.Environ())

	// Flags used for each command,
	switch executedCmd {
	case "artifact":
		// if flag is not empty
		artifactType := config.GetString(cmd, "artifact-type")
		v.ArtifactType = artifactType
		//if config.GetString(cmd, "artifact-id") != "" {
		//
		//}

	}

	// create JSON output

	out, err := json.Marshal(v)
	if err != nil {
		// TODO - what happens, log a warning?
		return
	}
	log.Info().Msgf("Executed command: %s", out)
}

func HashString(input string) string {
	// Create a new SHA256 hash object
	hasher := sha256.New()

	// Convert the input string to bytes and hash it
	hasher.Write([]byte(input))

	// Get the hashed result as a byte slice
	hashedBytes := hasher.Sum(nil)

	// Convert the byte slice to a hexadecimal string
	hashedString := hex.EncodeToString(hashedBytes)

	return hashedString
}
