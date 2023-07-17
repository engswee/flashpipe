package cmd

import (
	"encoding/json"
	"github.com/engswee/flashpipe/config"
	"github.com/engswee/flashpipe/logger"
	"github.com/engswee/flashpipe/odata"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func NewPackageCommand() *cobra.Command {

	packageCmd := &cobra.Command{
		Use:     "package",
		Aliases: []string{"pkg"},
		Short:   "Create/update integration package",
		Long: `Create or update integration package on the
SAP Integration Suite tenant.`,
		Run: func(cmd *cobra.Command, args []string) {
			runUpdatePackage(cmd)
		},
	}

	// Define cobra flags, the default value has the lowest (least significant) precedence
	packageCmd.Flags().String("package-file", "", "Path to location of package file [or set environment PACKAGE_FILE]")
	packageCmd.Flags().String("package-override-id", "", "Override package ID from file [or set environment PACKAGE_ID]")
	packageCmd.Flags().String("package-override-name", "", "Override package name from file [or set environment PACKAGE_NAME]")

	return packageCmd
}

func runUpdatePackage(cmd *cobra.Command) {
	log.Info().Msg("Executing update package command")

	packageFile := config.GetMandatoryString(cmd, "package-file")
	packageOverrideId := config.GetString(cmd, "package-override-id")
	packageOverrideName := config.GetString(cmd, "package-override-name")

	// Get package details from JSON file
	log.Info().Msgf("Getting package details from %v file", packageFile)
	packageDetails, err := getPackageDetails(packageFile)
	logger.ExitIfError(err)

	// Overwrite ID & Name
	if packageOverrideId != "" {
		packageDetails.Root.Id = packageOverrideId
	}
	if packageOverrideName != "" {
		packageDetails.Root.Name = packageOverrideName
	}

	// Initialise HTTP executer
	serviceDetails := odata.GetServiceDetails(cmd)
	exe := odata.InitHTTPExecuter(serviceDetails)
	ip := odata.NewIntegrationPackage(exe)

	packageId := packageDetails.Root.Id
	exists, err := ip.Exists(packageId)
	if !exists {
		log.Info().Msgf("Package %v does not exist. Creating package...", packageId)
		err = ip.Create(packageDetails)
		logger.ExitIfError(err)
		log.Info().Msgf("Package %v created", packageId)
	} else {
		// Update integration package
		log.Info().Msgf("Updating package %v", packageId)
		err = ip.Update(packageDetails)
		logger.ExitIfError(err)
		log.Info().Msgf("Package %v updated", packageId)
	}
}

func getPackageDetails(file string) (*odata.PackageSingleData, error) {
	var jsonData *odata.PackageSingleData

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(fileContent, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}
