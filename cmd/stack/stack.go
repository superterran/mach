// Package backup copies docker-machine certs and configurations to S3
package stack

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var stackCmd = CreateStackCmd()

var tmpDir = ""

// TestMode var determines if certain flows actually complete or not for unit testing
var TestMode = false

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

}

func runStack(cmd *cobra.Command, args []string) error {

	return MainStackFlow(args)
}

func MainStackFlow(args []string) error {

	return nil
}
