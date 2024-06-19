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

func NewSyncCommand() *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync designtime artifacts between tenant and Git",
		Long: `Synchronise designtime artifacts between SAP Integration Suite
tenant and a Git repository.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate Directory Naming Type
			dirNamingType := config.GetString(cmd, "dir-naming-type")
			switch dirNamingType {
			case "ID", "NAME":
			default:
				return fmt.Errorf("invalid value for --dir-naming-type = %v", dirNamingType)
			}
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
			// Validate target
			target := config.GetString(cmd, "target")
			switch target {
			case "local", "remote":
				log.Warn().Msg("--target = local/remote is deprecated, use --target = git/tenant")
			case "git", "tenant":
			default:
				return fmt.Errorf("invalid value for --target = %v", target)
			}
			// TODO - Validate secrets in env var, lower priority as it is no longer resolved in GitHub action workflow
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
	syncCmd.Flags().String("package-id", "", "ID of Integration Package")
	syncCmd.PersistentFlags().String("dir-git-repo", "", "Directory of Git repository")
	syncCmd.PersistentFlags().String("dir-artifacts", "", "Directory containing contents of artifacts")
	syncCmd.PersistentFlags().String("dir-work", "/tmp", "Working directory for in-transit files")
	syncCmd.Flags().String("dir-naming-type", "ID", "Name artifact directory by ID or Name. Allowed values: ID, NAME")
	syncCmd.Flags().String("draft-handling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR")
	syncCmd.PersistentFlags().StringSlice("ids-include", nil, "List of included artifact IDs")
	syncCmd.PersistentFlags().StringSlice("ids-exclude", nil, "List of excluded artifact IDs")
	syncCmd.PersistentFlags().String("target", "git", "Target of sync. Allowed values: git, tenant, local(deprecated), remote(deprecated)")
	syncCmd.PersistentFlags().String("git-commit-msg", "Sync repo from tenant", "Message used in commit")
	syncCmd.PersistentFlags().String("git-commit-user", "github-actions[bot]", "User used in commit")
	syncCmd.PersistentFlags().String("git-commit-email", "41898282+github-actions[bot]@users.noreply.github.com", "Email used in commit")
	syncCmd.Flags().StringSlice("script-collection-map", nil, "Comma-separated source-target ID pairs for converting script collection references during sync ")
	syncCmd.PersistentFlags().Bool("git-skip-commit", false, "Skip committing changes to Git repository")
	syncCmd.Flags().Bool("sync-package-details", false, "Sync details of Integration Package")

	_ = syncCmd.MarkFlagRequired("package-id")
	_ = syncCmd.MarkFlagRequired("dir-git-repo")
	syncCmd.MarkFlagsMutuallyExclusive("ids-include", "ids-exclude")

	return syncCmd
}

func runSync(cmd *cobra.Command) error {
	log.Info().Msg("Executing sync command")

	packageId := config.GetString(cmd, "package-id")
	gitRepoDir, err := config.GetStringWithEnvExpand(cmd, "dir-git-repo")
	if err != nil {
		return fmt.Errorf("security alert for --dir-git-repo: %w", err)
	}
	artifactsDir, err := config.GetStringWithEnvExpandWithDefault(cmd, "dir-artifacts", gitRepoDir)
	if err != nil {
		return fmt.Errorf("security alert for --dir-artifacts: %w", err)
	}
	workDir, err := config.GetStringWithEnvExpand(cmd, "dir-work")
	if err != nil {
		return fmt.Errorf("security alert for --dir-work: %w", err)
	}
	dirNamingType := config.GetString(cmd, "dir-naming-type")
	draftHandling := config.GetString(cmd, "draft-handling")
	includedIds := config.GetStringSlice(cmd, "ids-include")
	excludedIds := config.GetStringSlice(cmd, "ids-exclude")
	commitMsg := config.GetString(cmd, "git-commit-msg")
	commitUser := config.GetString(cmd, "git-commit-user")
	commitEmail := config.GetString(cmd, "git-commit-email")
	scriptCollectionMap := config.GetStringSlice(cmd, "script-collection-map")
	skipCommit := config.GetBool(cmd, "git-skip-commit")
	syncPackageLevelDetails := config.GetBool(cmd, "sync-package-details")
	target := config.GetString(cmd, "target")
	if target == "local" {
		target = "git"
	} else if target == "remote" {
		target = "tenant"
	}

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
			return fmt.Errorf("Package %v does not exist. Please run 'update package' command first", packageId)
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
