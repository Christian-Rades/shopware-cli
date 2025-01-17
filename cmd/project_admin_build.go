package cmd

import (
	"github.com/spf13/cobra"
)

var projectAdminBuildCmd = &cobra.Command{
	Use:   "admin-build",
	Short: "Builds the Administration",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		var projectRoot string
		var err error

		if projectRoot, err = findClosestShopwareProject(); err != nil {
			return err
		}

		forceNpmInstall, _ := cobraCmd.PersistentFlags().GetBool("npm-install")

		return buildAdministration(projectRoot, forceNpmInstall)
	},
}

func init() {
	projectRootCmd.AddCommand(projectAdminBuildCmd)
	projectAdminBuildCmd.PersistentFlags().Bool("npm-install", false, "Run npm install")
}
