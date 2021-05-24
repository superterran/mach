// Package compose is a passthru for `docker compose` command that (will) support templates
package compose

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var composeCmd = CreateComposeCmd()

// ComposeDirname is the bas directory for compositions, could be set to `composes` in .mach.yaml
var ComposeDirname = "."

// TestMode var determines if certain flows actually complete or not for unit testing
var TestMode = false

// path to configruation files
var cfgFile string

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(".mach")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func CreateComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compose <docker-compose> $@",
		Short: "Runs docker compose on compositions in a directory.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompose(cmd, args)
		},
	}
	return cmd
}

func init() {

	TestMode = strings.HasSuffix(os.Args[0], ".test")

	composeCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is loaded from working dir)")

	initConfig()

	viper.SetDefault("ComposeDirname", ComposeDirname)
	ComposeDirname = viper.GetString("ComposeDirname")

}

func runCompose(cmd *cobra.Command, args []string) error {

	return MainComposeFlow(args)
}

func MainComposeFlow(args []string) error {

	if len(args) > 0 {

		s := []string{"up", "down", "ps"}

		if contains(s, args[1]) {
			composeArgs := args[1:]
			RunCompose(args[0], composeArgs)
		} else if contains(s, args[0]) {
			fmt.Println("this far")
			matches, _ := filepath.Glob(ComposeDirname + "/**/docker-compose.yml")
			for _, match := range matches {
				dir := filepath.Dir(match)
				composition := filepath.Base(dir)
				composeArgs := args[0:]
				RunCompose(composition, composeArgs)

			}

		}

	}

	return nil
}

// RunCompose is a wrapper for `docker-compose`. It requires `docker-compose` installed locally, and the command is
// invoked from the directory of the composition. When running commands, pass flags to docker-composer with --, i.e.
// `mach compose satis up -- -d --force-recreate`.
func RunCompose(composition string, args []string) {

	fmt.Println(composition)
	fmt.Println(args)

	baseCmd := "docker-compose"

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	var composeDir string = path + "/../../" + ComposeDirname + "/" + composition

	fmt.Println(composeDir)

	s := []string{"up", "down", "ps"}
	if contains(s, args[0]) {

		cmd := exec.Command(baseCmd, args...)
		cmd.Dir = composeDir
		out, err := cmd.CombinedOutput()

		fmt.Println(string(out))

		if err != nil {
			log.Fatal(err)
		}
	}

}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
