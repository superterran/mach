package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// TestMode var determines if certain flows actually complete or not for unit testing
var TestMode = false

var tmpDir = ""

// MachineS3Bucket defines which bucket mach interacts with for storing config tarballs, pulled from `machine-s3-bucket` in .mach.conf.yaml
var MachineS3Bucket string = "mach-docker-machine-certificates"

// MachineS3Region defines which region the bucket is in, pulled from `machine-s3-region` in .mach.conf.yaml
var MachineS3Region string = "us-east-1"

// KeepTarball will trigger a clean-up of the tarball, set to true to prevent, or `-k` or `--keep-tarball`
var KeepTarball bool = false

// OutputOnly will break execution of the build tool and will post the generated dockerfile template to stdout
// invoke with `-o` or `--outout-only`
var OutputOnly = false

var rootCmd = CreateRootCmd()

func CreateRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mach",
		Short: "Tool for mocking out environments with docker",
		Long: `A tool for provisioning and running docker compositions both locally and in the cloud.
		
		usage: mach build php:8.1`,
		RunE: func(md *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	TestMode = strings.HasSuffix(os.Args[0], ".test")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is loaded from working dir)")

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".mach")
	}

	viper.AddConfigPath(".")
	viper.SetEnvPrefix("MACH")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func createTempDirectory() string {
	dir, _ := ioutil.TempDir("/tmp", "machine")
	tmpDir = dir
	return tmpDir
}
