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
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewAPIProxyCommand() *cobra.Command {
	apiproxyCmd := &cobra.Command{
		Use:     "apiproxy",
		Aliases: []string{"apim"},
		Short:   "Sync API Management proxies (with dependent artifacts) between tenant and Git",
		Long: `Synchronise API Management proxies (with dependent artifacts) between SAP Integration Suite
tenant and a Git repository.

Configuration:
  Settings can be loaded from the global config file (--config) under the
  'sync.apiproxy' section. CLI flags override config file settings.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// If artifacts directory is provided, validate that is it a subdirectory of Git repo
			gitRepoDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-git-repo", "sync.apiproxy.dirGitRepo")
			if err != nil {
				return fmt.Errorf("security alert for --dir-git-repo: %w", err)
			}
			if gitRepoDir != "" {
				artifactsDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-artifacts", "sync.apiproxy.dirArtifacts")
				if err != nil {
					return fmt.Errorf("security alert for --dir-artifacts: %w", err)
				}
				gitRepoDirClean := filepath.Clean(gitRepoDir) + string(os.PathSeparator)
				if artifactsDir != "" && !strings.HasPrefix(artifactsDir, gitRepoDirClean) {
					return fmt.Errorf("--dir-artifacts [%v] should be a subdirectory of --dir-git-repo [%v]", artifactsDir, gitRepoDirClean)
				}
			}
			// Validate target
			target := config.GetStringWithFallback(cmd, "target", "sync.apiproxy.target")
			switch target {
			case "git", "tenant":
			default:
				return fmt.Errorf("invalid value for --target = %v", target)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runSyncAPIProxy(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	return apiproxyCmd
}

func runSyncAPIProxy(cmd *cobra.Command) error {
	log.Info().Msg("Executing sync apiproxy command")

	// Support reading from config file under 'sync.apiproxy' key
	gitRepoDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-git-repo", "sync.apiproxy.dirGitRepo")
	if err != nil {
		return fmt.Errorf("security alert for --dir-git-repo: %w", err)
	}
	artifactsDir := config.GetStringWithFallback(cmd, "dir-artifacts", "sync.apiproxy.dirArtifacts")
	if artifactsDir == "" {
		artifactsDir = gitRepoDir
	}
	workDir, err := config.GetStringWithEnvExpandAndFallback(cmd, "dir-work", "sync.apiproxy.dirWork")
	if err != nil {
		return fmt.Errorf("security alert for --dir-work: %w", err)
	}
	includedIds := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "ids-include", "sync.apiproxy.idsInclude"))
	excludedIds := str.TrimSlice(config.GetStringSliceWithFallback(cmd, "ids-exclude", "sync.apiproxy.idsExclude"))
	commitMsg := config.GetStringWithFallback(cmd, "git-commit-msg", "sync.apiproxy.gitCommitMsg")
	commitUser := config.GetStringWithFallback(cmd, "git-commit-user", "sync.apiproxy.gitCommitUser")
	commitEmail := config.GetStringWithFallback(cmd, "git-commit-email", "sync.apiproxy.gitCommitEmail")
	skipCommit := config.GetBoolWithFallback(cmd, "git-skip-commit", "sync.apiproxy.gitSkipCommit")
	target := config.GetStringWithFallback(cmd, "target", "sync.apiproxy.target")

	serviceDetails := api.GetServiceDetails(cmd)
	// Initialise HTTP executer
	exe := api.InitHTTPExecuter(serviceDetails)

	syncer := sync.NewSyncer(target, "APIProxy", exe)
	apiproxyWorkDir := fmt.Sprintf("%v/apiproxy", workDir)
	err = syncer.Exec(sync.Request{WorkDir: apiproxyWorkDir, ArtifactsDir: artifactsDir, IncludedIds: includedIds, ExcludedIds: excludedIds})
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
	err = os.RemoveAll(apiproxyWorkDir)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}
