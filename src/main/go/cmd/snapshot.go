package cmd

import (
	"errors"
	"fmt"
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/engswee/flashpipe/repo"
	"github.com/engswee/flashpipe/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"
)

func NewSnapshotCommand() *cobra.Command {

	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot integration packages from tenant to Git",
		Long: `Snapshot all editable integration packages from SAP Integration Suite
tenant to a Git repository.`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Validate Draft Handling
			draftHandling := config.GetString(cmd, "drafthandling")
			switch draftHandling {
			case "SKIP", "ADD", "ERROR":
			default:
				return fmt.Errorf("invalid value for --drafthandling = %v", draftHandling)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runSnapshot(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	snapshotCmd.Flags().String("dir-gitsrc", "", "Base directory containing contents of artifacts (grouped into packages) [or set environment GIT_SRC_DIR]")
	snapshotCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	snapshotCmd.Flags().String("drafthandling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR [or set environment DRAFT_HANDLING]")
	snapshotCmd.Flags().String("git-commitmsg", "Tenant snapshot of "+time.Now().Format(time.UnixDate), "Message used in commit [or set environment COMMIT_MESSAGE]")
	snapshotCmd.Flags().Bool("syncpackagedetails", false, "Sync details of Integration Packages [or set environment SYNC_PACKAGE_LEVEL_DETAILS]")

	return snapshotCmd
}

func runSnapshot(cmd *cobra.Command) {
	log.Info().Msg("Executing snapshot command")

	gitSrcDir := config.GetMandatoryString(cmd, "dir-gitsrc")
	workDir := config.GetString(cmd, "dir-work")
	draftHandling := config.GetString(cmd, "drafthandling")
	commitMsg := config.GetString(cmd, "git-commitmsg")
	syncPackageLevelDetails := config.GetBool(cmd, "syncpackagedetails")

	serviceDetails := odata.GetServiceDetails(cmd)
	err := getTenantSnapshot(serviceDetails, gitSrcDir, workDir, draftHandling, syncPackageLevelDetails)
	logger.ExitIfError(err)

	err = repo.CommitToRepo(gitSrcDir, commitMsg)
	logger.ExitIfError(err)
}

func getTenantSnapshot(serviceDetails *odata.ServiceDetails, gitSrcDir string, workDir string, draftHandling string, syncPackageLevelDetails bool) error {
	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msg("📢 Begin taking a snapshot of the tenant")

	// Initialise HTTP executer
	exe := odata.InitHTTPExecuter(serviceDetails)

	// Get packages from the tenant
	ip := odata.NewIntegrationPackage(exe)
	ids, err := ip.GetPackagesList()
	logger.ExitIfError(err)
	if len(ids) == 0 {
		return errors.New("No packages found in the tenant")
	}

	log.Info().Msgf("Processing %d packages", len(ids))
	synchroniser := sync.New(exe)
	for i, id := range ids {
		log.Info().Msg("---------------------------------------------------------------------------------")
		log.Info().Msgf("Processing package %d/%d - ID: %v", i+1, len(ids), id)
		if syncPackageLevelDetails {
			synchroniser.SyncPackageDetails(id)
		}
		synchroniser.SyncArtifacts(id, fmt.Sprintf("%v/%v", workDir, id), fmt.Sprintf("%v/%v", gitSrcDir, id), nil, nil, draftHandling, "ID", "NONE", "", "")
	}

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msg("🏆 Completed taking a snapshot of the tenant")
	return nil
}
