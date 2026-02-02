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
	"github.com/engswee/flashpipe/internal/str"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewSnapshotCommand() *cobra.Command {

	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot integration packages from tenant to Git",
		Long: `Snapshot all editable integration packages from SAP Integration Suite
tenant to a Git repository.

Configuration:
  Settings can be loaded from the global config file (--config) under the
  'snapshot' section. CLI flags override config file settings.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate Draft Handling
			draftHandling := config.GetStringWithFallback(cmd, "draft-handling", "snapshot.draftHandling")
			switch draftHandling {
			case "SKIP", "ADD", "ERROR":
			default:
				return fmt.Errorf("invalid value for --draft-handling = %v", draftHandling)
			}
			// If artifacts directory is provided, validate that is it a subdirectory of Git repo
			gitRepoDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-git-repo", "snapshot.dirGitRepo")
			if err != nil {
				return fmt.Errorf("security alert for --dir-git-repo: %w", err)
			}

			if gitRepoDir != "" {
				artifactsDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-artifacts", "snapshot.dirArtifacts")
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
	// Note: These can be set in config file under 'snapshot' key
	snapshotCmd.PersistentFlags().String("dir-git-repo", "", "Directory of Git repository (config: snapshot.dirGitRepo)")
	snapshotCmd.PersistentFlags().String("dir-artifacts", "", "Directory containing contents of artifacts (grouped into packages) (config: snapshot.dirArtifacts)")
	snapshotCmd.PersistentFlags().String("dir-work", "/tmp", "Working directory for in-transit files (config: snapshot.dirWork)")
	snapshotCmd.Flags().String("draft-handling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR (config: snapshot.draftHandling)")
	snapshotCmd.PersistentFlags().StringSlice("ids-include", nil, "List of included package IDs (config: snapshot.idsInclude)")
	snapshotCmd.PersistentFlags().StringSlice("ids-exclude", nil, "List of excluded package IDs (config: snapshot.idsExclude)")

	snapshotCmd.Flags().String("git-commit-msg", "Tenant snapshot of "+time.Now().Format(time.UnixDate), "Message used in commit (config: snapshot.gitCommitMsg)")
	snapshotCmd.Flags().String("git-commit-user", "github-actions[bot]", "User used in commit (config: snapshot.gitCommitUser)")
	snapshotCmd.Flags().String("git-commit-email", "41898282+github-actions[bot]@users.noreply.github.com", "Email used in commit (config: snapshot.gitCommitEmail)")
	snapshotCmd.Flags().Bool("git-skip-commit", false, "Skip committing changes to Git repository (config: snapshot.gitSkipCommit)")
	snapshotCmd.Flags().Bool("sync-package-details", true, "Sync details of Integration Packages (config: snapshot.syncPackageDetails)")

	_ = snapshotCmd.MarkFlagRequired("dir-git-repo")
	snapshotCmd.MarkFlagsMutuallyExclusive("ids-include", "ids-exclude")

	return snapshotCmd
}

func runSnapshot(cmd *cobra.Command) error {
	log.Info().Msg("Executing snapshot command")

	// Support reading from config file under 'snapshot' key
	gitRepoDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-git-repo", "snapshot.dirGitRepo")
	if err != nil {
		return fmt.Errorf("security alert for --dir-git-repo: %w", err)
	}
	artifactsBaseDir := config.GetStringWithFallback(cmd, "dir-artifacts", "snapshot.dirArtifacts")
	if artifactsBaseDir == "" {
		artifactsBaseDir = gitRepoDir
	}
	workDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-work", "snapshot.dirWork")
	if err != nil {
		return fmt.Errorf("security alert for --dir-work: %w", err)
	}
	draftHandling := config.GetStringWithFallback(cmd, "draft-handling", "snapshot.draftHandling")
	includedIds := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "ids-include", "snapshot.idsInclude"))
	excludedIds := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "ids-exclude", "snapshot.idsExclude"))
	commitMsg := config.GetStringWithFallback(cmd, "git-commit-msg", "snapshot.gitCommitMsg")
	commitUser := config.GetStringWithFallback(cmd, "git-commit-user", "snapshot.gitCommitUser")
	commitEmail := config.GetStringWithFallback(cmd, "git-commit-email", "snapshot.gitCommitEmail")
	skipCommit := config.GetBoolWithFallback(cmd, "git-skip-commit", "snapshot.gitSkipCommit")
	syncPackageLevelDetails := config.GetBoolWithFallback(cmd, "sync-package-details", "snapshot.syncPackageDetails")

	serviceDetails := api.GetServiceDetails(cmd)
	err = getTenantSnapshot(serviceDetails, artifactsBaseDir, workDir, draftHandling, syncPackageLevelDetails, includedIds, excludedIds)
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

func getTenantSnapshot(serviceDetails *api.ServiceDetails, artifactsBaseDir string, workDir string, draftHandling string, syncPackageLevelDetails bool, includedIds []string, excludedIds []string) error {
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
			return err
		}
		if !readOnly {
			// Filter in/out artifacts
			if str.FilterIDs(id, includedIds, excludedIds) {
				continue
			}
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
