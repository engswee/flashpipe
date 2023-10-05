package cmd

import (
	"github.com/engswee/flashpipe/internal/config"
	"github.com/engswee/flashpipe/internal/odata"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewPackageCommand() *cobra.Command {

	packageCmd := &cobra.Command{
		Use:     "package",
		Aliases: []string{"pkg"},
		Short:   "Create/update integration package",
		Long: `Create or update integration package on the
SAP Integration Suite tenant.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = runUpdatePackage(cmd); err != nil {
				cmd.SilenceUsage = true
			}
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
	packageDetails, err := odata.GetPackageDetails(packageFile)
	if err != nil {
		return err
	}

	// Initialise HTTP executer
	serviceDetails := odata.GetServiceDetails(cmd)
	exe := odata.InitHTTPExecuter(serviceDetails)
	ip := odata.NewIntegrationPackage(exe)

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
