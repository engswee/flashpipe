package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/repo"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewSnapshotCommand() *cobra.Command {
	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot integration packages from tenant to Git",
		Long: `Snapshot all editable integration packages from SAP Integration Suite
tenant to a Git repository.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate Draft Handling
			draftHandling := config.GetString(cmd, "draft-handling")
			switch draftHandling {
			case "SKIP", "ADD", "ERROR":
			default:
				return fmt.Errorf("invalid value for --draft-handling = %v", draftHandling)
			}
			// If artifacts directory is provided, validate that is it a subdirectory of Git repo
			gitRepoDir, err := config.GetStringWithEnvExpand(cmd, "dir-git-repo")
			if err != nil {
				return fmt.Errorf("security alert for --dir-git-repo: %w", err)
			}

			if gitRepoDir != "" {
				artifactsDir, err := config.GetStringWithEnvExpand(cmd, "dir-artifacts")
				if err != nil {
					return fmt.Errorf("security alert for --dir-artifacts: %w", err)
				}
				gitRepoDirClean := filepath.Clean(gitRepoDir) + string(os.PathSeparator)
				if artifactsDir != "" && !strings.HasPrefix(artifactsDir, gitRepoDirClean) {
					return fmt.Errorf("--dir-artifacts [%v] should be a subdirectory of --dir-git-repo [%v]", artifactsDir, gitRepoDirClean)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runSnapshot(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	snapshotCmd.Flags().String("dir-git-repo", "", "Directory of Git repository")
	snapshotCmd.Flags().String("dir-artifacts", "", "Directory containing contents of artifacts (grouped into packages)")
	snapshotCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files")
	snapshotCmd.Flags().String("draft-handling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR")
	snapshotCmd.Flags().String("git-commit-msg", "Tenant snapshot of "+time.Now().Format(time.UnixDate), "Message used in commit")
	snapshotCmd.Flags().String("git-commit-user", "github-actions[bot]", "User used in commit")
	snapshotCmd.Flags().String("git-commit-email", "41898282+github-actions[bot]@users.noreply.github.com", "Email used in commit")
	snapshotCmd.Flags().Bool("git-skip-commit", false, "Skip committing changes to Git repository")
	snapshotCmd.Flags().Bool("sync-package-details", false, "Sync details of Integration Packages")

	_ = snapshotCmd.MarkFlagRequired("dir-git-repo")

	return snapshotCmd
}

func runSnapshot(cmd *cobra.Command) error {
	log.Info().Msg("Executing snapshot command")

	gitRepoDir, err := config.GetStringWithEnvExpand(cmd, "dir-git-repo")
	if err != nil {
		return fmt.Errorf("security alert for --dir-git-repo: %w", err)
	}
	artifactsBaseDir, err := config.GetStringWithEnvExpandWithDefault(cmd, "dir-artifacts", gitRepoDir)
	if err != nil {
		return fmt.Errorf("security alert for --dir-artifacts: %w", err)
	}
	workDir, err := config.GetStringWithEnvExpand(cmd, "dir-work")
	if err != nil {
		return fmt.Errorf("security alert for --dir-work: %w", err)
	}
	draftHandling := config.GetString(cmd, "draft-handling")
	commitMsg := config.GetString(cmd, "git-commit-msg")
	commitUser := config.GetString(cmd, "git-commit-user")
	commitEmail := config.GetString(cmd, "git-commit-email")
	skipCommit := config.GetBool(cmd, "git-skip-commit")
	syncPackageLevelDetails := config.GetBool(cmd, "sync-package-details")

	serviceDetails := api.GetServiceDetails(cmd)
	err = getTenantSnapshot(serviceDetails, artifactsBaseDir, workDir, draftHandling, syncPackageLevelDetails)
	if err != nil {
		return err
	}

	if !skipCommit {
		err = repo.CommitToRepo(gitRepoDir, commitMsg, commitUser, commitEmail)
		if err != nil {
			return err
		}
	}
	return nil
}

func getTenantSnapshot(serviceDetails *api.ServiceDetails, artifactsBaseDir string, workDir string, draftHandling string, syncPackageLevelDetails bool) error {
	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msg("📢 Begin taking a snapshot of the tenant")

	// Initialise HTTP executer
	exe := api.InitHTTPExecuter(serviceDetails)

	// Get packages from the tenant
	ip := api.NewIntegrationPackage(exe)
	ids, err := ip.GetPackagesList()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("No packages found in the tenant")
	}

	log.Info().Msgf("Processing %d packages", len(ids))
	synchroniser := sync.New(exe)
	for i, id := range ids {
		log.Info().Msg("---------------------------------------------------------------------------------")
		log.Info().Msgf("Processing package %d/%d - ID: %v", i+1, len(ids), id)
		packageWorkingDir := fmt.Sprintf("%v/%v", workDir, id)
		packageArtifactsDir := fmt.Sprintf("%v/%v", artifactsBaseDir, id)
		packageDataFromTenant, readOnly, _, err := synchroniser.VerifyDownloadablePackage(id)
		if err != nil {
			if err != nil {
				return err
			}
		}
		if !readOnly {
			if syncPackageLevelDetails {
				err = synchroniser.PackageToGit(packageDataFromTenant, id, packageWorkingDir, packageArtifactsDir)
				if err != nil {
					return err
				}
			}
			err = synchroniser.ArtifactsToGit(id, packageWorkingDir, packageArtifactsDir, nil, nil, draftHandling, "ID", nil)
			if err != nil {
				return err
			}

		}
	}

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msg("🏆 Completed taking a snapshot of the tenant")
	return nil
}
