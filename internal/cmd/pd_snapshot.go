package cmd

import (
	"fmt"
	"time"

	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/repo"
	"github.com/engswee/flashpipe/internal/str"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewPDSnapshotCommand() *cobra.Command {

	pdSnapshotCmd := &cobra.Command{
		Use:   "pd-snapshot",
		Short: "Download partner directory parameters from SAP CPI",
		Long: `Download all partner directory parameters from SAP CPI and save them locally.

This command retrieves both string and binary parameters from the SAP CPI Partner Directory
and organizes them in a local directory structure:

  {PID}/
    String.properties    - String parameters as key=value pairs
    Binary/              - Binary parameters as individual files
      {ParamId}.{ext}    - Binary parameter files
      _metadata.json     - Content type metadata

The snapshot operation supports two modes:
  - Replace mode (default): Overwrites existing local files
  - Add-only mode: Only adds new parameters, preserves existing values

Authentication is performed using OAuth 2.0 client credentials flow or Basic Auth.`,
		Example: `  # Snapshot with OAuth (environment variables)
  export FLASHPIPE_TMN_HOST="your-tenant.hana.ondemand.com"
  export FLASHPIPE_OAUTH_HOST="your-tenant.authentication.eu10.hana.ondemand.com"
  export FLASHPIPE_OAUTH_CLIENTID="your-client-id"
  export FLASHPIPE_OAUTH_CLIENTSECRET="your-client-secret"
  flashpipe pd-snapshot

  # Snapshot with explicit credentials and custom path
  flashpipe pd-snapshot \
    --tmn-host "your-tenant.hana.ondemand.com" \
    --oauth-host "your-tenant.authentication.eu10.hana.ondemand.com" \
    --oauth-clientid "your-client-id" \
    --oauth-clientsecret "your-client-secret" \
    --resources-path "./partner-directory"

  # Snapshot in add-only mode (don't overwrite existing values)
  flashpipe pd-snapshot --replace=false

  # Snapshot only specific PIDs
  flashpipe pd-snapshot --pids "SAP_SYSTEM_001,CUSTOMER_API"`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runPDSnapshot(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	// Define flags
	// Note: These can be set in config file under 'pd-snapshot' key
	pdSnapshotCmd.Flags().String("resources-path", "./partner-directory",
		"Path to save partner directory parameters")
	pdSnapshotCmd.Flags().Bool("replace", true,
		"Replace existing values (false = add only missing values)")
	pdSnapshotCmd.Flags().StringSlice("pids", nil,
		"Comma separated list of Partner IDs to snapshot (e.g., 'PID1,PID2')")

	return pdSnapshotCmd
}

func runPDSnapshot(cmd *cobra.Command) error {
	serviceDetails := api.GetServiceDetails(cmd)

	log.Info().Msg("Executing Partner Directory Snapshot command")

	// Support reading from config file under 'pd-snapshot' key
	resourcesPath := getConfigStringWithFallback(cmd, "resources-path", "pd-snapshot.resources-path")
	replace := getConfigBoolWithFallback(cmd, "replace", "pd-snapshot.replace")
	pids := getConfigStringSliceWithFallback(cmd, "pids", "pd-snapshot.pids")

	log.Info().Msgf("Resources Path: %s", resourcesPath)
	log.Info().Msgf("Replace Mode: %v", replace)
	if len(pids) > 0 {
		log.Info().Msgf("Filter PIDs: %v", pids)
	}

	// Trim PIDs
	pids = str.TrimSlice(pids)

	// Initialise HTTP executer
	exe := api.InitHTTPExecuter(serviceDetails)

	// Initialise Partner Directory API
	pdAPI := api.NewPartnerDirectory(exe)

	// Initialise Partner Directory Repository
	pdRepo := repo.NewPartnerDirectory(resourcesPath)

	// Execute snapshot
	if err := snapshotPartnerDirectory(pdAPI, pdRepo, replace, pids); err != nil {
		return err
	}

	log.Info().Msg("🏆 Partner Directory Snapshot completed successfully")
	return nil
}

func snapshotPartnerDirectory(pdAPI *api.PartnerDirectory, pdRepo *repo.PartnerDirectory, replace bool, pidsFilter []string) error {
	log.Info().Msg("Starting Partner Directory Snapshot...")

	// Download string parameters
	stringCount, err := snapshotStringParameters(pdAPI, pdRepo, replace, pidsFilter)
	if err != nil {
		return fmt.Errorf("failed to download string parameters: %w", err)
	}
	log.Info().Msgf("Downloaded %d string parameters", stringCount)

	// Download binary parameters
	binaryCount, err := snapshotBinaryParameters(pdAPI, pdRepo, replace, pidsFilter)
	if err != nil {
		return fmt.Errorf("failed to download binary parameters: %w", err)
	}
	log.Info().Msgf("Downloaded %d binary parameters", binaryCount)

	return nil
}

func snapshotStringParameters(pdAPI *api.PartnerDirectory, pdRepo *repo.PartnerDirectory, replace bool, pidsFilter []string) (int, error) {
	log.Debug().Msg("Fetching string parameters from Partner Directory")

	parameters, err := pdAPI.GetStringParameters("Pid,Id,Value")
	if err != nil {
		return 0, err
	}

	// Filter by PIDs if specified
	if len(pidsFilter) > 0 {
		filtered := make([]api.StringParameter, 0)
		for _, param := range parameters {
			if contains(pidsFilter, param.Pid) {
				filtered = append(filtered, param)
			}
		}
		parameters = filtered
	}

	log.Debug().Msgf("Fetched %d string parameters from Partner Directory", len(parameters))

	// Group by PID
	paramsByPid := make(map[string][]api.StringParameter)
	for _, param := range parameters {
		paramsByPid[param.Pid] = append(paramsByPid[param.Pid], param)
	}

	// Log all PIDs found
	pids := make([]string, 0, len(paramsByPid))
	for pid := range paramsByPid {
		pids = append(pids, pid)
	}
	log.Info().Msgf("Found %d Partner IDs with string parameters: %v", len(pids), pids)

	// Process each PID
	for pid, pidParams := range paramsByPid {
		log.Info().Msgf("Processing string parameters for PID: %s (%d parameters)", pid, len(pidParams))

		if err := pdRepo.WriteStringParameters(pid, pidParams, replace); err != nil {
			return 0, fmt.Errorf("failed to write string parameters for PID %s: %w", pid, err)
		}
	}

	return len(parameters), nil
}

func snapshotBinaryParameters(pdAPI *api.PartnerDirectory, pdRepo *repo.PartnerDirectory, replace bool, pidsFilter []string) (int, error) {
	log.Debug().Msg("Fetching binary parameters from Partner Directory")

	parameters, err := pdAPI.GetBinaryParameters("")
	if err != nil {
		return 0, err
	}

	// Filter by PIDs if specified
	if len(pidsFilter) > 0 {
		filtered := make([]api.BinaryParameter, 0)
		for _, param := range parameters {
			if contains(pidsFilter, param.Pid) {
				filtered = append(filtered, param)
			}
		}
		parameters = filtered
	}

	log.Debug().Msgf("Fetched %d binary parameters from Partner Directory", len(parameters))

	// Group by PID
	paramsByPid := make(map[string][]api.BinaryParameter)
	for _, param := range parameters {
		paramsByPid[param.Pid] = append(paramsByPid[param.Pid], param)
	}

	// Log all PIDs found
	pids := make([]string, 0, len(paramsByPid))
	for pid := range paramsByPid {
		pids = append(pids, pid)
	}
	log.Info().Msgf("Found %d Partner IDs with binary parameters: %v", len(pids), pids)

	// Process each PID
	for pid, pidParams := range paramsByPid {
		log.Info().Msgf("Processing binary parameters for PID: %s (%d parameters)", pid, len(pidParams))

		if err := pdRepo.WriteBinaryParameters(pid, pidParams, replace); err != nil {
			return 0, fmt.Errorf("failed to write binary parameters for PID %s: %w", pid, err)
		}
	}

	return len(parameters), nil
}
