package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Timeout int

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVarP(&Timeout, "timeout", "t", 0, "How long to wait for a lock to be unlocked")
	viper.BindPFlag("timeout", runCmd.Flags().Lookup("timeout"))

}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a command and lock it",
	Long:  `Run a command and try to be the only one that runs this command`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Do Stuff Here
		fmt.Printf("%V\n", args)
		return nil
	},
}
