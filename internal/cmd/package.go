package cmd

import (
	"time"

	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/sync"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewPackageCommand() *cobra.Command {

	packageCmd := &cobra.Command{
		Use:     "package",
		Aliases: []string{"pkg"},
		Short:   "Create/update integration package",
		Long: `Create or update integration package on the
SAP Integration Suite tenant.

Configuration:
  Settings can be loaded from the global config file (--config) under the
  'update.package' section. CLI flags override config file settings.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			startTime := time.Now()
			if err = runUpdatePackage(cmd); err != nil {
				cmd.SilenceUsage = true
			}
			analytics.Log(cmd, err, startTime)
			return
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	// Note: These can be set in config file under 'update.package' key
	packageCmd.Flags().String("package-file", "", "Path to location of package file (config: update.package.packageFile)")

	_ = packageCmd.MarkFlagRequired("package-file")
	return packageCmd
}

func runUpdatePackage(cmd *cobra.Command) error {
	log.Info().Msg("Executing update package command")

	// Support reading from config file under 'update.package' key
	packageFile := config.GetStringWithFallback(cmd, "package-file", "update.package.packageFile")

	// Initialise HTTP executer
	serviceDetails := api.GetServiceDetails(cmd)
	exe := api.InitHTTPExecuter(serviceDetails)
	packageSynchroniser := sync.NewSyncer("tenant", "CPIPackage", exe)

	return packageSynchroniser.Exec(sync.Request{PackageFile: packageFile})
}
