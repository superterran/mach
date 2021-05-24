package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/superterran/mach/cmd/backup"
	"github.com/superterran/mach/cmd/build"
	"github.com/superterran/mach/cmd/restore"
	"github.com/superterran/mach/cmd/stack"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "mach",
	Short: "Tool for mocking out environments with docker",
	Long: `A tool for provisioning and running docker compositions both locally and in the cloud.
	
	usage: mach build php:7.3`,
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is loaded from working dir, or $HOME/.mach.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(build.CreateBuildCmd())
	rootCmd.AddCommand(backup.CreateBackupCmd())
	rootCmd.AddCommand(restore.CreateRestoreCmd())
	rootCmd.AddCommand(stack.CreateStackCmd())

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".mach" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".mach")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}
