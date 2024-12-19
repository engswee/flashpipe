package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/file"
	"github.com/engswee/flashpipe/internal/str"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func NewRestoreCommand() *cobra.Command {

	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore integration packages from Git to tenant",
		Long:  `Restore all editable integration packages from a Git repository to SAP Integration Suite tenant.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {

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
			if err = runRestore(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	return restoreCmd
}

func runRestore(cmd *cobra.Command) error {
	log.Info().Msg("Executing snapshot restore command")

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
	includedIds := str.TrimSlice(config.GetStringSlice(cmd, "ids-include"))
	excludedIds := str.TrimSlice(config.GetStringSlice(cmd, "ids-exclude"))

	serviceDetails := api.GetServiceDetails(cmd)
	err = restoreSnapshot(serviceDetails, artifactsBaseDir, workDir, includedIds, excludedIds)
	if err != nil {
		return err
	}

	return nil
}

func restoreSnapshot(serviceDetails *api.ServiceDetails, artifactsBaseDir string, workDir string, includedIds []string, excludedIds []string) error {
	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msg("📢 Begin restoring snapshot to the tenant")

	// Get directory list
	baseSourceDir := filepath.Clean(artifactsBaseDir)
	entries, err := os.ReadDir(baseSourceDir)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	// Initialise HTTP executer
	exe := api.InitHTTPExecuter(serviceDetails)
	packageSynchroniser := sync.NewSyncer("tenant", "CPIPackage", exe)
	artifactsSynchroniser := sync.New(exe)

	// Go through each directory and check if there is an integration package details in it, if yes, then proceed to restore integration package and artifacts
	for _, entry := range entries {
		packageId := entry.Name()
		packageFile := fmt.Sprintf("%v/%v/%v.json", baseSourceDir, packageId, packageId)
		if entry.IsDir() && file.Exists(packageFile) {
			packageDir := fmt.Sprintf("%v/%v", baseSourceDir, packageId)
			log.Info().Msg("---------------------------------------------------------------------------------")
			log.Info().Msgf("Processing directory %v", packageDir)

			// Filter in/out packages
			if str.FilterIDs(packageId, includedIds, excludedIds) {
				continue
			}

			// 1 - Sync CPI Integration Package
			err = packageSynchroniser.Exec("", packageDir, nil, nil)
			if err != nil {
				return err
			}

			// 2 - Sync CPI Artifacts
			err = artifactsSynchroniser.ArtifactsToTenant(packageId, workDir, packageDir, nil, nil)
			if err != nil {
				return err
			}
		}
	}

	log.Info().Msg("---------------------------------------------------------------------------------")
	log.Info().Msg("🏆 Completed restoring snapshot to the tenant")
	return nil
}
