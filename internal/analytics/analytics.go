package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var Host string
var SiteId string
var ShowLogs string

func Log(cmd *cobra.Command, err error) {
	if Host != "" && SiteId != "" {
		if ShowLogs == "true" {
			log.Debug().Msg("Logging to Matomo Analytics")
		}

		logToAnalytics(cmd, err, Host, "https", 443, SiteId, ShowLogs == "true")
	}
}

func logToAnalytics(cmd *cobra.Command, cmdErr error, analyticsHost string, analyticsHostScheme string, analyticsHostPort int, analyticsSiteId string, showLogs bool) {

	params := constructQueryParameters(cmd, cmdErr, analyticsSiteId)

	urlPath := fmt.Sprintf("/matomo.php?%s", MapToString(params))

	exe := httpclnt.New("", "", "", "", "", "", analyticsHost, analyticsHostScheme, analyticsHostPort, showLogs)
	_, err := exe.ExecGetRequest(urlPath, nil)
	if err != nil && showLogs {
		log.Error().Msgf("Analytics logging error: %s", err.Error())
	}
}

func constructQueryParameters(cmd *cobra.Command, cmdErr error, analyticsSiteId string) map[string]string {
	tmnHost := config.GetString(cmd, "tmn-host")
	tmnUserId := config.GetString(cmd, "tmn-userid")
	oauthClientId := config.GetString(cmd, "oauth-clientid")
	uniqueKey := fmt.Sprintf("%v:%v", tmnUserId, oauthClientId)

	// Matomo API reference - https://developer.matomo.org/api-reference/tracking-api
	params := map[string]string{}
	params["idsite"] = analyticsSiteId // Build flag distinguishes between QA and Prod site
	params["rec"] = "1"
	params["new_visit"] = "1"
	params["action_name"] = cmd.Name()
	params["apiv"] = "1"
	params["uid"] = HashString(tmnHost)

	// Custom dimensions
	params["dimension1"] = HashString(uniqueKey)
	params["dimension2"] = cmd.Version

	// CI/CD platform
	envVars := strings.Join(os.Environ(), ",")
	if os.Getenv("SYSTEM_ISAZUREVM") == "1" {
		params["dimension3"] = "Azure"
	} else if os.Getenv("GITHUB_ACTIONS") == "true" {
		params["dimension3"] = "GitHubActions"
		if os.Getenv("FLASHPIPE_ACTION") == "true" {
			params["dimension18"] = "true"
		}
	} else if os.Getenv("TRAVIS") == "true" {
		params["dimension3"] = "TravisCI"
	} else if strings.Contains(envVars, "BITBUCKET_") {
		params["dimension3"] = "Bitbucket"
	} else if strings.Contains(envVars, "JENKINS") {
		params["dimension3"] = "Jenkins"
	} else if os.Getenv("GITLAB_CI") == "true" {
		params["dimension3"] = "Gitlab"
	} else {
		params["dimension3"] = "CLI/Unknown"
	}

	if cmdErr != nil {
		params["dimension4"] = "Error"
		params["dimension5"] = logger.GetErrorDetails(cmdErr)
	} else {
		params["dimension4"] = "Success"
	}

	// Flags used for each command,
	switch cmd.Name() {
	case "artifact":
		artifactType := config.GetString(cmd, "artifact-type")
		params["dimension6"] = artifactType
		parametersFile := config.GetString(cmd, "file-param")
		if parametersFile != "" {
			params["dimension7"] = "true"
		}
		manifestFile := config.GetString(cmd, "file-manifest")
		if manifestFile != "" {
			params["dimension8"] = "true"
		}
		scriptMap := config.GetStringSlice(cmd, "script-collection-map")
		if len(scriptMap) > 0 {
			params["dimension9"] = "true"
		}

	case "deploy":
		artifactType := config.GetString(cmd, "artifact-type")
		params["dimension6"] = artifactType
		delayLength := config.GetInt(cmd, "delay-length")
		params["dimension10"] = fmt.Sprintf("%v", delayLength)
		maxCheckLimit := config.GetInt(cmd, "max-check-limit")
		params["dimension11"] = fmt.Sprintf("%v", maxCheckLimit)

	case "sync":
		target := config.GetString(cmd, "target")
		params["dimension12"] = target

		dirNamingType := config.GetString(cmd, "dir-naming-type")
		params["dimension13"] = dirNamingType
		draftHandling := config.GetString(cmd, "draft-handling")
		params["dimension14"] = draftHandling

		includedIds := config.GetStringSlice(cmd, "ids-include")
		if len(includedIds) > 0 {
			params["dimension15"] = "true"
		}
		excludedIds := config.GetStringSlice(cmd, "ids-exclude")
		if len(excludedIds) > 0 {
			params["dimension16"] = "true"
		}
		scriptCollectionMap := config.GetStringSlice(cmd, "script-collection-map")
		if len(scriptCollectionMap) > 0 {
			params["dimension9"] = "true"
		}
		syncPackageLevelDetails := config.GetBool(cmd, "sync-package-details")
		params["dimension17"] = fmt.Sprintf("%v", syncPackageLevelDetails)

	}
	return params
}

func MapToString(m map[string]string) string {
	var parts []string

	for key, value := range m {
		pair := fmt.Sprintf("%s=%s", key, value)
		parts = append(parts, pair)
	}

	return strings.Join(parts, "&")
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
