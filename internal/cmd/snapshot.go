package cmd

import (
	"errors"
	"fmt"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/engswee/flashpipe/internal/repo"
	"github.com/engswee/flashpipe/internal/sync"
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
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
	snapshotCmd.Flags().String("dir-git-repo", "", "Directory of Git repository containing contents of artifacts (grouped into packages)")
	snapshotCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files [or set environment WORK_DIR]")
	snapshotCmd.Flags().String("drafthandling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR [or set environment DRAFT_HANDLING]")
	snapshotCmd.Flags().String("git-commitmsg", "Tenant snapshot of "+time.Now().Format(time.UnixDate), "Message used in commit [or set environment COMMIT_MESSAGE]")
	snapshotCmd.Flags().String("git-commit-user", "github-actions[bot]", "User used in commit")
	snapshotCmd.Flags().String("git-commit-email", "41898282+github-actions[bot]@users.noreply.github.com", "Email used in commit")
	snapshotCmd.Flags().Bool("syncpackagedetails", false, "Sync details of Integration Packages [or set environment SYNC_PACKAGE_LEVEL_DETAILS]")

	return snapshotCmd
}

func runSnapshot(cmd *cobra.Command) {
	log.Info().Msg("Executing snapshot command")

	gitRepoDir := config.GetMandatoryString(cmd, "dir-git-repo")
	workDir := config.GetString(cmd, "dir-work")
	draftHandling := config.GetString(cmd, "drafthandling")
	commitMsg := config.GetString(cmd, "git-commitmsg")
	commitUser := config.GetString(cmd, "git-commit-user")
	commitEmail := config.GetString(cmd, "git-commit-email")
	syncPackageLevelDetails := config.GetBool(cmd, "syncpackagedetails")

	serviceDetails := odata.GetServiceDetails(cmd)
	err := getTenantSnapshot(serviceDetails, gitRepoDir, workDir, draftHandling, syncPackageLevelDetails)
	logger.ExitIfError(err)

	err = repo.CommitToRepo(gitRepoDir, commitMsg, commitUser, commitEmail)
	logger.ExitIfError(err)
}

func getTenantSnapshot(serviceDetails *odata.ServiceDetails, gitRepoDir string, workDir string, draftHandling string, syncPackageLevelDetails bool) error {
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
		packageWorkingDir := fmt.Sprintf("%v/%v", workDir, id)
		packageArtifactsDir := fmt.Sprintf("%v/%v", gitRepoDir, id)
		if syncPackageLevelDetails {
			err = synchroniser.SyncPackageDetails(id, packageWorkingDir, packageArtifactsDir)
			logger.ExitIfError(err)
		}
		err = synchroniser.SyncArtifacts(id, packageWorkingDir, packageArtifactsDir, nil, nil, draftHandling, "ID", "")
		logger.ExitIfError(err)
	}

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msg("🏆 Completed taking a snapshot of the tenant")
	return nil
}
