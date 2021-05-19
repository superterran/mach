/*
Copyright © 2021 Doug Hatcher <superterran@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var backupCmd = createBackupCmd()

func createBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup <docker-machine>",
		Short: "Takes a working docker-machine entry and stores it to an S3 bucket",
		Long: `This allows you to store the docker-machine certs bundle in an S3 bucket, 
paired with restore command, this will let you transfer docker-machines to and from
systems using the AWS API. Will require progamtic credetials with permissions to upload
to S3.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackup(cmd, args)
		},
	}
	return cmd
}

func init() {

	rootCmd.AddCommand(backupCmd)
	testMode = strings.HasSuffix(os.Args[0], ".test")
}

func runBackup(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {

	}

	return nil
}
