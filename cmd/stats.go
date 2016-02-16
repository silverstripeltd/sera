package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics",
	Long:  `Show statistics from previous runs of sera`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Do Stuff Here
		return nil
	},
}
