package cmd

import (
	"fmt"
	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/repo"
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

func NewAPIMCommand() *cobra.Command {

	apimCmd := &cobra.Command{
		Use:   "apim",
		Short: "Sync API Management artifacts between tenant and Git",
		Long: `Synchronise API Management artifacts between SAP Integration Suite
tenant and a Git repository.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// If artifacts directory is provided, validate that is it a subdirectory of Git repo
			gitRepoDir := config.GetString(cmd, "dir-git-repo")
			if gitRepoDir != "" {
				artifactsDir := config.GetString(cmd, "dir-artifacts")
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
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runSyncAPIM(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}
	// Define cobra flags, the default value has the lowest (least significant) precedence

	return apimCmd
}

func runSyncAPIM(cmd *cobra.Command) error {
	log.Info().Msg("Executing sync apim command")

	gitRepoDir := config.GetString(cmd, "dir-git-repo")
	artifactsDir := config.GetStringWithDefault(cmd, "dir-artifacts", gitRepoDir)
	workDir := config.GetString(cmd, "dir-work")
	includedIds := config.GetStringSlice(cmd, "ids-include")
	excludedIds := config.GetStringSlice(cmd, "ids-exclude")
	commitMsg := config.GetString(cmd, "git-commit-msg")
	commitUser := config.GetString(cmd, "git-commit-user")
	commitEmail := config.GetString(cmd, "git-commit-email")
	skipCommit := config.GetBool(cmd, "git-skip-commit")
	target := config.GetString(cmd, "target")
	if target == "local" {
		target = "git"
	} else if target == "remote" {
		target = "tenant"
	}
	serviceDetails := api.GetServiceDetails(cmd)
	// Initialise HTTP executer
	exe := api.InitHTTPExecuter(serviceDetails)

	syncer := sync.NewSyncer(target, "APIM", exe)
	apimWorkDir := fmt.Sprintf("%v/apim", workDir)
	err := syncer.Exec(apimWorkDir, artifactsDir, str.TrimSlice(includedIds), str.TrimSlice(excludedIds))
	if err != nil {
		return err
	}
	if target == "git" && !skipCommit {
		err = repo.CommitToRepo(gitRepoDir, commitMsg, commitUser, commitEmail)
		if err != nil {
			return err
		}
	}
	// Clean up working directory
	err = os.RemoveAll(apimWorkDir)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}
