package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/superterran/mach/cmd/backup"
	"github.com/superterran/mach/cmd/build"
	"github.com/superterran/mach/cmd/compose"
	"github.com/superterran/mach/cmd/restore"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "mach",
	Short: "Tool for mocking out environments with docker",
	Long: `A tool for provisioning and running docker compositions both locally and in the cloud.
	
	usage: mach build php:8.1`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	initConfig()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is loaded from working dir)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(build.CreateBuildCmd())
	rootCmd.AddCommand(backup.CreateBackupCmd())
	rootCmd.AddCommand(restore.CreateRestoreCmd())
	rootCmd.AddCommand(compose.CreateComposeCmd())

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".mach")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}
