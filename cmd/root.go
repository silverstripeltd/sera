package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"fmt"
	"log/syslog"
	"log"
	"os"
	"io"
)

var (
	Verbose bool
	LogToSyslog bool
	Server string
	InfoLogger *log.Logger
	ErrorLog *log.Logger
)

func init() {

	viper.SetConfigName("sera")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.sera")
	viper.AddConfigPath(".")

	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	RootCmd.PersistentFlags().BoolVarP(&LogToSyslog, "log-to-syslog", "s", false, "log to syslog")
	viper.BindPFlag("syslog", RootCmd.PersistentFlags().Lookup("log-to-syslog"))
	RootCmd.PersistentFlags().StringVarP(&Server, "server", "m", "user:password@tcp(127.0.0.1:3306)/?timeout=500ms", "which mysql server to connect to")
	viper.BindPFlag("server", RootCmd.PersistentFlags().Lookup("server"))

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	setupLogging()
}

var RootCmd = &cobra.Command{
	Use:   "sera",
	Short: "Sera is a cluster syncronisation tool",
	Long: `Sera is a cluster syncronisation / scheduling tool.
Documentation is available at https://github.com/silverstripe-labs/sera`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func setupLogging() {
	if viper.Get("log-to-syslog").(bool) && viper.Get("verbose").(bool) {
		errSyslog, err := syslog.New(syslog.LOG_ERR, "sera")
		if err != nil {
			panic(fmt.Errorf("Fatal error setting up syslog.LOG_ERR: %s \n", err))
		}
		ErrorLog = log.New(io.MultiWriter(os.Stderr, errSyslog), "Error: ", log.Ldate|log.Ltime|log.Lshortfile)

		noticeSyslog, err := syslog.New(syslog.LOG_NOTICE, "sera")
		if err != nil {
			panic(fmt.Errorf("Fatal error setting up syslog.LOG_NOTICE: %s \n", err))
		}

		InfoLogger = log.New(noticeSyslog, "Info: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	//


}
