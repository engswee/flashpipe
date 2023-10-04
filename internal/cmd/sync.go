package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/logger"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/engswee/flashpipe/internal/repo"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
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
			gitRepoDir := config.GetString(cmd, "dir-git-repo")
			if gitRepoDir != "" {
				artifactsDir := config.GetString(cmd, "dir-artifacts")
				gitRepoDirClean := filepath.Clean(gitRepoDir) + string(os.PathSeparator)
				if artifactsDir != "" && !strings.HasPrefix(artifactsDir, gitRepoDirClean) {
					return fmt.Errorf("--dir-artifacts [%v] should be a subdirectory of --dir-git-repo [%v]", artifactsDir, gitRepoDirClean)
				}
			}
			// TODO - Validate secrets in env var, lower priority as it is no longer resolved in GitHub action workflow
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runSync(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	syncCmd.Flags().String("package-id", "", "ID of Integration Package")
	syncCmd.Flags().String("dir-git-repo", "", "Directory of Git repository")
	syncCmd.Flags().String("dir-artifacts", "", "Directory containing contents of artifacts")
	syncCmd.Flags().String("dir-work", "/tmp", "Working directory for in-transit files")
	syncCmd.Flags().String("dir-naming-type", "ID", "Name artifact directory by ID or Name. Allowed values: ID, NAME")
	syncCmd.Flags().String("draft-handling", "SKIP", "Handling when artifact is in draft version. Allowed values: SKIP, ADD, ERROR")
	syncCmd.Flags().StringSlice("ids-include", nil, "List of included artifact IDs")
	syncCmd.Flags().StringSlice("ids-exclude", nil, "List of excluded artifact IDs")
	syncCmd.Flags().String("target", "local", "Target of sync. Allowed values: local, remote")
	syncCmd.Flags().String("git-commit-msg", "Sync repo from tenant", "Message used in commit")
	syncCmd.Flags().String("git-commit-user", "github-actions[bot]", "User used in commit")
	syncCmd.Flags().String("git-commit-email", "41898282+github-actions[bot]@users.noreply.github.com", "Email used in commit")
	syncCmd.Flags().StringSlice("script-collection-map", nil, "Comma-separated source-target ID pairs for converting script collection references during sync ")
	syncCmd.Flags().Bool("git-skip-commit", false, "Skip committing changes to Git repository")
	syncCmd.Flags().Bool("sync-package-details", false, "Sync details of Integration Package")

	_ = syncCmd.MarkFlagRequired("package-id")
	_ = syncCmd.MarkFlagRequired("dir-git-repo")
	syncCmd.MarkFlagsMutuallyExclusive("ids-include", "ids-exclude")

	return syncCmd
}

func runSync(cmd *cobra.Command) {
	log.Info().Msg("Executing sync command")

	packageId := config.GetString(cmd, "package-id")
	gitRepoDir := config.GetString(cmd, "dir-git-repo")
	artifactsDir := config.GetStringWithDefault(cmd, "dir-artifacts", gitRepoDir)
	workDir := config.GetString(cmd, "dir-work")
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

	serviceDetails := odata.GetServiceDetails(cmd)
	// Initialise HTTP executer
	exe := odata.InitHTTPExecuter(serviceDetails)
	synchroniser := sync.New(exe)

	// Sync from tenant to Git
	if target == "local" {
		packageDataFromTenant, readOnly, err := synchroniser.VerifyDownloadablePackage(packageId)
		if err != nil {
			logger.ExitIfError(err)
		}
		if !readOnly {
			if syncPackageLevelDetails {
				err := synchroniser.PackageToLocal(packageDataFromTenant, packageId, workDir, artifactsDir)
				logger.ExitIfError(err)
			}

			err = synchroniser.ArtifactsToLocal(packageId, workDir, artifactsDir, includedIds, excludedIds, draftHandling, dirNamingType, scriptCollectionMap)
			logger.ExitIfError(err)

			if !skipCommit {
				err := repo.CommitToRepo(gitRepoDir, commitMsg, commitUser, commitEmail)
				logger.ExitIfError(err)
			}
		}
	}

	// Sync from Git to tenant
	if target == "remote" {
		err := synchroniser.ArtifactsToRemote(packageId, workDir, artifactsDir, includedIds, excludedIds)
		logger.ExitIfError(err)
	}
}
