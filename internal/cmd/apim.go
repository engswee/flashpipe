package cmd

import (
	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/repo"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"
)

func NewAPIMCommand() *cobra.Command {

	apimCmd := &cobra.Command{
		Use:   "apim",
		Short: "Sync API Management artifacts between tenant and Git",
		Long: `Synchronise API Management artifacts between SAP Integration Suite
tenant and a Git repository.`,
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
	//apimCmd.Flags().String("target", "local", "Target of sync. Allowed values: local, remote")
	//apimCmd.Flags().String("git-commit-msg", "Sync repo from tenant", "Message used in commit")
	//apimCmd.Flags().String("git-commit-user", "github-actions[bot]", "User used in commit")
	//apimCmd.Flags().String("git-commit-email", "41898282+github-actions[bot]@users.noreply.github.com", "Email used in commit")
	//apimCmd.Flags().Bool("git-skip-commit", false, "Skip committing changes to Git repository")

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

	serviceDetails := api.GetServiceDetails(cmd)
	// Initialise HTTP executer
	exe := api.InitHTTPExecuter(serviceDetails)

	syncer := sync.NewSyncer(target, "APIM", exe)

	err := syncer.Exec(workDir, artifactsDir, includedIds, excludedIds)
	if err != nil {
		return err
	}
	if target == "local" && !skipCommit {
		err = repo.CommitToRepo(gitRepoDir, commitMsg, commitUser, commitEmail)
		if err != nil {
			return err
		}
	}

	return nil
}
