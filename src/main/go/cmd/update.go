package cmd

import (
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Upload/update artifacts or integration package",
		Long: `Upload or update artifacts and/or integration package on the
SAP Integration Suite tenant.`,
	}
	return updateCmd
}
