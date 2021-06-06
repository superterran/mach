// Cmd compose is a passthru for `docker compose` command that supports templates
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var composeCmd = CreateComposeCmd()

// ComposeDirname is the bas directory for compositions, could be set to `composes` in .mach.yaml
var ComposeDirname = "."

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

	rootCmd.AddCommand(composeCmd)

	viper.SetDefault("ComposeDirname", ComposeDirname)

	composeCmd.Flags().BoolP("output-only", "o", OutputOnly, "send output to stdout, do not build")

	composeCmd.Flags().BoolP("first-only", "f", FirstOnly, "stop the build loop after the first image is found")

}

func runCompose(cmd *cobra.Command, args []string) error {

	ComposeDirname = viper.GetString("ComposeDirname")

	OutputOnly, _ = cmd.Flags().GetBool("output-only")

	FirstOnly, _ = cmd.Flags().GetBool("first-only")

	return MainComposeFlow(args)
}

// MainComposeFlow builds and runs compositions against an array of arguments. If none are passed,
// it will iterate  through all compositons in the compose dir
func MainComposeFlow(args []string) error {

	s := []string{"up", "down", "ps"}

	if len(args) > 1 {

		if contains(s, args[1]) {
			composeArgs := args[1:]
			RunCompose(args[0], composeArgs)
		}
	}

	if len(args) > 0 {
		if contains(s, args[0]) {
			matches, _ := filepath.Glob(ComposeDirname + "/**/docker-compose.yml*")
			for _, match := range matches {
				dir := filepath.Dir(match)
				composition := filepath.Base(dir)
				composeArgs := args[0:]
				RunCompose(composition, composeArgs)

				if FirstOnly {
					break
				}

			}

		}

	}

	return nil
}

// RunCompose is a wrapper for `docker-compose`. It requires `docker-compose` installed locally, and the command is
// invoked from the directory of the composition. When running commands, pass flags to docker-composer with --, i.e.
// `mach compose satis up -- -d --force-recreate`.
func RunCompose(composition string, args []string) {

	baseCmd := "docker-compose"

	var composeDir string = ComposeDirname + "/" + composition
	composeDir, _ = filepath.Abs(composeDir)

	if _, err := os.Stat(composeDir + "/docker-compose.yml.tpl"); err == nil {
		generateCompositionTemplate(composeDir + "/docker-compose.yml.tpl")
	}

	s := []string{"up", "down", "ps"}
	if contains(s, args[0]) {

		args = append(args, "-d")

		if !OutputOnly {
			cmd := exec.Command(baseCmd, args...)
			cmd.Dir = composeDir
			out, _ := cmd.CombinedOutput()

			fmt.Println(string(out))
		}
	}

}

func generateCompositionTemplate(filename string) {

	generateFilename := filepath.Dir(filename) + "/docker-compose.yml"

	wr := os.Stdout

	if !OutputOnly {
		wr, _ = os.Create(generateFilename)
	}

	tpl, err := template.ParseGlob(filename)
	if err != nil {
		panic(err)
	}

	tpl.ParseGlob(filepath.Dir(filename) + "/includes/*.tpl")
	tpl.Execute(wr, filepath.Base(filename))

	if !OutputOnly {
		wr.Close()
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
