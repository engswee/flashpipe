package cmd

import (
	"github.com/engswee/flashpipe/internal/analytics"
	"github.com/engswee/flashpipe/internal/api"
	"github.com/engswee/flashpipe/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"
)

func NewPackageCommand() *cobra.Command {

	packageCmd := &cobra.Command{
		Use:     "package",
		Aliases: []string{"pkg"},
		Short:   "Create/update integration package",
		Long: `Create or update integration package on the
SAP Integration Suite tenant.`,
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
	packageCmd.Flags().String("package-file", "", "Path to location of package file")

	_ = packageCmd.MarkFlagRequired("package-file")
	return packageCmd
}

func runUpdatePackage(cmd *cobra.Command) error {
	log.Info().Msg("Executing update package command")

	packageFile := config.GetString(cmd, "package-file")

	// Get package details from JSON file
	log.Info().Msgf("Getting package details from %v file", packageFile)
	packageDetails, err := api.GetPackageDetails(packageFile)
	if err != nil {
		return err
	}

	// Initialise HTTP executer
	serviceDetails := api.GetServiceDetails(cmd)
	exe := api.InitHTTPExecuter(serviceDetails)
	ip := api.NewIntegrationPackage(exe)

	packageId := packageDetails.Root.Id
	_, _, exists, err := ip.Get(packageId)
	if !exists {
		log.Info().Msgf("Package %v does not exist", packageId)
		err = ip.Create(packageDetails)
		if err != nil {
			return err
		}
		log.Info().Msgf("Package %v created", packageId)
	} else {
		// Update integration package
		err = ip.Update(packageDetails)
		if err != nil {
			return err
		}
		log.Info().Msgf("Package %v updated", packageId)
	}
	return nil
}
