package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of sera",
	Long:  `All software has versions. This is sera's`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Do Stuff Here
		return nil
	},
}
