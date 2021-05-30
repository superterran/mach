package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
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

	// initConfig()

	TestMode = strings.HasSuffix(os.Args[0], ".test")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is loaded from working dir)")

	rootCmd.AddCommand(CreateBuildCmd())
	rootCmd.AddCommand(CreateBackupCmd())
	rootCmd.AddCommand(CreateRestoreCmd())
	rootCmd.AddCommand(CreateComposeCmd())

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

func createTempDirectory() string {
	dir, err := ioutil.TempDir("/tmp", "machine")
	if err != nil {
		log.Fatal(err)
	}

	tmpDir = dir
	return tmpDir
}

func removeMachineArchive(machine string) {
	e := os.Remove(machine + ".tar.gz")
	if e != nil {
		log.Fatal(e)
	}
}
