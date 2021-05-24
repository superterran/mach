// Package stack is a passthru for `docker stack` command that (will) support templates
package stack

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

var stackCmd = CreateStackCmd()

var tmpDir = ""

// StacksDirname is the bas directory for compositions, could be set to `stacks` in .mach.yaml
var StacksDirname = "."

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

func CreateStackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stack <docker-stack> $@",
		Short: "Runs docker stack on compositions in a directory.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStack(cmd, args)
		},
	}
	return cmd
}

func init() {

	TestMode = strings.HasSuffix(os.Args[0], ".test")

	stackCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is loaded from working dir)")

	initConfig()

	viper.SetDefault("StacksDirname", StacksDirname)
	StacksDirname = viper.GetString("StacksDirname")

}

func runStack(cmd *cobra.Command, args []string) error {

	return MainStackFlow(args)
}

func MainStackFlow(args []string) error {
	if len(args) > 1 {
		RunStack(args)
	} else if len(args) > 0 {
		matches, _ := filepath.Glob(StacksDirname + "/*/*,yml")
		for _, match := range matches {
			file := filepath.Base(match)

			perargs := []string{file}
			perargs = append(perargs, args...)
			RunStack(perargs)
		}

	}

	return nil
}

func RunStack(args []string) {

	stack := args[1]
	baseCmd := "docker"

	var stackFile string = StacksDirname + "/" + stack + ".yml"

	cmdArgs := []string{"stack"}
	cmdArgs = append(cmdArgs, args[0:]...)
	cmdArgs = append(cmdArgs, "--compose-file")
	cmdArgs = append(cmdArgs, stackFile)

	fmt.Println(cmdArgs)

	s := []string{"deploy", "ps"}
	if contains(s, args[1]) {

		cmd := exec.Command(baseCmd, cmdArgs...)
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
