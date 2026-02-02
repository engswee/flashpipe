package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/str"

	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/repo"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewSyncCommand() *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync designtime artifacts between tenant and Git",
		Long: `Synchronise designtime artifacts between SAP Integration Suite
tenant and a Git repository.

Configuration:
  Settings can be loaded from the global config file (--config) under the
  'sync' section. CLI flags override config file settings.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate Directory Naming Type
			dirNamingType := config.GetStringWithFallback(cmd, "dir-naming-type", "sync.dirNamingType")
			switch dirNamingType {
			case "ID", "NAME":
			default:
				return fmt.Errorf("invalid value for --dir-naming-type = %v", dirNamingType)
			}
			// Validate Draft Handling
			draftHandling := config.GetStringWithFallback(cmd, "draft-handling", "sync.draftHandling")
			switch draftHandling {
			case "SKIP", "ADD", "ERROR":
			default:
				return fmt.Errorf("invalid value for --draft-handling = %v", draftHandling)
			}
			// If artifacts directory is provided, validate that is it a subdirectory of Git repo
			gitRepoDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-git-repo", "sync.dirGitRepo")
			if err != nil {
				return fmt.Errorf("security alert for --dir-git-repo: %w", err)
			}
			if gitRepoDir != "" {
				artifactsDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-artifacts", "sync.dirArtifacts")
				if err != nil {
					return fmt.Errorf("security alert for --dir-artifacts: %w", err)
				}
				gitRepoDirClean := filepath.Clean(gitRepoDir) + string(os.PathSeparator)
				if artifactsDir != "" && !strings.HasPrefix(artifactsDir, gitRepoDirClean) {
					return fmt.Errorf("--dir-artifacts [%v] should be a subdirectory of --dir-git-repo [%v]", artifactsDir, gitRepoDirClean)
				}
			}
			// Validate target
			target := config.GetStringWithFallback(cmd, "target", "sync.target")
			switch target {
			case "git", "tenant":
			default:
				return fmt.Errorf("invalid value for --target = %v", target)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runSync(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	// Note: These can be set in config file under 'sync' key
	syncCmd.Flags().String("package-id", "", "ID of Integration Package (config: sync.packageId)")
	syncCmd.PersistentFlags().String("dir-git-repo", "", "Directory of Git repository (config: sync.dirGitRepo)")
	syncCmd.PersistentFlags().String("dir-artifacts", "", "Directory containing contents of artifacts (config: sync.dirArtifacts)")
	syncCmd.PersistentFlags().String("dir-work", "/tmp", "Working directory for in-transit files (config: sync.dirWork)")
	syncCmd.Flags().String("dir-naming-type", "ID", "Name artifact directory by ID or Name. Allowed values: ID, NAME (config: sync.dirNamingType)")
	syncCmd.Flags().String("draft-handling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR (config: sync.draftHandling)")
	syncCmd.PersistentFlags().StringSlice("ids-include", nil, "List of included artifact IDs (config: sync.idsInclude)")
	syncCmd.PersistentFlags().StringSlice("ids-exclude", nil, "List of excluded artifact IDs (config: sync.idsExclude)")
	syncCmd.PersistentFlags().String("target", "git", "Target of sync. Allowed values: git, tenant (config: sync.target)")
	syncCmd.PersistentFlags().String("git-commit-msg", "Sync repo from tenant", "Message used in commit (config: sync.gitCommitMsg)")
	syncCmd.PersistentFlags().String("git-commit-user", "github-actions[bot]", "User used in commit (config: sync.gitCommitUser)")
	syncCmd.PersistentFlags().String("git-commit-email", "41898282+github-actions[bot]@users.noreply.github.com", "Email used in commit (config: sync.gitCommitEmail)")
	syncCmd.Flags().StringSlice("script-collection-map", nil, "Comma-separated source-target ID pairs for converting script collection references during sync (config: sync.scriptCollectionMap)")
	syncCmd.PersistentFlags().Bool("git-skip-commit", false, "Skip committing changes to Git repository (config: sync.gitSkipCommit)")
	syncCmd.Flags().Bool("sync-package-details", false, "Sync details of Integration Package (config: sync.syncPackageDetails)")

	_ = syncCmd.MarkFlagRequired("package-id")
	_ = syncCmd.MarkFlagRequired("dir-git-repo")
	syncCmd.MarkFlagsMutuallyExclusive("ids-include", "ids-exclude")

	return syncCmd
}

func runSync(cmd *cobra.Command) error {
	log.Info().Msg("Executing sync command")

	// Support reading from config file under 'sync' key
	packageId := config.GetStringWithFallback(cmd, "package-id", "sync.packageId")
	gitRepoDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-git-repo", "sync.dirGitRepo")
	if err != nil {
		return fmt.Errorf("security alert for --dir-git-repo: %w", err)
	}
	artifactsDir := config.GetStringWithFallback(cmd, "dir-artifacts", "sync.dirArtifacts")
	if artifactsDir == "" {
		artifactsDir = gitRepoDir
	}
	workDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-work", "sync.dirWork")
	if err != nil {
		return fmt.Errorf("security alert for --dir-work: %w", err)
	}
	dirNamingType := config.GetStringWithFallback(cmd, "dir-naming-type", "sync.dirNamingType")
	draftHandling := config.GetStringWithFallback(cmd, "draft-handling", "sync.draftHandling")
	includedIds := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "ids-include", "sync.idsInclude"))
	excludedIds := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "ids-exclude", "sync.idsExclude"))
	commitMsg := config.GetStringWithFallback(cmd, "git-commit-msg", "sync.gitCommitMsg")
	commitUser := config.GetStringWithFallback(cmd, "git-commit-user", "sync.gitCommitUser")
	commitEmail := config.GetStringWithFallback(cmd, "git-commit-email", "sync.gitCommitEmail")
	scriptCollectionMap := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "script-collection-map", "sync.scriptCollectionMap"))
	skipCommit := config.GetBoolWithFallback(cmd, "git-skip-commit", "sync.gitSkipCommit")
	syncPackageLevelDetails := config.GetBoolWithFallback(cmd, "sync-package-details", "sync.syncPackageDetails")
	target := config.GetStringWithFallback(cmd, "target", "sync.target")

	serviceDetails := api.GetServiceDetails(cmd)
	// Initialise HTTP executer
	exe := api.InitHTTPExecuter(serviceDetails)
	synchroniser := sync.New(exe)

	// Sync from tenant to Git
	if target == "git" {
		packageDataFromTenant, readOnly, _, err := synchroniser.VerifyDownloadablePackage(packageId)
		if err != nil {
			return err
		}
		if !readOnly {
			if syncPackageLevelDetails {
				err = synchroniser.PackageToGit(packageDataFromTenant, packageId, workDir, artifactsDir)
				if err != nil {
					return err
				}
			}

			err = synchroniser.ArtifactsToGit(packageId, workDir, artifactsDir, includedIds, excludedIds, draftHandling, dirNamingType, scriptCollectionMap)
			if err != nil {
				return err
			}

			if !skipCommit {
				err = repo.CommitToRepo(gitRepoDir, commitMsg, commitUser, commitEmail)
				if err != nil {
					return err
				}
			}
		}
	}

	// Sync from Git to tenant
	if target == "tenant" {
		// Check for existence of package in tenant
		_, _, packageExists, err := synchroniser.VerifyDownloadablePackage(packageId)
		if !packageExists {
			// If the definition for the integration package is available, then update it from the file
			packageFile := fmt.Sprintf("%v/%v.json", artifactsDir, packageId)
			if file.Exists(packageFile) {
				packageSynchroniser := sync.NewSyncer("tenant", "CPIPackage", exe)
				err = packageSynchroniser.Exec(sync.Request{PackageFile: packageFile})
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Package %v does not exist. Please run 'update package' command first", packageId)
			}
		}
		if err != nil {
			return err
		}

		err = synchroniser.ArtifactsToTenant(packageId, workDir, artifactsDir, includedIds, excludedIds)
		if err != nil {
			return err
		}
	}
	return nil
}
